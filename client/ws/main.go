package main

import (
	"flag"
	log "github.com/golang/glog"
	"github.com/google/uuid"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"
	client2 "walle/client/ws/client"
)

var (
	addr      = flag.String("addr", "localhost:3120", "http service address")
	clientNum = flag.Int("clients", 3, "websocket client num")
)

func main() {
	flag.Parse()
	log.Infof("client start worker")
	rand.Seed(time.Now().UnixNano())

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	clients := make([]*client2.Client, *clientNum)
	for i := 1; i < *clientNum; i++ {
		randomStr := uuid.New().String()
		client := client2.NewClient(int64(i), randomStr[:6], *addr)
		clients = append(clients, client)
		client.Authentification()
		go client.KeepHeartbeat()
		go client.ReadMessage()
		client.ChangeRoom("dsb")
	}
	for {
		s := <-c
		log.Infof("walle-logic get a signal %s", s.String())
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			for _, client := range clients {
				client.Close()
				log.Infof("client key:%s mid:%d has exit", client.Key, client.Mid)
			}
			log.Infof("websocket clients all exit")
			log.Flush()
			return
		case syscall.SIGHUP:
		default:
			return
		}
	}
}
