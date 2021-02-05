/**
* @Time: 2021/1/31 下午12:02
* @Author: miku
* @File: operations
* @Version: 1.0.0
* @Description:
 */

package comet

import (
	"context"
	"github.com/Terry-Mao/goim/pkg/strings"
	log "github.com/golang/glog"
	"time"
	"walle/api/logic"
	"walle/api/protocol"
)

// Connect connected a connection
func (s *Server) Connect(c context.Context, p *protocol.Proto, cookie string) (mid int64, key, rid string, accepts []int32, heartbeat time.Duration, err error) {
	var reply logic.ConnectReply
	if err = s.rpcClient.Call(c, "Connect", &logic.ConnectReq{
		Server: s.ServerId,
		Cookie: cookie,
		Token:  p.Body,
	}, &reply); err != nil {
		return
	}
	return reply.Mid, reply.Key, reply.RoomID, reply.Accepts, time.Duration(reply.Heartbeat), nil
}

// Operate 管理websocket发送过来的请求操作
func (s *Server) Operate(c context.Context, ch *Channel, p *protocol.Proto, b *Bucket) (err error) {
	switch p.Op {
	case protocol.OpChangeRoom:
		if s.conf.Debug {
			log.Infof("接收到了切换房间的消息 切换到roomID:%s", string(p.Body))
		}
		if err := b.ChangeRoom(string(p.Body), ch); err != nil {
			log.Errorf("b.ChangeRoom(%s) error%v", string(p.Body), err)
		}
		p.Op = protocol.OpChangeRoomReply
	case protocol.OpSub:
		if ops, err := strings.SplitInt32s(string(p.Body), ","); err == nil {
			ch.Watch(ops...)
		}
		p.Op = protocol.OpSubReply
	case protocol.OpUnsub:
		if ops, err := strings.SplitInt32s(string(p.Body), ","); err == nil {
			ch.UnWatch(ops...)
		}
		p.Op = protocol.OpSubReply
	default:
		// TODO: 其他操作
	}
	ch.WriteWebsocketMessage(p)
	return nil
}

// DisConnect disconnect a connection
func (s *Server) DisConnect(c context.Context, mid int64, key string) (err error) {
	var reply logic.DisconnectReply
	err = s.rpcClient.Call(c, "DisConnect", &logic.DisconnectReq{
		Server: s.ServerId,
		Key:    key,
		Mid:    mid,
	}, &reply)
	return
}

// Heartbeat heartbeat a connection session.
func (s *Server) Heartbeat(c context.Context, mid int64, key string) (err error) {
	var reply logic.HeartbeatReply
	err = s.rpcClient.Call(c, "Heartbeat", &logic.DisconnectReq{
		Server: s.ServerId,
		Key:    key,
		Mid:    mid,
	}, &reply)
	return
}
