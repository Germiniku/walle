/**
* @Time: 2021/1/28 上午11:22
* @Author: miku
* @File: nats
* @Version: 1.0.0
* @Description:
 */

package dao

import (
	log "github.com/golang/glog"
	"github.com/nats-io/nats.go"
)

func (d *Dao) RecvMessage(subject string, channel chan *nats.Msg) {
	if _, err := d.Nats.ChanSubscribe(subject, channel); err != nil {
		log.Errorf("[RecvMessage] d.Nats.ChanSubscribe failed,err:%v", err)
	}
	return
}
