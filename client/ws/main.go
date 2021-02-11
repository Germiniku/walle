package main

import (
	"flag"
	"fmt"
	log "github.com/golang/glog"
	"github.com/google/uuid"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
	client2 "walle/client/ws/client"
)

var (
	addr      = flag.String("addr", "localhost:3120", "http service address")
	clientNum = flag.Int("clients", 100, "websocket client num")
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
		time.Sleep(time.Millisecond * 10)
		client.ChangeRoom(client.Key)
		go test(client.Key, 10)
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

func test(roomID string, count int) {
	for i := 0; i < count; i++ {
		go pushBroadcastMsg(roomID)
		time.Sleep(time.Millisecond * 10)
	}
}

func pushBroadcastMsg(roomID string) {
	url := fmt.Sprintf("http://localhost:3000/api/push/all?Op=9")
	client := &http.Client{}
	payload := strings.NewReader("有人说你就是傻逼")
	request, err := http.NewRequest("POST", url, payload)
	if err != nil {
		log.Errorf("http.NewRequest(POST,%s,%v)", url, payload)
		return
	}
	_, err = client.Do(request)
	if err != nil {
		log.Errorf("client.Do error:%v", err)
	}
}
