package handler

import (
	"net/http"
	"strings"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"
)

const (
	AllowHeadersKey = "Access-Control-Allow-Headers"
	AllowedHeaders  = "Accept, Content-Type, Content-Length, Accept-Encoding, Authorization, ResponseType"
	AllowOriginKey  = "Access-Control-Allow-Origin"
	AllowMethodsKey = "Access-Control-Allow-Methods"
	AllowedMethods  = "GET, POST"
)

// GRPC handler allows for incoming connections to be handled by either GRPC service or RESTful GRPC-Gateway.
func GRPC(grpcServer *grpc.Server, otherHandler http.Handler) http.Handler {
	return h2c.NewHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.ProtoMajor == 2 && strings.Contains(r.Header.Get("Content-Type"), "application/grpc") {
			grpcServer.ServeHTTP(w, r)
		} else {
			otherHandler.ServeHTTP(w, r)
		}
	}), &http2.Server{})
}

// CORS allows for Cross-Origin Request Sharing for all domains (*) for all routes on a handler.
// See https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS for more details.
func CORS(handler http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(AllowHeadersKey, AllowedHeaders)
		w.Header().Set(AllowOriginKey, r.Header.Get("Origin"))
		w.Header().Set(AllowMethodsKey, AllowedMethods)

		if r.Method == "OPTIONS" {
			return
		}

		handler.ServeHTTP(w, r)
	}
}
