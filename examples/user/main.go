package main

import (
	"context"
	"fmt"

	"github.com/devil-dwj/go-wms/api/engine"
	"github.com/devil-dwj/go-wms/api/middleware"
	"github.com/devil-dwj/go-wms/api/runtime"
	"github.com/devil-dwj/go-wms/examples/pb"
	"github.com/devil-dwj/go-wms/log"
	"google.golang.org/protobuf/types/known/emptypb"
)

type UserServer struct {
}

func NewUserServer() *UserServer {
	return &UserServer{}
}

func (u *UserServer) Login(
	context context.Context,
	req *pb.LoginReq,
) (*pb.LoginRsp, error) {

	return &pb.LoginRsp{
		Passport: 1,
		Name:     req.Account,
		Role:     "role",
	}, nil
}

func (u *UserServer) Info(context context.Context, req *emptypb.Empty) (*pb.LoginRsp, error) {
	return &pb.LoginRsp{
		Passport: 1,
		Name:     "david",
		Role:     "role",
	}, nil
}

func (u *UserServer) UserInfo(
	context context.Context,
	req *pb.LoginReq,
) (*pb.LoginRsp, error) {

	return &pb.LoginRsp{
		Passport: 1,
		Name:     req.Account,
		Role:     "role",
	}, nil
}

type Config struct {
	WmsDSN string
	Port   uint16
}

type AuthMiddle struct {
}

func (a *AuthMiddle) Auth(record *middleware.MiddleWareRecord) {
	fmt.Println("auth !! ", record.Request.Header)
}

func main() {

	//c := new(Config)
	//config.MustLoad("config.json", c)

	l := log.MustLog("examples/example.log")
	auth := &AuthMiddle{}

	apiOpts := []runtime.ApiOption{
		runtime.WithEngine(engine.NewGinEngine(l)),
		runtime.WithPort(8686),
		runtime.ChainMiddle(
			middleware.Logger,
			middleware.Recovery,
		),
	}

	a := runtime.NewApi(apiOpts...)
	a.Use(auth.Auth)

	userServer := NewUserServer()
	pb.RegisterUserRouter(a, userServer)

	a.Run()
}
