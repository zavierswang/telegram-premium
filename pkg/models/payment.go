package models

import "time"

type Payment struct {
	ID          int64     `gorm:"column:id;primaryKey"`
	UserID      string    `gorm:"column:user_id;not null;size:64"`      //提交订单人UserID
	Username    string    `gorm:"column:username;not null;size:32"`     //提交订单人
	ForUsername string    `gorm:"column:for_username;not null;size:32"` //实际给于人
	Month       int       `gorm:"column:month;not null;size:8"`         //订单套餐
	Amount      float64   `gorm:"column:amount;default:0.0"`            //订单所需金额
	Payments    float64   `gorm:"column:payments;default:0.0"`          //实际所需金额
	Type        string    `gorm:"column:type;size:12"`                  //订单类型（manual, payment）
	Mode        string    `gorm:"column:mode;size:18"`                  //订单支付方式(cash, balance)
	MessageID   int       `gorm:"column:message_id;not null"`           //订单消息ID
	Finished    bool      `gorm:"column:finished;default:false"`        //是否完成订单
	Expired     bool      `gorm:"column:expired;default:false"`         //是否过期
	Status      int       `gorm:"column:status;default:0;size:6"`       //订单状态 1:成功; 2:正在进行; 3:已收到转帐; 4:Api提交完成；5:Api提交失败; 6:取消
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime"`     //订单创建人
}

func (p *Payment) TableName() string {
	return "tb_payment"
}
