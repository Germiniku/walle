/**
* @Time: 2021/1/28 下午1:33
* @Author: miku
* @File: config
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
	flag.StringVar(&confPath, "conf", "pipe.toml", "logic config path")
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
	MQ        *MQ
	Comet     *Comet
	Discovery *Discovery
	Room      *Room
}

type Room struct {
	Batch  int
	Signal xtime.Duration
	Idle   xtime.Duration
}

type Discovery struct {
	Server      string
	Endpoints   []string
	DIalTimeout int
	Env         string
	AppId       string
}

type MQ struct {
	Mode         string
	Topic        string
	Group        string
	Brokers      []string
	ConnTimeout  xtime.Duration
	PingInterval xtime.Duration
}

type Comet struct {
	RoutineChan int
	RoutineSize int
}
