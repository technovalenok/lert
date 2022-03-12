package client

import (
	"net/http"

	"go.uber.org/zap"
)

type LoggingRoundTripper struct {
	Wrapped http.RoundTripper
}

func (lrt LoggingRoundTripper) RoundTrip(req *http.Request) (res *http.Response, e error) {
	zap.S().Infof("Request -> %s %s", req.Method, req.URL)
	if res, e = lrt.Wrapped.RoundTrip(req); e != nil {
		zap.S().Errorf("Error: %v", e)
		return nil, e
	} else {
		zap.S().Infof("Response -> status: %s", res.Status) // TODO: how to log body?
	}
	return
}

func NewHTTPClient() *http.Client {
	return &http.Client{
		Transport: LoggingRoundTripper{http.DefaultTransport},
	}
}
