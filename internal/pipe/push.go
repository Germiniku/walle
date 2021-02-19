/**
* @Time: 2021/1/28 下午3:30
* @Author: miku
* @File: push
* @Version: 1.0.0
* @Description:
 */

package pipe

import (
	log "github.com/golang/glog"
	"walle/api/comet"
	pb "walle/api/logic"
	"walle/api/protocol"
)

func (p *Pipe) push(msg *pb.PushMsg) {
	switch msg.Type {
	case pb.PushMsg_PUSH:
		{
			p.pushKeys(msg.Operation, msg.Server, msg.Keys, msg.Msg)
		}
	case pb.PushMsg_ROOM:
		{
			p.broadcastRoom(msg.Operation, msg.Room, msg.Msg)
		}
	case pb.PushMsg_BROADCAST:
		{
			p.broadcast(msg.Operation, msg.Msg)
		}
	}
}

func (p *Pipe) broadcastRoom(operation int32, room string, msg []byte) {
	if err := p.getRoom(room).Push(operation, msg); err != nil {
		log.Errorf("broadcastRoom push message error:%v", err)
	}
	return
}

func (p *Pipe) broadcast(operation int32, msg []byte) {
	req := &comet.BroadcastReq{
		Proto:   &protocol.Proto{Op: protocol.OpRaw, Body: msg},
		ProtoOp: operation,
	}
	for serverID, server := range p.cometServer {
		log.Infof("push broadcast message serverID:%s ", serverID)
		server.Broadcast(req)
	}
	return
}

func (p *Pipe) pushKeys(operation int32, serverID string, keys []string, msg []byte) {
	req := &comet.PushMsgReq{
		Proto:   &protocol.Proto{Op: protocol.OpRaw, Body: msg},
		Keys:    keys,
		ProtoOp: operation,
	}
	if svc, ok := p.cometServer[serverID]; ok {
		log.Infof("push one message to serverID:%s", serverID)
		svc.push(req)
	}
	return
}

func (p *Pipe) broadcastRoomByBatch(roomID string, proto *protocol.Proto) (err error) {
	args := comet.BroadcastRoomReq{
		RoomID: roomID,
		Proto:  proto,
	}
	for serverID, server := range p.cometServer {
		if err := server.BroadcastRoom(&args); err != nil {
			log.Errorf("server.BroadcastRoom(%v) roomID:%s serverID:%s error(%v)", args, roomID, serverID, err)
		}
	}
	log.Infof("broadcastRoom comets:%d", len(p.cometServer))
	return
}
