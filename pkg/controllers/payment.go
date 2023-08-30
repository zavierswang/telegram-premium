package controllers

import (
	"context"
	"github.com/mr-linch/go-tg/tgb"
)

type AlgoPayment interface {
	render(ctx context.Context, callback *tgb.CallbackQueryUpdate) error
}

type Payment struct {
	alg AlgoPayment
}

func (p *Payment) render(ctx context.Context, callback *tgb.CallbackQueryUpdate) error {
	return p.alg.render(ctx, callback)
}

func (p *Payment) set(algo AlgoPayment) {
	p.alg = algo
}

func NewPayment() *Payment {
	return &Payment{alg: &Cash{}}
}

func SetPayment(cast float64, ctx context.Context, callback *tgb.CallbackQueryUpdate) error {
	payment := NewPayment()
	if cast >= 0 {
		payment.set(&Balance{})
		return payment.render(ctx, callback)
	}
	return payment.render(ctx, callback)
}
