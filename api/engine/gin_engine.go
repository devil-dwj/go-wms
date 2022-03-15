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
	"github.com/gin-gonic/gin"
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

func NewGinEngine(l *zap.Logger) *GinEngine {
	e := &GinEngine{
		Engine: gin.New(),
		l:      l,
	}

	e.Engine.Use(Cors())

	return e
}

func (engine *GinEngine) RegisterHandler(port int, handler runtime.EngineHandler) {
	engine.port = port
	engine.handler = handler
}

func (engine *GinEngine) Use(handlers ...runtime.MiddlewareFunc) {
	for _, handler := range handlers {
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

			handler(r)
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
				name = strings.Split(name, ",")[0]
				if name == "" {
					continue
				}

				param := c.Query(name)
				fieldType := fieldInfo.Type.Name()
				if fieldType == "int32" {
					paramInt, err := strconv.Atoi(param)
					if err != nil {
						return err
					}

					refV.FieldByName(fieldInfo.Name).Set(reflect.ValueOf(int32(paramInt)))
				} else {

					refV.FieldByName(fieldInfo.Name).Set(reflect.ValueOf(param))
				}

			}

			return nil
		}

		reply, err := engine.handler(path, df, c.Request.Context())
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

		reply, err := engine.handler(path, df, c.Request.Context())
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
	c.JSON(
		http.StatusBadRequest,
		gin.H{
			"code": 1,
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
