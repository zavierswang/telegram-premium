package models

import "time"

type Voucher struct {
	ID             int64     `gorm:"column:id;primaryKey;"`
	UserID         string    `gorm:"column:user_id;not null;size:64;用户ID"` //提交订单人UserID
	Username       string    `gorm:"column:username;not null;size:32;用户名"` //提交订单人
	Balance        float64   `gorm:"column:balance;default:0.0;comment:充值金额"`
	Status         int       `gorm:"column:status;default:0;comment:充值状态"`
	ReceiveAddress string    `gorm:"column:receive_address;size:64;金额到帐地址"`
	FromAddress    string    `gorm:"column:from_address;size:64;金额来源地址"`
	MessageID      int       `gorm:"column:message_id;default:0;comment:消息ID"`
	CreatedAt      time.Time `gorm:"column:created_at;autoCreateTime;comment:创建时间"`
}

func (p *Voucher) TableName() string {
	return "tb_voucher"
}
