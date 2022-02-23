package main

import (
	"github.com/devil-dwj/go-wms/api"
	"github.com/devil-dwj/go-wms/examples/pb"
	"github.com/devil-dwj/go-wms/log"
)

type UserServer struct {
}

func NewUserServer() pb.UserHandler {
	return &UserServer{}
}

func (u *UserServer) Login(req *pb.LoginReq) (*pb.LoginRsp, error) {
	return &pb.LoginRsp{
		Passport: 1,
		Name:     "dwj",
		Role:     "role",
	}, nil
}

func main() {

	l := log.MustLog("examples/example.log")

	a := api.New(8866, l)

	userServer := NewUserServer()
	pb.RegisterUserRouters(a, userServer)

	a.Run()
}
