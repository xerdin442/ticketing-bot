package middleware

import (
	"github.com/xerdin442/ticketing-bot/internal/api/handlers"
)

type Middleware struct {
	handlers.Base
}

func New(b handlers.Base) *Middleware {
	return &Middleware{Base: b}
}
