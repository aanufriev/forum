package middleware

import (
	"context"
	"net/http"

	"github.com/aanufriev/forum/configs"

	"github.com/lithammer/shortuuid"
	"github.com/sirupsen/logrus"
)

type RequestID string

func log(r *http.Request) {
	logrus.WithFields(logrus.Fields{
		configs.RequestID: r.Context().Value(configs.RequestID),
		"method":          r.Method,
		"remote_addr":     r.RemoteAddr,
	}).Info(r.URL.Path)
}

func AccessLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := shortuuid.New()
		ctx := context.WithValue(r.Context(), RequestID(configs.RequestID), reqID)
		r = r.WithContext(ctx)
		log(r)
		next.ServeHTTP(w, r)
	})
}
