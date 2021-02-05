/**
* @Time: 2021/1/31 上午9:09
* @Author: miku
* @File: conf
* @Version: 1.0.0
* @Description:
 */

package conf

import (
	"flag"
	"github.com/BurntSushi/toml"
	log "github.com/golang/glog"
	"time"
	xtime "walle/pkg/time"
)

type Config struct {
	TCP *TCP
	Debug bool
	Protocol *Protocol
	Discovery *Discovery
	Bucket *Bucket
	RPCServer *RPCServer
	Websocket *Websocket
}

type Websocket struct {
	Bind string
}

var (
	confPath string
	Conf *Config
)

func init(){
	flag.StringVar(&confPath,"conf","comet.toml","logic config path")
}

func Init(){
	Conf = Default()
	if _, err := toml.DecodeFile(confPath, Conf);err != nil{
		log.Errorf("comet config file decode failed,err:%v",err)
	}
	return
}

func Default()*Config{
	return &Config{
		Debug: true,
		TCP: &TCP{
			Bind:         []string{":3102"},
			Sndbuf:       4096,
			Rcvbuf:       4096,
			KeepAlive:    false,
			Reader:       32,
			ReadBuf:      1024,
			ReadBufSize:  8192,
			Writer:       32,
			WriteBuf:     1024,
			WriteBufSize: 8192,
		},
		Discovery: &Discovery{
			Server:      "comet",
			Endpoints:   []string{"http://127.0.0.1:2379"},
			DIalTimeout: 5,
			Env:         "env",
			AppId:       "127.0.0.1",
		},
		Protocol: &Protocol{
			Timer:            32,
			TimerSize:        2048,
			SvrProto:         10,
			CliProto:         5,
			HandshakeTimeout: xtime.Duration(time.Second * time.Duration(8)),
		},
		Bucket: &Bucket{
			Size:          32,
			Channel:       1024,
			Room:          1024,
			RoutineAmount: 32,
			RoutineSize:   1024,
		},
		RPCServer: &RPCServer{Addr: ":3109"},
		Websocket: &Websocket{Bind: ":3120"},

	}
}


type RPCServer struct {
	Addr string
}

// Bucket is bucket config.
type Bucket struct {
	Size          int
	Channel       int
	Room          int
	RoutineAmount uint64
	RoutineSize   int
}

type Discovery struct {
	Server string
	Endpoints []string
	DIalTimeout int
	Env string
	AppId string
}

// Protocol is protocol config.
type Protocol struct {
	Timer            int
	TimerSize        int
	SvrProto         int
	CliProto         int
	HandshakeTimeout xtime.Duration
}

// TCP is tcp config.
type TCP struct {
	Bind         []string
	Sndbuf       int
	Rcvbuf       int
	KeepAlive    bool
	Reader       int
	ReadBuf      int
	ReadBufSize  int
	Writer       int
	WriteBuf     int
	WriteBufSize int
}