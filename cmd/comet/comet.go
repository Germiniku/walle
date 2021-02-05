/**
* @Time: 2021/2/1 上午10:22
* @Author: miku
* @File: comet.go
* @Version: 1.0.0
* @Description:
 */

package main

import (
	"flag"
	log "github.com/golang/glog"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"
	"walle/internal/comet"
	"walle/internal/comet/conf"
	"walle/internal/comet/rpcx"
	"walle/internal/comet/ws"
)

func main() {
	flag.Parse()
	rand.Seed(time.Now().UTC().UnixNano())
	conf.Init()
	srv := comet.New(conf.Conf)
	rpcServer := rpcx.New(conf.Conf, srv)
	ws.New(conf.Conf, srv)
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		log.Infof("walle-comet get a signal %s", s.String())
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			srv.Close()
			rpcServer.Close()
			log.Infof("goim-comet exit")
			log.Flush()
			return
		case syscall.SIGHUP:
		default:
			return
		}
	}
}
