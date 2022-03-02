package api

import (
	"fmt"

	"github.com/devil-dwj/go-wms/api/middleware"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func init() {
	gin.SetMode(gin.ReleaseMode)
}

type Api interface {
	GET(path string, handle gin.HandlerFunc) gin.IRoutes
	POST(path string, handle gin.HandlerFunc) gin.IRoutes
	Use(middleware ...gin.HandlerFunc) gin.IRoutes
	Run() error
}

type api struct {
	log *zap.Logger
	*gin.Engine
	port   uint16
	prefix string
	v      string
}

func New(p uint16, l *zap.Logger) Api {
	a := &api{
		log:    l,
		Engine: gin.New(),
		port:   p,
		prefix: "api",
		v:      "v1",
	}

	a.Use(middleware.WithCors())
	a.Use(middleware.WithLogger(l))
	a.Use(middleware.WithRecovery(true))

	return a
}

func (a *api) POST(path string, handle gin.HandlerFunc) gin.IRoutes {
	return a.Engine.POST(path, handle)
}

func (a *api) GET(path string, handle gin.HandlerFunc) gin.IRoutes {
	return a.Engine.GET(path, handle)
}

func (a *api) Use(middleware ...gin.HandlerFunc) gin.IRoutes {
	return a.Engine.Use(middleware...)
}

func (a *api) Run() error {
	a.log.Sugar().Infof("Start REST server at port: %d", a.port)
	return a.Engine.Run(fmt.Sprintf(":%d", a.port))
}
