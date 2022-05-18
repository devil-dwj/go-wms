package engine

import (
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/devil-dwj/go-wms/api/middleware"
	"github.com/devil-dwj/go-wms/api/runtime"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.uber.org/zap"
)

func init() {
	gin.SetMode(gin.ReleaseMode)
}

type GinEngine struct {
	*gin.Engine
	l       *zap.Logger
	port    int
	handler runtime.EngineHandler
}

func NewGinEngine(port int, l *zap.Logger) *GinEngine {
	e := &GinEngine{
		Engine: gin.New(),
		port:   port,
		l:      l,
	}

	e.Engine.Use(Cors())

	return e
}

func (engine *GinEngine) Handler(handler runtime.EngineHandler) {
	engine.handler = handler
}

func (engine *GinEngine) Log(handler runtime.MiddlewareFunc) {
	engine.Engine.Use(func(ctx *gin.Context) {
		r := &middleware.MiddleWareRecord{
			Logger:  engine.l,
			Request: ctx.Request,
			Start:   time.Now(),
		}

		ctx.Next()

		r.Status = ctx.Writer.Status()
		r.Err = ctx.Errors.String()
		r.Cost = time.Since(r.Start)

		_ = handler(r)
	})
}

func (engine *GinEngine) Tracer(serverName string) {
	engine.Engine.Use(otelgin.Middleware(serverName))
}

func (engine *GinEngine) Static(path string) {
	engine.Engine.StaticFS(path, http.Dir(path))
}

func (engine *GinEngine) Use(handlers ...runtime.MiddlewareFunc) {
	for _, handler := range handlers {
		engine.Engine.Use(func(ctx *gin.Context) {
			r := &middleware.MiddleWareRecord{
				Logger:  engine.l,
				Request: ctx.Request,
				Start:   time.Now(),
			}

			err := handler(r)
			if err != nil {
				engine.fail(ctx, err)
				ctx.Abort()
				return
			}

			ctx.Next()
		})
	}
}

func (engine *GinEngine) GET(path string) {
	engine.Engine.GET(path, func(c *gin.Context) {
		df := func(v interface{}) error {
			refV := reflect.ValueOf(v).Elem()

			for i := 0; i < refV.NumField(); i++ {
				fieldInfo := refV.Type().Field(i)
				tag := fieldInfo.Tag
				name := tag.Get("json")
				arr := strings.Split(name, ",")
				if len(arr) < 1 {
					continue
				}
				name = arr[0]
				if name == "" {
					continue
				}

				param := c.Query(name)
				fieldType := fieldInfo.Type.Name()
				if fieldType == "int32" {
					paramInt, err := strconv.Atoi(param)
					if err != nil {
						return fmt.Errorf("query parma [%s]", name)
					}
					refV.FieldByName(fieldInfo.Name).Set(reflect.ValueOf(int32(paramInt)))
				} else if fieldType == "int64" {
					paramInt, err := strconv.Atoi(param)
					if err != nil {
						return fmt.Errorf("query parma [%s]", name)
					}
					refV.FieldByName(fieldInfo.Name).Set(reflect.ValueOf(int64(paramInt)))
				} else {
					refV.FieldByName(fieldInfo.Name).Set(reflect.ValueOf(param))
				}
			}

			return nil
		}

		ctx := runtime.NewRequestContext(c.Request.Context(), c.Request)
		reply, err := engine.handler(path, df, ctx)
		if err != nil {
			engine.fail(c, err)
		} else {
			engine.success(c, reply)
		}
	})
}

func (engine *GinEngine) POST(path string) {
	engine.Engine.POST(path, func(c *gin.Context) {
		df := func(v interface{}) error {
			if err := c.ShouldBind(v); err != nil {
				return err
			}

			return nil
		}

		ctx := runtime.NewRequestContext(c.Request.Context(), c.Request)
		reply, err := engine.handler(path, df, ctx)
		if err != nil {
			engine.fail(c, err)
		} else {
			engine.success(c, reply)
		}
	})
}

func (engine *GinEngine) Run() error {
	return engine.Engine.Run(fmt.Sprintf(":%d", engine.port))
}

func (engine *GinEngine) fail(c *gin.Context, err error) {
	var code int32 = 1
	if e, ok := err.(interface {
		Code() int32
	}); ok {
		code = e.Code()
	}
	c.Error(err)
	c.JSON(
		http.StatusBadRequest,
		gin.H{
			"code": code,
			"msg":  err.Error(),
			"data": "",
		})
}

func (engine *GinEngine) success(c *gin.Context, data interface{}) {
	c.JSON(
		http.StatusOK,
		gin.H{
			"code": 0,
			"msg":  "",
			"data": data,
		})
}

func Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method

		origin := c.Request.Header.Get("Origin")

		if origin != "" {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Headers", "Content-Type,AccessToken,X-CSRF-Token, Authorization") //自定义 Header
			c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
			c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Content-Type")
			c.Header("Access-Control-Allow-Credentials", "true")

		}

		if method == "OPTIONS" {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Headers", "Content-Type,AccessToken,X-CSRF-Token, Authorization") //自定义 Header
			c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
			c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Content-Type")
			c.Header("Access-Control-Allow-Credentials", "true")
			c.AbortWithStatus(http.StatusNoContent)
		}

		c.Next()
	}
}

func CorsContrib() gin.HandlerFunc {
	return cors.New(
		cors.Config{
			AllowAllOrigins: false,
			AllowOrigins:    []string{"*"},
			AllowMethods: []string{
				"OPTIONS",
				"GET",
				"POST",
				"PUT",
				"PATCH",
				"DELETE",
				"FETCH",
			},
			AllowHeaders:           []string{"Authorization, Content-Length, X-CSRF-Token, Token,session", "Content-Type", "x-requested-with"},
			AllowCredentials:       true,
			ExposeHeaders:          []string{"Content-Length", "Content-Type"},
			MaxAge:                 86400,
			AllowWildcard:          true,
			AllowBrowserExtensions: true,
			AllowWebSockets:        true,
			AllowFiles:             true,
		},
	)
}
