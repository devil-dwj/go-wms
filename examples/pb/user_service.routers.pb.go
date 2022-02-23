// Code generated by protoc-gen-jjpms. DO NOT EDIT.
// source: user_service.proto

package pb

import (
	"github.com/devil-dwj/go-wms/api"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

var (
	_, _ = strconv.Atoi("1")
)

type UserHandler interface {
	Login(req *LoginReq) (*LoginRsp, error)
}

type UserRouter interface {
	Login(c *gin.Context)
}

type User_Router struct {
	UserHandler
}

func RegisterUserRouters(a api.Api, h UserHandler) {
	r := &User_Router{h}
	a.POST("api/v1/login", r.Login)
}

func (h *User_Router) Login(c *gin.Context) {
	req := &LoginReq{}
	if err := c.ShouldBindJSON(req); err != nil {
		h.fail(c, err)
		return
	}

	rsp, err := h.UserHandler.Login(req)
	if err != nil {
		h.fail(c, err)
		return
	}

	h.returnBack(c, err, rsp)
}

func (h *User_Router) fail(c *gin.Context, err error) {
	c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": err.Error(), "data": ""})
}

func (h *User_Router) success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "", "data": data})
}

func (h *User_Router) returnBack(c *gin.Context, err error, data interface{}) {
	if err != nil {
		h.fail(c, err)
	} else {
		h.success(c, data)
	}
}
