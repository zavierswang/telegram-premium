package bootstrap

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/mr-linch/go-tg"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap/buffer"
	"html/template"
	"math"
	"os"
	"strconv"
	"telegram-premium/pkg/controllers"
	"telegram-premium/pkg/core/cst"
	"telegram-premium/pkg/core/global"
	"telegram-premium/pkg/core/logger"
	"telegram-premium/pkg/models"
	"telegram-premium/pkg/services"
	"telegram-premium/pkg/services/tron"
	"telegram-premium/pkg/utils"
	"time"
)

var (
	params  map[string]string
	headers map[string]string
)

func StartCron() {
	global.App.Cron = cron.New(cron.WithSeconds())
	go func() {
		listenUSDT := &ListenUSDT{
			ctx: context.Background(),
		}
		_, err := global.App.Cron.AddJob("*/30 * * * * *", listenUSDT)
		if err != nil {
			logger.Error("[scheduler] add job listenUSDT failed %v", err)
			return
		}
		global.App.Cron.Start()
		defer global.App.Cron.Stop()
		select {}
	}()
}

type ListenUSDT struct {
	ctx        context.Context
	ticker     int
	trc20Queue []string
	algo       Algo
}

func (l *ListenUSDT) Run() {
	l.ticker++
	now := time.Now()
	params = map[string]string{
		"limit":            "30",
		"start":            "0",
		"relatedAddress":   global.App.Config.Telegram.ReceiveAddress,
		"contract_address": cst.ContractAddress,
		"sort":             "-timestamp",
		"count":            "true",
		"filterTokenValue": "0",
		"start_timestamp":  strconv.FormatInt(now.Add(-120*time.Second).UnixMilli(), 10),
		"end_timestamp":    strconv.FormatInt(now.UnixMilli(), 10),
	}
	headers = map[string]string{
		"TRON-PRO-API-KEY": global.App.Config.Telegram.TronScanApiKey,
	}
	transfers, err := tron.TRC20Transfer(params, headers, true, true)
	if err != nil {
		logger.Error("[scheduler] tron.TRC20Transfer failed %v", err)
		return
	}
	// 首次启动并获取到的数据暂存到队列中
	if l.ticker == 1 {
		for _, transfer := range transfers {
			l.trc20Queue = append(l.trc20Queue, transfer.TransactionId)
		}
		logger.Info("scheduler ListenUSDT %d times, trc20 %+v", l.ticker, l.trc20Queue)
		return
	}
	slice1 := l.trc20Queue
	var slice2 []string
	for _, transfer := range transfers {
		slice2 = append(slice2, transfer.TransactionId)
	}
	// 比对历史数据获取最新的交易号
	txIds, _ := utils.Comp(slice1, slice2)
	for _, txId := range txIds {
		for _, transfer := range transfers {
			quant, _ := strconv.ParseFloat(transfer.Quant, 64)
			//忽略小于0.01USDT的金额
			if quant < math.Pow10(4) {
				logger.Warn("to tiny quant, will be ignore")
				return
			}

			if transfer.TransactionId == txId {
				l.exec(quant, transfer)
			}
		}
	}
	// 历史数据过多，删除部分数据
	if len(l.trc20Queue) >= 500 {
		logger.Info("[scheduler] clean remain queue ...")
		l.trc20Queue = l.trc20Queue[400:]
	}
	// 合并历史交易号，获新的数据
	l.trc20Queue = utils.Union(l.trc20Queue, slice2)
	return
}

func (l *ListenUSDT) exec(quant float64, transfer tron.Transfer) {
	var payments []models.Payment
	global.App.DB.Find(&payments, "amount = ? AND finished = ? AND expired = ? AND status in ?", quant, false, false, []int{2, 3, 4})
	if len(params) != 0 {
		l.algo = &payment{}
	}
	var vouchers []models.Voucher
	global.App.DB.Find(&vouchers, "balance = ? AND status in ?", quant, []int{2, 3, 4})
	if len(vouchers) != 0 {
		l.algo = &voucher{}
	}
	if l.algo == nil {
		logger.Warn("[scheduler] not found record in database [tb_payment, tb_voucher]")
		return
	}

	l.algo.exec(quant, transfer)
}

