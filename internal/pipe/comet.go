/**
* @Time: 2021/1/28 下午3:27
* @Author: miku
* @File: comet
* @Version: 1.0.0
* @Description:
 */

package pipe

import (
	"github.com/smallnest/rpcx/client"
	"sync/atomic"
	"time"
	"walle/api/comet"
	"walle/internal/pipe/conf"
)

type Comet struct {
	serverId      string
	ip            string
	pushChan      []chan *comet.PushMsgReq
	roomChan      []chan *comet.BroadcastRoomReq
	broadcastChan chan *comet.BroadcastReq
	pushChanNum   uint64
	roomChanNum   uint64
	routineSize   uint64
	rpcClient     client.XClient
}

func NewComet(c *conf.Comet, serverId, ip string) (*Comet, error) {
	cmt := &Comet{
		serverId:      serverId,
		pushChan:      make([]chan *comet.PushMsgReq, c.RoutineChan),
		broadcastChan: make(chan *comet.BroadcastReq, c.RoutineSize),
		roomChan:      make([]chan *comet.BroadcastRoomReq, c.RoutineSize),
		routineSize:   uint64(c.RoutineSize),
		ip:            ip,
		rpcClient:     newRpcClient(ip),
	}
	for i := 0; i < c.RoutineSize; i++ {
		cmt.pushChan[i] = make(chan *comet.PushMsgReq, c.RoutineChan)
		cmt.roomChan[i] = make(chan *comet.BroadcastRoomReq, c.RoutineChan)
		go cmt.process(cmt.pushChan[i], cmt.roomChan[i], cmt.broadcastChan)
	}
	return cmt, nil
}

func (c *Comet) process(pushChan chan *comet.PushMsgReq, roomChan chan *comet.BroadcastRoomReq, broadcastChan chan *comet.BroadcastReq) {
	for {
		select {
		case pushArg := <-pushChan:
			c.callPushMsg(pushArg)
		case roomArg := <-roomChan:
			c.callBroadcastRoom(roomArg)
		case pushArg := <-broadcastChan:
			c.callBroadcastMsg(pushArg)
		default:
			time.Sleep(time.Second)
		}
	}
}

func (c *Comet) push(req *comet.PushMsgReq) {
	idx := atomic.AddUint64(&c.pushChanNum, 1) % c.routineSize
	c.pushChan[idx] <- req
	return
}

func (c *Comet) BroadcastRoom(arg *comet.BroadcastRoomReq) (err error) {
	idx := atomic.AddUint64(&c.roomChanNum, 1) % c.routineSize
	c.roomChan[idx] <- arg
	return
}

func (c *Comet) Broadcast(req *comet.BroadcastReq) {
	c.broadcastChan <- req
	return
}

func (c *Comet) Close() {

}
