package runtime

import (
	"context"
	"net/http"

	"github.com/devil-dwj/go-wms/base/database/redis"
)

type paramKey struct{}

func NewParamContext(ctx context.Context, param string) context.Context {
	return context.WithValue(ctx, paramKey{}, param)
}

func ParamFromContext(ctx context.Context) (req string, ok bool) {
	req, ok = ctx.Value(paramKey{}).(string)
	return
}

type requestKey struct{}

func NewRequestContext(ctx context.Context, req *http.Request) context.Context {
	return context.WithValue(ctx, requestKey{}, req)
}

func RequestFromContext(ctx context.Context) (req *http.Request, ok bool) {
	req, ok = ctx.Value(requestKey{}).(*http.Request)
	return
}

type redisKey struct{}

func NewRedisContext(ctx context.Context, r interface{}) context.Context {
	return context.WithValue(ctx, redisKey{}, r)
}

func RedisFromContext(ctx context.Context) (r redis.Basic, ok bool) {
	r, ok = ctx.Value(redisKey{}).(redis.Basic)
	return
}
