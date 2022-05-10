package runtime

import (
	"github.com/devil-dwj/go-wms/base/database/redis"
)

type apiOptions struct {
	Engine
	log      MiddlewareFunc
	recovery MiddlewareFunc
	chain    []MiddlewareFunc
	trace    bool
	r        redis.Basic
}

type ApiOption interface {
	apply(*apiOptions)
}

type funcApiOption struct {
	f func(*apiOptions)
}

func (fdo *funcApiOption) apply(do *apiOptions) {
	fdo.f(do)
}

func newFuncApiOption(f func(*apiOptions)) *funcApiOption {
	return &funcApiOption{f: f}
}

func WithEngine(en Engine) ApiOption {
	return newFuncApiOption(func(ao *apiOptions) {
		ao.Engine = en
	})
}

func WithLog(f MiddlewareFunc) ApiOption {
	return newFuncApiOption(func(ao *apiOptions) {
		ao.log = f
	})
}

func WithRecovery(f MiddlewareFunc) ApiOption {
	return newFuncApiOption(func(ao *apiOptions) {
		ao.recovery = f
	})
}

func ChainMiddle(funcs ...MiddlewareFunc) ApiOption {
	return newFuncApiOption(func(ao *apiOptions) {
		ao.chain = append(ao.chain, funcs...)
	})
}

func WithTracing(b bool) ApiOption {
	return newFuncApiOption(func(ao *apiOptions) {
		ao.trace = b
	})
}

func WithRedis(r redis.Basic) ApiOption {
	return newFuncApiOption(func(ao *apiOptions) {
		ao.r = r
	})
}
