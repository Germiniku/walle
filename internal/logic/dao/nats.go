/**
* @Time: 2021/1/28 上午11:22
* @Author: miku
* @File: nats
* @Version: 1.0.0
* @Description:
 */

package dao

import (
	"context"
	log "github.com/golang/glog"
	"github.com/golang/protobuf/proto"
	"strconv"
	pb "walle/api/logic"
	"walle/protocol"
)

func (d *Dao) BroadcastMsg(c context.Context, op int32, msg []byte) (err error) {
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
		Topic: d.conf.Nats.Topic,
		Value: bytes,
	}
	log.Infof("send broadcast msg key:%d topic:%s", key, d.conf.Nats.Topic)
	d.sendMessage(m)
	return nil
}

func (d *Dao) BroadcastRoomMsg(c context.Context, op int32, Type, roomID string, msg []byte) (err error) {
	pushMsg := &pb.PushMsg{
		Type:      pb.PushMsg_ROOM,
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
		Topic: d.conf.Nats.Topic,
		Value: bytes,
	}
	log.Infof("send broadcast room msg key:%d topic:%s", key, d.conf.Nats.Topic)
	d.sendMessage(m)
	return nil
}

func (d *Dao) PushMsg(c context.Context, op int32, server string, keys []string, msg []byte) (err error) {
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
		Topic: d.conf.Nats.Topic,
		Value: bytes,
	}
	log.Infof("send message key:%s topic:%s", keys[0], d.conf.Nats.Topic)
	d.sendMessage(m)
	return nil
}

// SendMessage 将msg发布到NATS
func (d *Dao) sendMessage(msg *protocol.ProducerMessage) {
	key := protocol.PackagePublishKey(msg.Topic, msg.Key)
	log.Infof("logic send nat one message key:%s value:%s", key, string(msg.Value))
	if err := d.Nats.Publish(key, msg.Value); err != nil {
		log.Errorf("[sendMessage] d.Nats.Publish key:%s", key)
	}
	return
}
