/**
* @Time: 2021/2/11 下午1:35
* @Author: miku
* @File: mq
* @Version: 1.0.0
* @Description:
 */

package mq

import (
	cluster "github.com/bsm/sarama-cluster"
	"github.com/nats-io/nats.go"
	"walle/internal/pipe/conf"
)

type MQ interface {
	Subcribe(subject string) (*cluster.Consumer, chan *nats.Msg)
}

func New(conf *conf.MQ) (mq MQ, err error) {
	switch conf.Mode {
	case "kafka":
		mq, err = NewKafka(conf)
	case "nats":
		mq, err = NewNats(conf)
	default:
		mq, err = NewKafka(conf)
	}
	return
}
