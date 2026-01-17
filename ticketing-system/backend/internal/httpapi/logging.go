package httpapi

import (
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

func requestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		next.ServeHTTP(ww, r)
		reqID := middleware.GetReqID(r.Context())
		log.Printf("request method=%s path=%s status=%d duration=%s req_id=%s", r.Method, r.URL.Path, ww.Status(), time.Since(start), reqID)
	})
}

func logRequestError(r *http.Request, message string, err error) {
	reqID := middleware.GetReqID(r.Context())
	if err != nil {
		log.Printf("request_error message=%s error=%s req_id=%s", message, err.Error(), reqID)
		return
	}
	log.Printf("request_error message=%s req_id=%s", message, reqID)
}
