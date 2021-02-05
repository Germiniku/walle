/**
* @Time: 2021/1/31 上午9:09
* @Author: miku
* @File: server
* @Version: 1.0.0
* @Description:
 */

package comet

import (
	"fmt"
	log "github.com/golang/glog"
	"github.com/google/uuid"
	"github.com/rpcxio/libkv/store"
	etcd_client "github.com/rpcxio/rpcx-etcd/client"
	"github.com/smallnest/rpcx/client"
	"github.com/zhenjl/cityhash"
	"math/rand"
	"time"
	"walle/internal/comet/conf"
)

const (
	minServerHeartbeat = time.Minute * 10
	maxServerHeartbeat = time.Minute * 30
	// grpc options
	grpcInitialWindowSize     = 1 << 24
	grpcInitialConnWindowSize = 1 << 24
	grpcMaxSendMsgSize        = 1 << 24
	grpcMaxCallMsgSize        = 1 << 24
	grpcKeepAliveTime         = time.Second * 10
	grpcKeepAliveTimeout      = time.Second * 3
	grpcBackoffMaxDelay       = time.Second * 3
)

var server *Server

type Server struct {
	ServerId  string
	conf      *conf.Config
	round     *Round
	buckets   []*Bucket
	rpcClient client.XClient
	bucketIdx uint32
}

func (s *Server) Buckets() []*Bucket {
	return s.buckets
}

// New return comet server
func New(c *conf.Config) *Server {
	server = new(Server)
	options := &store.Config{}
	options.ConnectionTimeout = time.Second * 5
	log.Infof("prefix:%s", fmt.Sprintf("/%s/%s", c.Discovery.Env, c.Discovery.AppId))
	d, err := etcd_client.NewEtcdV3Discovery(
		fmt.Sprintf("/%s/%s", c.Discovery.Env, c.Discovery.AppId),
		"logic",
		c.Discovery.Endpoints,
		options,
	)
	if err != nil {
		log.Errorf("etcd_client.NewEtcdV3Discovery error:%v", err)
		return nil
	}
	server.rpcClient = client.NewXClient("logic", client.Failover, client.RoundRobin, d, client.DefaultOption)
	server.ServerId = uuid.New().String()
	server.conf = c
	server.buckets = make([]*Bucket, c.Bucket.Size)
	server.bucketIdx = uint32(c.Bucket.Size)
	for i := 0; i < c.Bucket.Size; i++ {
		server.buckets[i] = NewBucket(c.Bucket)
	}
	server.round = NewRound(c)
	return server
}

func (s *Server) Bucket(key string) *Bucket {
	idx := cityhash.CityHash32([]byte(key), uint32(len(key))) % s.bucketIdx
	if s.conf.Debug {
		log.Infof("%s hit channel bucket index: %d use cityhash", key, idx)
	}
	return s.buckets[idx]
}

// RandServerHearbeat 返回一个心跳时间随机数
func (s *Server) RandServerHearbeat() time.Duration {
	return minServerHeartbeat + time.Duration(rand.Int63n(int64(maxServerHeartbeat)))
}

// Close 关闭服务
func (s *Server) Close() {
	s.rpcClient.Close()
}
