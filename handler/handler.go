package handler

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

type Http struct {
	handler http.Handler
}

func (l *Http) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	zap.S().Infof("<- Request: %s %s", r.Method, r.URL)
	start := time.Now()
	l.handler.ServeHTTP(w, r)
	zap.S().Infof("<- Response: %v sec", time.Since(start))
}

func NewHttpHandler(handlerToWrap http.Handler) *Http {
	return &Http{handlerToWrap}
}