type Algo interface {
	exec(amount float64, transfer tron.Transfer)
}

// 现金支付
type payment struct{}

func (p *payment) exec(amount float64, transfer tron.Transfer) {
	logger.Info("[scheduler] strategy payment exec")
	// 更新订单记录状态，已经收到金额
	var order models.Payment
	global.App.DB.Model(&order).
		Where("amount = ? AND finished = ? AND expired = ?", amount, false, false).
		Updates(map[string]interface{}{"finished": false, "status": cst.OrderStatusReceived})
	switch global.App.Config.Telegram.Channel {
	case "manual":
		logger.Info("[scheduler] manual premium order")
		tpl := controllers.GiftTmpl{
			ForUsername: order.ForUsername,
			Month:       order.Month,
			Amount:      order.Amount,
			Mode:        order.Mode,
			Current:     order.CreatedAt.Format(cst.DateTimeFormatter),
			Type:        order.Type,
		}
		pf, _ := os.ReadFile(cst.PaymentNoticeTemplateFile)
		tmpl, _ := template.New("notice").Funcs(template.FuncMap{"format": utils.FormatTime}).Parse(string(pf))
		buf := new(buffer.Buffer)
		_ = tmpl.Execute(buf, tpl)
		layout := tg.NewButtonLayout[tg.InlineKeyboardButton](1).Row(
			tg.NewInlineKeyboardButtonURL("联系客服", fmt.Sprintf("https://t.me/%s", global.App.Config.App.Support)),
		)
		inlineKeyboard := tg.NewInlineKeyboardMarkup(layout.Keyboard()...)
		userId, _ := strconv.ParseInt(order.UserID, 10, 64)
		chatId := tg.ChatID(userId)
		_ = global.App.Client.SendMessage(chatId, buf.String()).ParseMode(tg.HTML).ReplyMarkup(inlineKeyboard).DoVoid(context.Background())

		//群组通知，管理处理
		groupId := tg.Username(global.App.Config.App.Group)
		tmpl, _ = template.ParseFiles(cst.PaymentGroupTemplateFile)
		tpl = controllers.GiftTmpl{
			ForUsername: order.ForUsername,
			Month:       order.Month,
			Amount:      order.Amount,
			Mode:        order.Mode,
			Current:     order.CreatedAt.Format(cst.DateTimeFormatter),
			Type:        order.Type,
			Status:      "收到订单款",
		}
		buf = new(buffer.Buffer)
		err := tmpl.Execute(buf, tpl)
		if err != nil {
			logger.Error("[scheduler] template execute file %s, failed %v", cst.PaymentGroupTemplateFile, err)
			return
		}
		layout = tg.NewButtonLayout[tg.InlineKeyboardButton](1).Row()
		layout.Insert(
			tg.NewInlineKeyboardButtonCallback("已处理并通知用户", fmt.Sprintf("manual %d", order.ID)),
		)
		inlineKeyboard = tg.NewInlineKeyboardMarkup(layout.Keyboard()...)
		_ = global.App.Client.SendMessage(groupId, buf.String()).ParseMode(tg.HTML).ReplyMarkup(inlineKeyboard).DoVoid(context.Background())
	case "tmt":
		logger.Info("[scheduler] tmt premium order")
		tpl := controllers.GiftTmpl{
			ForUsername: order.ForUsername,
			Month:       order.Month,
			Amount:      order.Amount,
			Mode:        order.Mode,
			Current:     order.CreatedAt.Format(cst.DateTimeFormatter),
			Type:        order.Type,
		}
		tmpl, _ := template.ParseFiles(cst.PaymentNoticeTemplateFile)
		buf := new(buffer.Buffer)
		err := tmpl.Execute(buf, tpl)
		if err != nil {
			logger.Error("[scheduler] template execute file %s, failed %v", cst.PaymentNoticeTemplateFile, err)
			return
		}
		layout := tg.NewButtonLayout[tg.InlineKeyboardButton](1).Row(
			tg.NewInlineKeyboardButtonURL("联系客服", fmt.Sprintf("https://t.me/%s", global.App.Config.App.Support)),
		)
		inlineKeyboard := tg.NewInlineKeyboardMarkup(layout.Keyboard()...)
		userId, _ := strconv.ParseInt(order.UserID, 10, 64)
		chatId := tg.ChatID(userId)
		messageId := order.MessageID
		err = services.TelegramPremiumSubmit(order.ForUsername, order.Payments)
		if err != nil {
			logger.Error("[scheduler] telegram premium submit tmt api failed %v", err)
			tpl.Status = "失败"
			global.App.DB.Model(models.Payment{}).
				Where("amount = ? AND finished = ? AND expired = ?", amount, false, false).
				Updates(map[string]interface{}{"finished": true, "status": cst.OrderStatusApiFailure})
		} else {
			logger.Info("[scheduler] telegram premium submit tmt api successfully")
			tpl.Status = "成功"
			global.App.DB.Model(models.Payment{}).
				Where("amount = ? AND finished = ? AND expired = ?", amount, false, false).
				Updates(map[string]interface{}{"finished": true, "status": cst.OrderStatusSuccess})
		}
		logger.Info("[scheduler] notice user and groups")
		_ = global.App.Client.DeleteMessage(chatId, messageId).DoVoid(context.Background())
		_ = global.App.Client.SendMessage(chatId, buf.String()).ParseMode(tg.HTML).ReplyMarkup(inlineKeyboard).DoVoid(context.Background())
		gid := tg.Username(global.App.Config.App.Group)
		_ = global.App.Client.SendMessage(gid, buf.String()).ParseMode(tg.HTML).ReplyMarkup(inlineKeyboard).DoVoid(context.Background())
		return
	}
}

