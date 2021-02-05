/**
* @Time: 2021/1/28 上午10:57
* @Author: miku
* @File: dao
* @Version: 1.0.0
* @Description:
 */

package dao

import (
	"fmt"
	"github.com/Germiniku/discovery"
	"github.com/nats-io/nats.go"
	"time"
	"walle/internal/pipe/conf"
)

type Dao struct {
	conf      *conf.Config
	Nats      *nats.Conn
	sub       *nats.Subscription
	Discovery discovery.Server
}

func New(c *conf.Config) *Dao {
	cfg := discovery.NewConfig(c.Discovery.Endpoints, c.Discovery.DIalTimeout, c.Discovery.Env, c.Discovery.AppId, "")
	discoverySrv, err := discovery.New("etcd", cfg)
	if err != nil {
		panic(err)
	}
	dao := &Dao{
		conf:      c,
		Nats:      NewNats(c.Nats),
		Discovery: discoverySrv,
	}
	return dao
}

func NewNats(conf *conf.Nats) *nats.Conn {
	conn, err := nats.Connect(conf.Addrs,
		nats.Timeout(time.Duration(conf.ConnTimeout)),
		nats.PingInterval(time.Duration(conf.PingInterval)*time.Second),
	)
	if err != nil {
		fmt.Println("nats addr:", conf.Addrs)
		panic(err)
	}
	return conn
}

func (d *Dao) Ping() error {
	return nil
}

func (d *Dao) Close() error {
	d.Nats.Close()
	return nil
}
