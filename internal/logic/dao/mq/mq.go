/**
* @Time: 2021/2/10 下午4:27
* @Author: miku
* @File: mq
* @Version: 1.0.0
* @Description:
 */

package mq

import (
	"context"
	"errors"
	"fmt"
	"walle/internal/logic/conf"
)

type MQ interface {
	BroadcastMsg(c context.Context, op int32, msg []byte) (err error)
	BroadcastRoomMsg(c context.Context, op int32, Type, roomID string, msg []byte) (err error)
	PushMsg(c context.Context, op int32, server string, keys []string, msg []byte) (err error)
	Close()
}

func New(conf *conf.MQ) (mq MQ, err error) {
	switch conf.Mode {
	case "kafka":
		mq, err = newKafka(conf)
	case "nats":
		mq, err = newNATS(conf)
	default:
		return nil, errors.New(fmt.Sprintf("not support MQ:%s", conf.Mode))
	}
	return
}
