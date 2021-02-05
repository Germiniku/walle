/**
* @Time: 2021/1/29 下午1:54
* @Author: miku
* @File: rpcx
* @Version: 1.0.0
* @Description:
 */

package pipe

import (
	"context"
	log "github.com/golang/glog"
	"github.com/smallnest/rpcx/client"
	"walle/api/comet"
)

func newRpcClient(addr string) client.XClient {
	d, err := client.NewPeer2PeerDiscovery(addr, "")
	if err != nil {
		log.Errorf("client.NewPeer2PeerDiscovery(%s) error:%v", addr, err)
		return nil
	}
	rpcClient := client.NewXClient("comet", client.Failover, client.RoundRobin, d, client.DefaultOption)
	return rpcClient
}

// callPushMsg 私聊
func (c *Comet) callPushMsg(msg *comet.PushMsgReq) {
	var reply comet.PushMsgReply
	log.Infof("调用了comet PushMsg")
	if err := c.rpcClient.Call(context.Background(), "PushMsg", msg, &reply); err != nil {
		log.Errorf("call PushMsg error:%v", err)
	}
}

// callBroadcastMsg 全频道广播
func (c *Comet) callBroadcastMsg(msg *comet.BroadcastReq) {
	var reply comet.BroadcastReply
	if err := c.rpcClient.Call(context.Background(), "Broadcast", msg, &reply); err != nil {
		log.Error("call Broadcast error:%v", err)
	}
}

// callBroadcastRoom 房间广播
func (c *Comet) callBroadcastRoom(msg *comet.BroadcastRoomReq) {
	var reply comet.BroadcastRoomReq
	if err := c.rpcClient.Call(context.Background(), "BroadcastRoom", msg, &reply); err != nil {
		log.Errorf("call BroadcastRoom error:%v", err)
	}
}
