package runtime

import (
	"context"
	"fmt"

	"github.com/devil-dwj/go-wms/api/middleware"
	"github.com/opentracing/opentracing-go"
	"go.uber.org/zap"
)

type methodHandler func(
	srv interface{},
	ctx context.Context,
	dec func(interface{}) error,
) (interface{}, error)

type MethodDesc struct {
	Name    string
	Method  string
	Path    string
	Handler methodHandler
}

type routerInfo struct {
	serviceName string
	serveImpl   interface{}
	methods     map[string]*MethodDesc
}

// 注册路由
type RouterRegistrar interface {
	RegisterRouter(desc *RouterDesc, impl interface{})
}

type RouterDesc struct {
	ServiceName string
	Methods     []MethodDesc
}

type EngineHandler func(
	path string,
	dec func(interface{}) error,
	ctx context.Context,
) (
	interface{},
	error,
)

type MiddlewareFunc func(context.Context, *middleware.MiddleWareRecord) error

type RestRegister func(string)

// 选择实现 log
type Log interface {
	Log(MiddlewareFunc)
}

// 选择实现 tracing
type Tracing interface {
	Tracer(opentracing.Tracer, MiddlewareFunc)
}

// 实现rest
type Engine interface {
	Handler(EngineHandler)
	Use(...MiddlewareFunc)
	POST(path string)
	GET(path string)
	Run() error
}

type Api struct {
	opts         apiOptions
	l            *zap.Logger
	routers      map[string]*routerInfo
	restHandlers map[string]RestRegister
}

func NewApi(l *zap.Logger, opt ...ApiOption) *Api {
	opts := apiOptions{}
	for _, o := range opt {
		o.apply(&opts)
	}

	a := &Api{
		l:            l,
		opts:         opts,
		routers:      make(map[string]*routerInfo),
		restHandlers: make(map[string]RestRegister),
	}

	chain := opts.chain
	for _, c := range chain {
		a.Use(c)
	}

	if opts.log != nil {
		if l, ok := opts.Engine.(interface {
			Log(MiddlewareFunc)
		}); ok {
			l.Log(opts.log)
		}
	}

	if opts.trace {
		if l, ok := opts.Engine.(interface {
			Tracer(string)
		}); ok {
			l.Tracer("user-service")
		}
	}

	if opts.static != "" {
		if l, ok := opts.Engine.(interface {
			Static(path string)
		}); ok {
			l.Static(opts.static)
		}
	}

	a.restRegist()
	a.handler()

	return a
}

func (a *Api) RegisterRouter(rd *RouterDesc, srv interface{}) {
	info := &routerInfo{
		serviceName: rd.ServiceName,
		serveImpl:   srv,
		methods:     make(map[string]*MethodDesc),
	}

	for i := range rd.Methods {
		d := &rd.Methods[i]
		info.methods[d.Path] = d
		if h, ok := a.restHandlers[d.Method]; ok {
			h(d.Path)
		}
	}

	a.routers[rd.ServiceName] = info
}

func (a *Api) Use(middle ...MiddlewareFunc) {
	a.opts.Engine.Use(middle...)
}

func (a *Api) Run() error {
	fmt.Println("start api server")

	return a.opts.Engine.Run()
}

func (a *Api) restRegist() {
	a.restHandlers["GET"] = a.opts.GET
	a.restHandlers["POST"] = a.opts.Engine.POST
}

func (a *Api) handler() {
	a.opts.Engine.Handler(func(path string, dec func(interface{}) error, ctx context.Context) (interface{}, error) {
		for _, info := range a.routers {
			if md, ok := info.methods[path]; ok {
				if a.opts.recovery != nil {
					defer func() {
						if err := recover(); err != nil {
							if r, b := RequestFromContext(ctx); b {
								a.opts.recovery(ctx, &middleware.MiddleWareRecord{
									Logger:  a.l,
									Request: r,
									Err:     err,
								})
							}
						}
					}()
				}

				passCtx := NewRedisContext(ctx, a.opts.r)
				return md.Handler(
					info.serveImpl,
					passCtx,
					dec,
				)
			}
		}

		return nil, fmt.Errorf("not find register method: %s", path)
	})
}
