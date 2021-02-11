/**
* @Time: 2021/1/28 上午11:22
* @Author: miku
* @File: nats
* @Version: 1.0.0
* @Description:
 */

package mq

import (
	"context"
	log "github.com/golang/glog"
	"github.com/golang/protobuf/proto"
	"github.com/nats-io/nats.go"
	"strconv"
	"time"
	pb "walle/api/logic"
	"walle/internal/logic/conf"
	"walle/protocol"
)

type Nats struct {
	conf *conf.MQ
	conn *nats.Conn
}

func newNATS(conf *conf.MQ) (MQ, error) {
	conn, err := nats.Connect(conf.Addrs[0],
		nats.Timeout(time.Duration(conf.ConnTimeout)),
		nats.PingInterval(time.Duration(conf.PingInterval)*time.Second),
	)
	if err != nil {
		log.Errorf("nats addr:", conf.Addrs)
		return nil, err
	}
	log.Infof("logic connect Nats Success")
	nat := &Nats{
		conn: conn,
		conf: conf,
	}
	return nat, nil
}

func (n *Nats) Close() {
	n.conn.Close()
}

func (n *Nats) BroadcastMsg(c context.Context, op int32, msg []byte) (err error) {
	pushMsg := &pb.PushMsg{
		Type:      pb.PushMsg_BROADCAST,
		Operation: op,
		Msg:       msg,
	}
	bytes, err := proto.Marshal(pushMsg)
	if err != nil {
		return
	}
	key := strconv.FormatInt(int64(op), 10)
	m := &protocol.ProducerMessage{
		Key:   key,
		Topic: n.conf.Topic,
		Value: bytes,
	}
	log.Infof("send broadcast msg key:%n topic:%s", key, n.conf.Topic)
	n.sendMessage(m)
	return nil
}

func (n *Nats) BroadcastRoomMsg(c context.Context, op int32, Type, roomID string, msg []byte) (err error) {
	pushMsg := &pb.PushMsg{
		Type:      pb.PushMsg_ROOM,
		Operation: op,
		Msg:       msg,
		Room:      roomID,
	}
	bytes, err := proto.Marshal(pushMsg)
	if err != nil {
		return
	}
	key := strconv.FormatInt(int64(op), 10)
	m := &protocol.ProducerMessage{
		Key:   key,
		Topic: n.conf.Topic,
		Value: bytes,
	}
	log.Infof("send broadcast room msg key:%n topic:%s", key, n.conf.Topic)
	n.sendMessage(m)
	return nil
}

func (n *Nats) PushMsg(c context.Context, op int32, server string, keys []string, msg []byte) (err error) {
	pushMsg := &pb.PushMsg{
		Type:      pb.PushMsg_PUSH,
		Operation: op,
		Server:    server,
		Keys:      keys,
		Msg:       msg,
	}
	bytes, err := proto.Marshal(pushMsg)
	if err != nil {
		return
	}
	m := &protocol.ProducerMessage{
		Key:   keys[0],
		Topic: n.conf.Topic,
		Value: bytes,
	}
	log.Infof("send message key:%s topic:%s", keys[0], n.conf.Topic)
	n.sendMessage(m)
	return nil
}

// SendMessage 将msg发布到NATS
func (n *Nats) sendMessage(msg *protocol.ProducerMessage) {
	key := protocol.PackagePublishKey(msg.Topic, msg.Key)
	log.Infof("logic send nat one message key:%s value:%s", key, string(msg.Value))
	if err := n.conn.Publish(key, msg.Value); err != nil {
		log.Errorf("[sendMessage] n.Nats.Publish key:%s", key)
	}
	return
}
