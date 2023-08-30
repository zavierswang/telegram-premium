package routes

import (
	"github.com/mr-linch/go-tg"
	"github.com/mr-linch/go-tg/tgb"
	"regexp"
	"telegram-premium/pkg/controllers"
	"telegram-premium/pkg/middleware"
)

func Telegram(router *tgb.Router) {
	router.Use(middleware.SessionManager)
	router.Use(tgb.MiddlewareFunc(middleware.Hook))

	router.Message(controllers.Start, tgb.Command("start"), tgb.ChatType(tg.ChatTypePrivate))
	router.Message(controllers.Premium, tgb.TextEqual(controllers.Menu.Premium), tgb.ChatType(tg.ChatTypePrivate))
	router.Message(controllers.Gift, tgb.Any(middleware.IsSessionStep(middleware.SessionPremiumChoose), tgb.Regexp(regexp.MustCompile(`^@?\w+`))), tgb.ChatType(tg.ChatTypePrivate))
	router.Message(controllers.Voucher, tgb.TextEqual(controllers.Menu.Balance), tgb.ChatType(tg.ChatTypePrivate))
	router.Message(controllers.Personal, tgb.TextEqual(controllers.Menu.Personal), tgb.ChatType(tg.ChatTypePrivate))
	router.Message(controllers.Orders, tgb.TextEqual(controllers.Menu.History), tgb.ChatType(tg.ChatTypePrivate))

	router.CallbackQuery(controllers.Choose, tgb.Regexp(regexp.MustCompile(`^\d+`)))
	router.CallbackQuery(controllers.Confirm, tgb.TextEqual("confirm"), tgb.ChatType(tg.ChatTypePrivate))
	router.CallbackQuery(controllers.Close, tgb.TextEqual("close"), tgb.ChatType(tg.ChatTypePrivate))
	router.CallbackQuery(controllers.Cancel, tgb.Regexp(regexp.MustCompile(`^cancel`)), tgb.ChatType(tg.ChatTypePrivate))
	router.CallbackQuery(controllers.Notifier, tgb.Regexp(regexp.MustCompile(`^manual\s+\d+`)))
	router.CallbackQuery(controllers.VoucherChoose, tgb.Any(tgb.Regexp(regexp.MustCompile(`^voucher\s+\d+`)), middleware.IsSessionStep(middleware.SessionPremiumVoucher)), tgb.ChatType(tg.ChatTypePrivate))
}
