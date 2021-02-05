/**
* @Time: 2021/1/31 下午7:44
* @Author: miku
* @File: server
* @Version: 1.0.0
* @Description:
 */

package rpcx

import (
	"context"
	"encoding/json"
	"fmt"
	log "github.com/golang/glog"
	"github.com/rcrowley/go-metrics"
	"github.com/rpcxio/rpcx-etcd/serverplugin"
	"github.com/smallnest/rpcx/server"
	"time"
	pb "walle/api/comet"
	"walle/internal/comet"
	"walle/internal/comet/conf"
	"walle/internal/comet/errors"
)

type Server struct {
	conf      *conf.Config
	Comet     *comet.Server
	rpcServer *server.Server
	addr      string
}

// New comet rpcx server
func New(conf *conf.Config, comet *comet.Server) *Server {
	s := &Server{
		addr:  conf.RPCServer.Addr,
		Comet: comet,
		conf:  conf,
	}
	s.Register()
	go s.Serve()
	return s
}

// discoveryMetadata 注册服务元数据
func (s *Server) discoveryMetadata() string {
	var metadata struct {
		ServerId string `json:"server_id"`
		Ip       string `json:"ip"`
	}
	var (
		data []byte
		err  error
	)
	metadata.Ip = "tcp@" + s.conf.RPCServer.Addr
	metadata.ServerId = s.Comet.ServerId
	if data, err = json.Marshal(&metadata); err != nil {
		log.Errorf("json.Marshal metadata:%v error:%v", metadata, err)
		data = nil
	}
	return string(data)
}

// Register 服务注册
func (s *Server) Register() {
	server := server.NewServer()
	s.rpcServer = server
	s.addRegistryPlugin(s.conf)
	metadata := s.discoveryMetadata()
	s.rpcServer.RegisterName("comet", s, metadata)
}

func (s *Server) addRegistryPlugin(conf *conf.Config) {
	r := &serverplugin.EtcdV3RegisterPlugin{
		ServiceAddress: "tcp@" + conf.RPCServer.Addr,
		EtcdServers:    conf.Discovery.Endpoints,
		//EtcdServers:    conf.Discovery.Endpoints,
		BasePath:       fmt.Sprintf("/%s/%s", conf.Discovery.Env, conf.Discovery.AppId),
		Metrics:        metrics.NewRegistry(),
		UpdateInterval: time.Minute,
	}
	err := r.Start()
	if err != nil {
		return
	}
	s.rpcServer.Plugins.Add(r)
}

func (s *Server) Serve() {
	go s.rpcServer.Serve("tcp", s.conf.RPCServer.Addr)
}

// PushMsg 单独推送消息
func (s *Server) PushMsg(c context.Context, args *pb.PushMsgReq, reply *pb.PushMsgReply) (err error) {
	if len(args.Keys) == 0 {
		return errors.ErrPushMsgArg
	}
	log.Infof("PushMsg: keys:%v protoOp:%d op:%d", args.Keys, args.ProtoOp, args.Proto.Op)
	for _, key := range args.Keys {
		log.Infof("PushMsg: key:%s", key)
		if channel := s.Comet.Bucket(key).Channel(key); channel != nil {
			log.Infof("走到了这里")
			if !channel.NeedPush(args.ProtoOp) {
				log.Infof("这条消息不需要转发")
				continue
			}
			log.Infof("这条消息需要转发")
			channel.Push(args.Proto)
		}
	}
	return nil
}

// Broadcast 全频道广播消息
func (s *Server) Broadcast(c context.Context, args *pb.BroadcastReq, reply *pb.BroadcastRoomReply) (err error) {
	log.Infof("接收到了广播消息 proto:%v", args.Proto)
	if args.Proto == nil {
		return errors.ErrBroadCastArg
	}
	go func() {
		for _, bucket := range s.Comet.Buckets() {
			bucket.Broadcast(args.Proto, args.ProtoOp)
			if args.Speed > 0 {
				t := bucket.ChannelCount() / int(args.Speed)
				time.Sleep(time.Second * time.Duration(t))
			}
		}
	}()
	return nil
}

// BroadcastRoom 房间广播消息
func (s *Server) BroadcastRoom(c context.Context, args *pb.BroadcastRoomReq, reply *pb.BroadcastRoomReply) (err error) {
	if args.Proto == nil || args.RoomID == "" {
		return errors.ErrBroadCastRoomArg
	}
	go func() {
		for _, bucket := range s.Comet.Buckets() {
			bucket.BroadcastRoom(args)
		}
	}()
	return nil
}

// Rooms 获取当前所有房间
func (s *Server) Rooms(c context.Context, args *pb.RoomsReq, reply *pb.RoomsReply) (err error) {
	reply.Rooms = make(map[string]bool)
	for _, bucket := range s.Comet.Buckets() {
		for roomID := range bucket.Rooms() {
			reply.Rooms[roomID] = true
		}
	}
	return nil
}

func (s *Server) Close() {
	s.rpcServer.UnregisterAll()
	s.rpcServer.Close()
}
