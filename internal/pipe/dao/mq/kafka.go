/**
* @Time: 2021/2/11 下午1:16
* @Author: miku
* @File: kafka
* @Version: 1.0.0
* @Description:
 */

package mq

import (
	cluster "github.com/bsm/sarama-cluster"
	"github.com/nats-io/nats.go"
	"walle/internal/pipe/conf"
)

type Kafka struct {
	consumer *cluster.Consumer
	conf     *conf.MQ
}

func NewKafka(c *conf.MQ) (MQ, error) {
	k := new(Kafka)
	k.conf = c
	config := cluster.NewConfig()
	config.Consumer.Return.Errors = true
	config.Group.Return.Notifications = true
	consumer, err := cluster.NewConsumer(c.Brokers, c.Group, []string{c.Group}, config)
	if err != nil {
		return nil, err
	}
	k.consumer = consumer
	return k, nil
}

func (k *Kafka) Subcribe(subject string) (*cluster.Consumer, chan *nats.Msg) {
	return k.consumer, nil
}