// 余额充值
type voucher struct{}

func (v *voucher) exec(amount float64, transfer tron.Transfer) {
	logger.Info("[scheduler] strategy voucher exec")
	var order models.Voucher
	err := global.App.DB.First(&order, "balance = ? AND status = ?", amount, cst.OrderStatusRunning).Error
	if err != nil {
		return
	}
	logger.Info("[scheduler isRecharge] telegram premium recharge balance order: %d", order.ID)
	order.Status = cst.OrderStatusSuccess
	global.App.DB.Save(&order)
	var totalBalance sql.NullFloat64
	var totalPayment sql.NullFloat64
	// balance记录总额(balance 帐户余额购买、cash 现金购买)
	global.App.DB.Raw("SELECT SUM(amount) AS amount FROM tb_payment WHERE user_id = ? AND finished = ? AND type = ? ", order.UserID, true, "balance").Scan(&totalPayment)
	// 充值记录总额
	global.App.DB.Raw("SELECT SUM(balance) AS balance FROM tb_voucher WHERE user_id = ? AND status = ?", order.UserID, cst.OrderStatusSuccess).Scan(&totalBalance)
	remainBalance := totalBalance.Float64 - totalPayment.Float64
	// 消息通知
	tpl := VoucherTmpl{
		Username:    order.Username,
		Amount:      fmt.Sprintf("%.3f", order.Balance),
		TotalAmount: fmt.Sprintf("%.3f", remainBalance),
		CreatedAt:   order.CreatedAt.Format(cst.DateTimeFormatter),
	}
	tmpl, _ := template.ParseFiles(cst.VoucherNoticeTemplateFile)
	buf := new(buffer.Buffer)
	err = tmpl.Execute(buf, tpl)
	if err != nil {
		logger.Info("[scheduler] template execute file %s, failed %v", cst.VoucherNoticeTemplateFile, err)
		return
	}
	userId, _ := strconv.ParseInt(order.UserID, 10, 64)
	chatId := tg.ChatID(userId)
	_ = global.App.Client.DeleteMessage(chatId, order.MessageID).DoVoid(context.Background())
	_ = global.App.Client.SendMessage(chatId, buf.String()).ParseMode(tg.HTML).DoVoid(context.Background())

	//群组消息通知
	gid := tg.Username(global.App.Config.App.Group)
	_ = global.App.Client.SendMessage(gid, buf.String()).ParseMode(tg.HTML).DoVoid(context.Background())
}

type VoucherTmpl struct {
	Username    string
	Amount      string
	TotalAmount string
	CreatedAt   string
}
