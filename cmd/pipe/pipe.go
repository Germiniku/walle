/**
* @Time: 2021/1/28 下午1:29
* @Author: miku
* @File: pipe
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
	"walle/internal/pipe"
	"walle/internal/pipe/conf"
)

func main(){
	flag.Parse()
	conf.Init()
	srv := pipe.New(conf.Conf)
	srv.Run()
	log.Infof("walle-pipe start work")
	c := make(chan os.Signal,1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		log.Infof("walle-pipe get a signal %s", s.String())
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			srv.Close()
			log.Infof("walle-pipe exit")
			log.Flush()
			return
		case syscall.SIGHUP:
		default:
			return
		}
	}
}