package controllers

import (
	"context"
	"fmt"
	"github.com/mr-linch/go-tg"
	"github.com/mr-linch/go-tg/tgb"
	"telegram-premium/pkg/core/global"
	"telegram-premium/pkg/middleware"
)

func Voucher(ctx context.Context, update *tgb.MessageUpdate) error {
	layoutBalance := tg.NewButtonLayout[tg.InlineKeyboardButton](3).Row()
	layoutBalance.Insert(
		tg.NewInlineKeyboardButtonCallback("30 USDT", "voucher 30"),
		tg.NewInlineKeyboardButtonCallback("60 USDT", "voucher 60"),
		tg.NewInlineKeyboardButtonCallback("90 USDT", "voucher 90"),
		tg.NewInlineKeyboardButtonCallback("120 USDT", "voucher 120"),
		tg.NewInlineKeyboardButtonCallback("150 USDT", "voucher 150"),
		tg.NewInlineKeyboardButtonCallback("200 USDT", "voucher 200"),
	)
	inlineKeyboard := tg.NewInlineKeyboardMarkup(layoutBalance.Keyboard()...)
	layoutBase := tg.NewButtonLayout[tg.InlineKeyboardButton](2).Row()
	layoutBase.Insert(
		tg.NewInlineKeyboardButtonCallback("关闭", "close"),
		tg.NewInlineKeyboardButtonURL("联系客服", fmt.Sprintf("https://t.me/%s", global.App.Config.App.Support)),
	)
	inlineKeyboardBase := tg.NewInlineKeyboardMarkup(layoutBase.Keyboard()...)
	inlineKeyboard.InlineKeyboard = append(inlineKeyboard.InlineKeyboard, inlineKeyboardBase.InlineKeyboard...)
	sess := middleware.SessionManager.Get(ctx)
	sess.Step = middleware.SessionPremiumVoucher
	return update.Answer("请选择充值金额：").ParseMode(tg.HTML).ReplyMarkup(inlineKeyboard).DoVoid(ctx)
}
