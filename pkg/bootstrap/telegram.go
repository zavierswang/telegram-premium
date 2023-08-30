package bootstrap

import (
	"context"
	"github.com/mr-linch/go-tg"
	"github.com/mr-linch/go-tg/tgb"
	"os"
	"os/signal"
	"syscall"
	"telegram-premium/pkg/controllers"
	"telegram-premium/pkg/core/cst"
	"telegram-premium/pkg/core/global"
	"telegram-premium/pkg/core/logger"
	"telegram-premium/pkg/routes"
)

func Telegram() error {
	opts := []tg.ClientOption{tg.WithClientServerURL(cst.TelegramApi)}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill, syscall.SIGTERM)
	defer cancel()
	token := global.App.Config.Telegram.Token
	if global.App.Config.App.Env == "release" {
		token = cst.TelegramToken
	}
	global.App.Client = tg.New(token, opts...)
	me, err := global.App.Client.Me(ctx)
	if err != nil {
		logger.Error("authorized failed %v", err)
		return err
	}
	logger.Info("authorized %s successfully.", me.Username.Link())
	//telegram认证成功，启动cron任务
	StartCron()
	controllers.Update(token)
	r := tgb.NewRouter()
	routes.Telegram(r)
	return tgb.NewPoller(r, global.App.Client).Run(ctx)
}
