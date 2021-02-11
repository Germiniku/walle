/**
* @Time: 2021/1/28 上午10:57
* @Author: miku
* @File: dao
* @Version: 1.0.0
* @Description:
 */

package dao

import (
	"github.com/Germiniku/discovery"
	log "github.com/golang/glog"
	"github.com/nats-io/nats.go"
	"walle/internal/pipe/conf"
	"walle/internal/pipe/dao/mq"
)

type Dao struct {
	conf      *conf.Config
	MQ        mq.MQ
	sub       *nats.Subscription
	Discovery discovery.Server
}

func New(c *conf.Config) *Dao {

	cfg := discovery.NewConfig(c.Discovery.Endpoints, c.Discovery.DIalTimeout, c.Discovery.Env, c.Discovery.AppId, "")
	discoverySrv, err := discovery.New("etcd", cfg)
	if err != nil {
		panic(err)
	}
	mq, err := mq.New(c.MQ)
	if err != nil {
		log.Errorf("MQ Initial error:%v", err)
	}
	dao := &Dao{
		conf:      c,
		MQ:        mq,
		Discovery: discoverySrv,
	}
	return dao
}

func (d *Dao) Ping() error {
	return nil
}

func (d *Dao) Close() error {
	return nil
}
