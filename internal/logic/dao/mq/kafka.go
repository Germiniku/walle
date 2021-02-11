/**
* @Time: 2021/2/10 下午4:27
* @Author: miku
* @File: kafka
* @Version: 1.0.0
* @Description:
 */

package mq

import (
	"context"
	"github.com/Shopify/sarama"
	log "github.com/golang/glog"
	"github.com/golang/protobuf/proto"
	"strconv"
	pb "walle/api/logic"
	"walle/internal/logic/conf"
)

type Kafka struct {
	client sarama.SyncProducer
	conf   *conf.MQ
}

func newKafka(conf *conf.MQ) (kafka *Kafka, err error) {
	var client sarama.SyncProducer
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForLocal //
	config.Producer.Partitioner = sarama.NewHashPartitioner
	config.Producer.Compression = sarama.CompressionSnappy
	config.Producer.Retry.Max = 10 // retry up to 10 times to produce messasge
	config.Producer.Return.Successes = true
	if client, err = sarama.NewSyncProducer(conf.Addrs, config); err != nil {
		log.Errorf("create kafka client failed,err:%v", err)
		return nil, err
	}
	k := &Kafka{
		client: client,
		conf:   conf,
	}
	return k, nil
}

func (k *Kafka) Close() {
	k.client.Close()
}

func (k *Kafka) PushMsg(c context.Context, op int32, server string, keys []string, msg []byte) (err error) {
	pushMsg := &pb.PushMsg{
		Type:      pb.PushMsg_PUSH,
		Operation: op,
		Msg:       msg,
		Server:    server,
		Keys:      keys,
	}
	b, err := proto.Marshal(pushMsg)
	if err != nil {
		return
	}
	message := &sarama.ProducerMessage{
		Key:   sarama.StringEncoder(keys[0]),
		Topic: k.conf.Topic,
		Value: sarama.ByteEncoder(b),
	}
	if _, _, err = k.client.SendMessage(message); err != nil {
		log.Errorf("PushMsg.send(pushMsg:%v) error:%v", pushMsg, err)
	}
	return
}
func (k *Kafka) BroadcastRoomMsg(c context.Context, op int32, Type, roomID string, msg []byte) (err error) {
	pushMsg := &pb.PushMsg{
		Type:      pb.PushMsg_BROADCAST,
		Operation: op,
		Msg:       msg,
		Room:      roomID,
	}
	bytes, err := proto.Marshal(pushMsg)
	if err != nil {
		return
	}
	message := &sarama.ProducerMessage{
		Key:   sarama.StringEncoder(roomID),
		Topic: k.conf.Topic,
		Value: sarama.ByteEncoder(bytes),
	}
	if _, _, err = k.client.SendMessage(message); err != nil {
		log.Errorf("PushMsg.send(BroadcastMsg:%v) error:%v", pushMsg, err)
	}
	return
}

func (k *Kafka) BroadcastMsg(c context.Context, op int32, msg []byte) (err error) {
	pushMsg := &pb.PushMsg{
		Type:      pb.PushMsg_BROADCAST,
		Operation: op,
		Msg:       msg,
	}
	bytes, err := proto.Marshal(pushMsg)
	if err != nil {
		return
	}
	message := &sarama.ProducerMessage{
		Key:   sarama.StringEncoder(strconv.FormatInt(int64(op), 10)),
		Topic: k.conf.Topic,
		Value: sarama.ByteEncoder(bytes),
	}
	if _, _, err = k.client.SendMessage(message); err != nil {
		log.Errorf("PushMsg.send(BroadcastMsg:%v) error:%v", pushMsg, err)
	}
	return
}
