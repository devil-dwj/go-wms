package main

import (
	"github.com/devil-dwj/go-wms/api"
	"github.com/devil-dwj/go-wms/config"
	"github.com/devil-dwj/go-wms/database/mysql"
	"github.com/devil-dwj/go-wms/examples/pb"
	"github.com/devil-dwj/go-wms/log"
)

type UserServer struct {
	prps pb.UserProcedure
}

func NewUserServer(pr pb.UserProcedure) pb.UserHandler {
	return &UserServer{prps: pr}
}

func (u *UserServer) Login(req *pb.LoginReq) (*pb.LoginRsp, error) {
	_ = u.prps.GetRawDB()

	return &pb.LoginRsp{
		Passport: 1,
		Name:     "dwj",
		Role:     "role",
	}, nil
}

type Config struct {
	WmsDSN string
	Port   uint16
}

func main() {

	c := new(Config)
	config.MustLoad("config.json", c)

	l := log.MustLog("examples/example.log")

	a := api.New(8866, l)

	userServer := NewUserServer(pb.NewUserProcedure(mysql.WmsDB))
	pb.RegisterUserRouters(a, userServer)

	a.Run()
}
