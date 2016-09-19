package web

import (
	"github.com/rs/xhandler"
	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"net/http"
	"time"
)

func AppengineHandler(next xhandler.HandlerC) xhandler.HandlerC {
	return xhandler.HandlerFuncC(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		ctx = appengine.NewContext(r)
		ctx, _ = context.WithTimeout(ctx, time.Duration(30)*time.Second)
		next.ServeHTTPC(ctx, w, r)
	})
}

func AppengineTaskHandler(next xhandler.HandlerC) xhandler.HandlerC {
	return xhandler.HandlerFuncC(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		ctx = appengine.NewContext(r)
		next.ServeHTTPC(ctx, w, r)
	})
}
