package main

import (
	"flag"
	"fmt"
	_ "net/http/pprof"
	"time"
	goserver "xjserver"
	pb "xjserver/example/msg"
	"xjserver/network"
	"xjserver/server"
)

var addr = flag.String("addr", "0.0.0.0:8080", "http service address")

type clientMgr struct {
	clients map[network.Session]struct{}
}

var BServerID int32 = 101

func main() {

	mgr := &clientMgr{
		clients: make(map[network.Session]struct{}),
	}

	g, err := server.NewGate(100, *addr, goserver.Config{Nats: "127.0.0.1:4222"})
	if err != nil {
		fmt.Println(err)
		return
	}

	g.RegisterNetWorkEvent(func(conn network.Session) {
		mgr.clients[conn] = struct{}{}
	}, func(conn network.Session) {
		delete(mgr.clients, conn)
	})
	g.RouteSessionMsg((*pb.ReqHello)(nil), BServerID)

	g.Run()

	// 向B服务器发送消息
	g.GetServerById(BServerID).Notify(&pb.ReqSend{Name: "我是send消息"})

	// 向B服务器发送request请求
	g.GetServerById(BServerID).Request(&pb.ReqRequest{
		Name: "我是request请求",
	}, func(resp *pb.RespRequest, err error) {
		if err != nil {
			return
		}
		fmt.Println("返回消息:", resp.Name)
	})

	// 向B服务器发送call请求
	resp := pb.RespCall{}
	err = g.GetServerById(BServerID).Call(&pb.ReqCall{Name: "我是call请求"}, &resp)
	if err != nil {
		return
	}
	fmt.Println("返回消息:", resp.Name)

	time.Sleep(time.Hour)
}
