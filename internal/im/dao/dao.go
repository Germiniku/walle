/**
* @Time: 2021/1/30 下午2:16
* @Author: miku
* @File: dao
* @Version: 1.0.0
* @Description:
 */

package dao

import (
	"fmt"
	"context"
	log "github.com/golang/glog"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
	"walle/internal/im/conf"
)

type Dao struct {
	conf *conf.Config
	DB *Mongo
}

type Mongo struct {
	DB *mongo.Database
	User *mongo.Collection
}

func New(c *conf.Config)*Dao{
	dao := &Dao{
		conf: c,
		DB: NewDB(c.DB),
	}
	return dao
}

func NewDB(conf *conf.DB)*Mongo{
	var db *Mongo
	// 设置客户端连接配置
	url := fmt.Sprintf("mongodb://%s:%d",conf.Addr,conf.Port)
	clientOptions := options.Client().ApplyURI(url)
	clientOptions.SetConnectTimeout(time.Second * time.Duration(conf.ConnectTimeout))
	clientOptions.SetMinPoolSize(conf.MinPoolSize)
	clientOptions.SetMaxPoolSize(conf.MaxPoolSize)
	clientOptions.SetHeartbeatInterval(time.Second * time.Duration(conf.HeartbeatInterval))
	// 连接到MongoDb
	clinet,err := mongo.Connect(context.TODO(),clientOptions)
	if err != nil{
		panic(err)
	}
	err = clinet.Ping(context.TODO(), nil)
	if err != nil {
		log.Errorf("Ping Mongo Error:%v",err)
	}
	db.DB = clinet.Database(conf.DB)
	db.User = db.DB.Collection("user")
	log.Infof("logic connect Mongo Success")
	return db
}