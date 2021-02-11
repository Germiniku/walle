/**
* @Time: 2021/2/11 下午1:16
* @Author: miku
* @File: nats
* @Version: 1.0.0
* @Description:
 */

package mq

import (
	cluster "github.com/bsm/sarama-cluster"
	log "github.com/golang/glog"
	"github.com/nats-io/nats.go"
	"time"
	"walle/internal/pipe/conf"
)

type Nats struct {
	conn *nats.Conn
	conf *conf.MQ
}

func NewNats(conf *conf.MQ) (MQ, error) {
	nat := new(Nats)
	nat.conf = conf
	conn, err := nats.Connect(conf.Brokers[0],
		nats.Timeout(time.Duration(conf.ConnTimeout)),
		nats.PingInterval(time.Duration(conf.PingInterval)*time.Second),
	)
	if err != nil {
		return nil, err
	}
	nat.conn = conn
	return nat, nil
}

func (n *Nats) Subcribe(subject string) (*cluster.Consumer, chan *nats.Msg) {
	channel := make(chan *nats.Msg, 1024)
	if _, err := n.conn.ChanSubscribe(subject, channel); err != nil {
		log.Errorf("[RecvMessage] d.Nats.ChanSubscribe failed,err:%v", err)
		return nil, nil
	}
	return nil, channel
}
