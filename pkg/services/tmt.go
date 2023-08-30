package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go.uber.org/zap/buffer"
	"io"
	"net/http"
	"strconv"
	"telegram-premium/pkg/core/cst"
	"telegram-premium/pkg/core/global"
	"telegram-premium/pkg/core/logger"
	"telegram-premium/pkg/utils"
	"time"
)

func TelegramPremiumSubmit(username string, payments float64) error {
	uri := fmt.Sprintf("%s%s", cst.ET51Api, cst.ET51Vip)
	var taocan string
	switch payments {
	case global.App.Config.Telegram.ThreeMonth:
		taocan = "0"
	case global.App.Config.Telegram.SixMonth:
		taocan = "1"
	case global.App.Config.Telegram.TwelveMonth:
		taocan = "2"
	default:
		taocan = "0"
	}
	orderId := strconv.FormatInt(time.Now().UnixMilli(), 10)
	sign := fmt.Sprintf("%s@%s%s%s%s", global.App.Config.Telegram.ChannelId, username, orderId, taocan, global.App.Config.Telegram.ChannelKey)
	logger.Info("premium submit sign: %s", sign)
	sign = utils.MD5([]byte(sign))
	//DONE：需要增加回调地址
	body := map[string]string{
		"uid":        global.App.Config.Telegram.ChannelId,
		"username":   fmt.Sprintf("@%s", username),
		"orderid":    orderId,
		"taocan":     taocan,
		"sign":       sign,
		"notify_url": "",
	}
	logger.Info("telegram premium submit request body: %+v", body)
	buf, _ := json.Marshal(body)
	resp, err := http.Post(uri, "application/json", bytes.NewReader(buf))
	if err != nil {
		logger.Error("submit telegram premium suite failed %v", err)
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	bf := new(buffer.Buffer)
	_, _ = io.Copy(bf, resp.Body)
	logger.Info("telegram premium submit response: %s", bf.String())
	return nil
}
