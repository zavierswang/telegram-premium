package controllers

import (
	"context"
	"fmt"
	"github.com/alexeyco/simpletable"
	"github.com/mr-linch/go-tg"
	"github.com/mr-linch/go-tg/tgb"
	"go.uber.org/zap/buffer"
	"strings"
	"telegram-premium/pkg/core/cst"
	"telegram-premium/pkg/core/global"
	"telegram-premium/pkg/core/logger"
	"telegram-premium/pkg/models"
	"time"
)

func Orders(ctx context.Context, update *tgb.MessageUpdate) error {
	userId := update.Message.From.ID.PeerID()
	username := update.Message.From.Username.PeerID()
	username = strings.ReplaceAll(username, "@", "")
	logger.Info("[%s %s] trigger action [orders] controller", userId, username)

	var orders []models.Payment
	global.App.DB.Find(&orders, "user_id = ? AND username = ? AND finished = ? AND expired = ? AND status = ?", userId, username, true, false, cst.OrderStatusSuccess).
		Order("-created_at").
		Limit(20)
	table := simpletable.New()
	table.Header = &simpletable.Header{
		Cells: []*simpletable.Cell{
			{Align: simpletable.AlignLeft, Text: "赠与用户"},
			{Align: simpletable.AlignCenter, Text: "订单金额"},
			{Align: simpletable.AlignCenter, Text: "订阅套餐"},
			{Align: simpletable.AlignCenter, Text: "订阅时间"},
		},
	}
	var totalAmount float64
	for _, order := range orders {
		row := []*simpletable.Cell{
			{Align: simpletable.AlignLeft, Text: order.ForUsername},
			{Align: simpletable.AlignLeft, Text: fmt.Sprintf("%.3f USDT", order.Amount)},
			{Align: simpletable.AlignLeft, Text: fmt.Sprintf("%d个月", order.Month)},
			{Align: simpletable.AlignLeft, Text: order.CreatedAt.Format(time.DateOnly)},
		}
		totalAmount += order.Amount
		table.Body.Cells = append(table.Body.Cells, row)
	}
	table.Footer = &simpletable.Footer{
		Cells: []*simpletable.Cell{
			{},
			{},
			{Align: simpletable.AlignRight, Text: "总计金额"},
			{Align: simpletable.AlignCenter, Text: fmt.Sprintf("%.3f USDT", totalAmount)},
		},
	}
	table.SetStyle(simpletable.StyleCompactLite)
	buf := new(buffer.Buffer)
	buf.AppendString("您的Telegram Premium购买记录：\n\n")
	buf.AppendString(fmt.Sprintf("<pre>%s</pre>", table.String()))

	return update.Answer(buf.String()).ParseMode(tg.HTML).DoVoid(ctx)
}
