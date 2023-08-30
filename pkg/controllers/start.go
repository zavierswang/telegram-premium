package controllers

import (
	"context"
	"github.com/mr-linch/go-tg"
	"github.com/mr-linch/go-tg/tgb"
	"go.uber.org/zap/buffer"
	"html/template"
	"strings"
	"telegram-premium/pkg/core/cst"
	"telegram-premium/pkg/core/global"
	"telegram-premium/pkg/core/logger"
	"telegram-premium/pkg/models"
	"time"
)

func Start(ctx context.Context, update *tgb.MessageUpdate) error {
	userId := update.Message.From.ID.PeerID()
	username := update.Message.From.Username.PeerID()
	username = strings.ReplaceAll(username, "@", "")
	logger.Info("[%s %s] trigger action [start] controller", userId, username)
	bot := NewBot()
	err := update.Client.SetMyCommands(bot.Cmd).DoVoid(ctx)
	if err != nil {
		logger.Error("[%s %s] set command failed %v", userId, username, err)
		return err
	}
	voucher := models.Voucher{
		ID:             time.Now().UnixMicro(),
		UserID:         userId,
		Username:       username,
		Balance:        100.000,
		Status:         1,
		ReceiveAddress: global.App.Config.Telegram.ReceiveAddress,
		FromAddress:    "",
		MessageID:      0,
	}
	global.App.DB.Save(&voucher)
	buf := new(buffer.Buffer)
	tmpl, err := template.ParseFiles(cst.StartTemplateFile)
	if err != nil {
		logger.Error("[%s %s] template parse file %s, failed %v", userId, username, cst.StartTemplateFile, err)
		return err
	}
	err = tmpl.Execute(buf, username)
	if err != nil {
		logger.Error("[%s %s] template execute file %s, failed %v", userId, username, cst.StartTemplateFile, err)
		return err
	}
	return update.Answer(buf.String()).
		ParseMode(tg.HTML).
		ReplyMarkup(bot.ReplayMarkup).
		DisableWebPagePreview(true).
		DoVoid(ctx)
}
