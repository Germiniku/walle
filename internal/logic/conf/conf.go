/**
* @Time: 2021/1/28 上午10:43
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
	xtime "walle/pkg/time"
)

var (
	confPath string
	Conf     *Config
)

func init() {
	flag.StringVar(&confPath, "conf", "logic.toml", "logic config path")
}

func Init() {
	Conf = Default()
	if _, err := toml.DecodeFile(confPath, Conf); err != nil {
		log.Errorf("logic config file decode failed,err:%v", err)
	}
	return
}

func Default() *Config {
	return &Config{}
}

type Config struct {
	HTTPServer *HTTPServer
	MQ         *MQ
	Redis      *Redis
	Discovery  *Discovery
	RpcServer  *RpcServer
	DB         *DB
	Server     *Server
}

type Server struct {
	ChatRecordRoutine int
	ChatRecordBatch   int32
}

type DB struct {
	Addr              string
	Port              int
	ConnectTimeout    int
	MinPoolSize       uint64
	MaxPoolSize       uint64
	HeartbeatInterval int
	DB                string
}

type RpcServer struct {
	Addr string
}

type Discovery struct {
	Server        string
	Endpoints     []string
	DIalTimeout   int
	Env           string
	AppId         string
	ServerAddress string
}

type HTTPServer struct {
	Addr string
}

type Redis struct {
	Addr               string
	DialPassword       string
	MaxIdle            int
	MaxActive          int
	IdleTimeout        int
	Expire             xtime.Duration
	DialConnectTimeout xtime.Duration
	DialWriteTimeout   xtime.Duration
	DialReadTimeout    xtime.Duration
}

type MQ struct {
	Addrs        []string
	Topic        string
	Mode         string
	ConnTimeout  xtime.Duration
	PingInterval xtime.Duration
}
