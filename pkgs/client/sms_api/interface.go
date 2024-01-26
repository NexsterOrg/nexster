package smsapi

import "context"

type Interface interface {
	SendSms(ctx context.Context, from, msg, to string) (err error)
}
