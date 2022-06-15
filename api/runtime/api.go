package runtime

import (
	"context"
	"fmt"

	"github.com/devil-dwj/go-wms/api/middleware"
)

type RequestKey struct{}

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

type MiddlewareFunc func(v *middleware.MiddleWareRecord) error

type Engine interface {
	RegisterHandler(int, EngineHandler)
	Log(MiddlewareFunc)
	Use(...MiddlewareFunc)
	POST(path string)
	GET(path string)
	Run() error
}

type Api struct {
	opts    apiOptions
	Routers map[string]*routerInfo
}

type apiOptions struct {
	Engine
	port   int
	log    MiddlewareFunc
	chain  []MiddlewareFunc
	static []string
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

func WithPort(p int) ApiOption {
	return newFuncApiOption(func(ao *apiOptions) {
		ao.port = p
	})
}

func WithLog(f MiddlewareFunc) ApiOption {
	return newFuncApiOption(func(ao *apiOptions) {
		ao.log = f
	})
}

func ChainMiddle(funcs ...MiddlewareFunc) ApiOption {
	return newFuncApiOption(func(ao *apiOptions) {
		ao.chain = append(ao.chain, funcs...)
	})
}

func WithStatic(path []string) ApiOption {
	return newFuncApiOption(func(ao *apiOptions) {
		ao.static = path
	})
}

func NewApi(opt ...ApiOption) *Api {
	opts := apiOptions{}
	for _, o := range opt {
		o.apply(&opts)
	}

	a := &Api{
		opts:    opts,
		Routers: make(map[string]*routerInfo),
	}

	if opts.log != nil {
		a.Log(opts.log)
	}

	chain := opts.chain
	for _, c := range chain {
		a.Use(c)
	}

	if len(opts.static) != 0 {
		if l, ok := a.opts.Engine.(interface {
			Static(path string)
		}); ok {
			for _, path := range opts.static {
				l.Static(path)
			}
		}
	}

	a.opts.Engine.RegisterHandler(a.opts.port, a.engineBackHandler)

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
		if d.Method == "POST" {
			a.opts.Engine.POST(d.Path)
		} else if d.Method == "GET" {
			a.opts.Engine.GET(d.Path)
		}
	}

	a.Routers[rd.ServiceName] = info
}

func (a *Api) Log(f MiddlewareFunc) {
	a.opts.Engine.Log(f)
}

func (a *Api) Use(middle ...MiddlewareFunc) {
	a.opts.Engine.Use(middle...)
}

func (a *Api) Run() error {
	fmt.Println("start api server ", a.opts.port)

	return a.opts.Engine.Run()
}

func (a *Api) engineBackHandler(
	path string,
	dec func(interface{}) error,
	ctx context.Context,
) (interface{}, error) {
	for _, info := range a.Routers {
		methodDesc, ok := info.methods[path]
		if ok {
			return methodDesc.Handler(
				info.serveImpl,
				ctx,
				dec,
			)
		}
	}

	return nil, fmt.Errorf("not find register method: %s", path)
}
