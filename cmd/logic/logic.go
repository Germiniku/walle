/**
* @Time: 2021/1/28 上午10:48
* @Author: miku
* @File: logic
* @Version: 1.0.0
* @Description:
 */

package main

import (
	"flag"
	log "github.com/golang/glog"
	"os"
	"os/signal"
	"syscall"
	"walle/internal/logic"
	"walle/internal/logic/conf"
	"walle/internal/logic/http"
	"walle/internal/logic/rpcx"
)

func main() {
	flag.Parse()
	conf.Init()
	srv := logic.New(conf.Conf)
	httpServer := http.New(srv, conf.Conf.HTTPServer)
	rpcServer := rpcx.New(conf.Conf, srv)
	rpcServer.Serve()
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		log.Infof("walle-logic get a signal %s", s.String())
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			srv.Close()
			httpServer.Close()
			rpcServer.Close()
			log.Infof("goim-logic exit")
			log.Flush()
			return
		case syscall.SIGHUP:
		default:
			return
		}
	}
}
