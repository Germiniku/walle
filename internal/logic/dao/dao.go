/**
* @Time: 2021/1/28 上午10:57
* @Author: miku
* @File: dao
* @Version: 1.0.0
* @Description:
 */

package dao

import (
	"context"
	"fmt"
	log "github.com/golang/glog"
	"github.com/gomodule/redigo/redis"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
	"walle/internal/logic/conf"
	"walle/internal/logic/dao/mq"
)

type Dao struct {
	conf        *conf.Config
	Redis       *redis.Pool
	MQ          mq.MQ
	redisExpire int32
	DB          *mongo.Database
}

func New(c *conf.Config) *Dao {
	mq, err := mq.New(c.MQ)
	if err != nil {
		panic(err)
	}
	dao := &Dao{
		conf:        c,
		Redis:       NewRedis(c.Redis),
		MQ:          mq,
		redisExpire: int32(time.Duration(c.Redis.Expire) / time.Second),
		//DB:          NewDB(c.DB),
	}
	if err := dao.Ping(); err != nil {
		panic(err)
	}
	return dao
}

func NewDB(conf *conf.DB) *mongo.Database {
	// 设置客户端连接配置
	url := fmt.Sprintf("mongodb://%s:%d", conf.Addr, conf.Port)
	clientOptions := options.Client().ApplyURI(url)
	clientOptions.SetConnectTimeout(time.Second * time.Duration(conf.ConnectTimeout))
	clientOptions.SetMinPoolSize(conf.MinPoolSize)
	clientOptions.SetMaxPoolSize(conf.MaxPoolSize)
	clientOptions.SetHeartbeatInterval(time.Second * time.Duration(conf.HeartbeatInterval))
	// 连接到MongoDb
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		panic(err)
	}
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Errorf("Ping Mongo Error:%v", err)
	}
	log.Infof("logic connect Mongo Success")
	return client.Database(conf.DB)
}

func NewRedis(conf *conf.Redis) *redis.Pool {
	fmt.Println("DialConnectTimeout", conf.DialConnectTimeout)
	pool := &redis.Pool{
		Dial: func() (redis.Conn, error) {
			conn, err := redis.Dial("tcp", conf.Addr,
				redis.DialPassword(conf.DialPassword),
				redis.DialConnectTimeout(time.Duration(conf.DialConnectTimeout)),
				redis.DialWriteTimeout(time.Duration(conf.DialWriteTimeout)),
				redis.DialReadTimeout(time.Duration(conf.DialReadTimeout)),
			)
			return conn, err
		},
		MaxIdle:     conf.MaxIdle,
		MaxActive:   conf.MaxActive,
		IdleTimeout: time.Duration(conf.IdleTimeout),
	}
	conn := pool.Get()
	if result, err := redis.String(conn.Do("PING")); err != nil || result != "PONG" {
		log.Errorf("connect redis failed,error:%v", err)
		panic(err)
	}
	log.Infof("logic connect Redis Success")
	return pool
}

func (d *Dao) Ping() (err error) {
	if err = d.pingRedis(); err != nil {
		return
	}
	return nil
}

func (d *Dao) Close() {
	d.MQ.Close()
	d.Redis.Close()
}
