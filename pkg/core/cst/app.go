package cst

const (
	AppName           = "telegram-premium"
	BaseName          = "telegram"
	DateTimeFormatter = "2006-01-02 15:04:05"
	TimeFormatter     = "02 15:04:05"
	UserAgent         = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/113.0.0.0 Safari/537.36"
)

const (
	OrderStatus = iota
	OrderStatusSuccess
	OrderStatusRunning
	OrderStatusReceived
	OrderStatusApiSuccess
	OrderStatusApiFailure
	OrderStatusFailure
	OrderStatusNotSufficientFunds
	OrderStatusCancel
)

const (
	OkxMarketTradesApi  = "https://www.okx.com/priapi/v5/market/trades"
	OkxTradingOrdersApi = "https://www.okx.com/v3/c2c/tradingOrders/books"

	ET51Api = "https://open.et15.com"
	ET51Vip = "/api/Daikai/submit" //代开Telegram会员
)
