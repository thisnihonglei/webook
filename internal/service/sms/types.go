package sms

import "context"

type Service interface {
	send(ctx context.Context, tplId string, args []string, numbers ...string) error
}
