/**
* @Time: 2021/1/29 下午2:36
* @Author: miku
* @File: rpcx
* @Version: 1.0.0
* @Description:
 */

package rpcx

import (
	"context"
	"fmt"
	log "github.com/golang/glog"
	"github.com/rcrowley/go-metrics"
	"github.com/rpcxio/rpcx-etcd/serverplugin"
	"github.com/smallnest/rpcx/server"
	"time"
	pb "walle/api/logic"
	"walle/internal/logic"
	"walle/internal/logic/conf"
)

type Server struct {
	conf      *conf.Config
	logic     *logic.Logic
	addr      string
	rpcServer *server.Server
}

func New(conf *conf.Config, logic *logic.Logic) *Server {
	s := &Server{
		conf:  conf,
		addr:  conf.RpcServer.Addr,
		logic: logic,
	}
	return s
}

// Register 服务注册
func (s *Server) Register() {
	server := server.NewServer()
	s.rpcServer = server
	s.addRegistryPlugin(s.conf)
	s.rpcServer.RegisterName("logic", s, "")
}

func (s *Server) addRegistryPlugin(conf *conf.Config) {
	fmt.Println("endpoints:%v", conf.Discovery.Endpoints)
	r := &serverplugin.EtcdV3RegisterPlugin{
		ServiceAddress: "tcp@" + conf.RpcServer.Addr,
		EtcdServers:    []string{"localhost:2379"},
		//EtcdServers:    conf.Discovery.Endpoints,
		BasePath:       fmt.Sprintf("/%s/%s", conf.Discovery.Env, conf.Discovery.AppId),
		Metrics:        metrics.NewRegistry(),
		UpdateInterval: time.Minute,
	}
	err := r.Start()
	if err != nil {
		log.Fatal(err)
	}
	s.rpcServer.Plugins.Add(r)
}

// Connect 连接 添加会话状态
func (s *Server) Connect(ctx context.Context, args *pb.ConnectReq, reply *pb.ConnectReply) error {
	mid, key, rid, accepts, err := s.logic.Connect(ctx, args.Server, args.Cookie, args.Token)
	if err != nil {
		return err
	}
	reply.Mid = mid
	reply.Key = key
	reply.RoomID = rid
	reply.Accepts = accepts
	reply.Heartbeat = int64(time.Second * 5)
	return nil
}

// Disconnect 断开连接 清除会话状态
func (s *Server) Disconnect(ctx context.Context, args *pb.DisconnectReq, reply *pb.DisconnectReply) (err error) {
	if err = s.logic.Disconnect(ctx, args.Mid, args.Key, args.Server); err != nil {
		reply.Done = false
		return
	} else {
		reply.Done = true
		return
	}
}

// Heartbeat 心跳保活
func (s *Server) Heartbeat(ctx context.Context, args *pb.HeartbeatReq, reply *pb.HeartbeatReply) (err error) {
	if err = s.logic.Heartbeat(ctx, args.Mid, args.Key, args.Server); err != nil {
		log.Errorf("[Heartbeat] s.logic.Heartbeart() error:%v", err)
	}
	log.Infof("heartbeat key:%s mid:%d", args.Key, args.Mid)
	return
}

func (s *Server) Serve() {
	s.Register()
	go s.rpcServer.Serve("tcp", s.conf.RpcServer.Addr)
}

func (s *Server) Close() {
	s.rpcServer.UnregisterAll()
	s.rpcServer.Close()
}
