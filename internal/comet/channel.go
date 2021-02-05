/**
* @Time: 2021/1/31 上午9:25
* @Author: miku
* @File: channel
* @Version: 1.0.0
* @Description:
 */

package comet

import (
	"context"
	"encoding/json"
	"github.com/Terry-Mao/goim/api/protocol"
	log "github.com/golang/glog"
	"github.com/gorilla/websocket"
	"sync"
	"time"
	pb "walle/api/protocol"
	"walle/internal/comet/conf"
	"walle/pkg/bufio"
)

// Channel 客户端连接
type Channel struct {
	Room           *Room
	Conn           *websocket.Conn
	signal         chan *pb.Proto
	Mid            int64  // 用户id
	IP             string // ip 地址
	Key            string
	watchOps       map[int32]struct{}
	lastHeartBeat  time.Time
	Heartbeat      time.Duration
	serverHearbeat time.Duration
	Writer         bufio.Writer
	Reader         bufio.Reader
	Timer          *time.Timer
	Prev           *Channel
	Next           *Channel
	mutex          sync.RWMutex
}

func NewChannel(srvProto int, conn *websocket.Conn) *Channel {
	c := new(Channel)
	c.Conn = conn
	c.IP = conn.RemoteAddr().String()
	c.signal = make(chan *pb.Proto, srvProto)
	c.watchOps = make(map[int32]struct{})
	c.lastHeartBeat = time.Now()
	return c
}

// SetServerHeartboat 设置服务心跳检查时间
func (c *Channel) SetServerHeartboat(heartbeat time.Duration) {
	c.serverHearbeat = heartbeat
}

// SetHeartBeatEvent 设置心跳检查事件
func (c *Channel) SetHeartBeatEvent() {
	c.Conn.SetPingHandler(c.pingHandler)
}

// HeartBeatFailed 心跳断开
func (c *Channel) HeartBeatFailed() {
	if conf.Conf.Debug {
		log.Infof("key:%s heart beat failed", c.Key)
	}
	c.Conn.SetWriteDeadline(time.Now().Add(time.Millisecond * 100))
}

func (c *Channel) CloseConn() {
	if c.Conn != nil {
		c.Conn.Close()
	}
	return
}

// pingHandler 处理客户端ping心跳
func (c *Channel) pingHandler(data string) (err error) {
	var heartbeatMsg pb.Proto
	if err = json.Unmarshal([]byte(data), &heartbeatMsg); err != nil {
		log.Errorf("json.Unmarshal(%s) error:%v", data, err)
		return
	}
	// 判断是否是心跳请求
	if heartbeatMsg.Op == protocol.OpHeartbeat {
		// 重置定时器
		c.Timer.Reset(time.Second * 40)
		heartbeatMsg.Op = protocol.OpHeartbeatReply
		heartbeatMsg.Body = nil
		if err = c.WriteWebsocketHeart(&heartbeatMsg); err != nil {
			log.Errorf("channel.WriteWebsocketHeart error:%v", err)
		}
		log.Infof("lastHeartBeat:%v serverHeartbeat:%v", c.lastHeartBeat, c.serverHearbeat)
		if now := time.Now(); now.Sub(c.lastHeartBeat) > c.serverHearbeat {
			log.Infof("发向logic 心跳保活")
			if err = server.Heartbeat(context.Background(), c.Mid, c.Key); err == nil {
				c.lastHeartBeat = now
			}
		}
		if conf.Conf.Debug {
			log.Infof("websocket heartbeat receive key:%s mid:%d", c.Key, c.Mid)
		}
	}
	return
}

func (c *Channel) WriteWebsocketHeart(p *pb.Proto) (err error) {
	err = c.Write(websocket.TextMessage, p)
	return
}

func (c *Channel) WriteWebsocketMessage(p *pb.Proto) (err error) {
	log.Infof("发送了文本消息呢")
	err = c.Write(websocket.TextMessage, p)
	return
}

func (c *Channel) Write(messageType int, p *pb.Proto) (err error) {
	data, err := json.Marshal(p)
	if err != nil {
		log.Errorf("channel:%s write message marshal error:%v", c.Key, err)
	}
	c.Conn.WriteMessage(messageType, data)
	return
}

func (s *Channel) Signal() {
	s.signal <- pb.ProtoReady
}

// Watch 监听相关操作符
func (c *Channel) Watch(accepts ...int32) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	for _, op := range accepts {
		c.watchOps[op] = struct{}{}
	}
	return
}

// UnWatch 取消监听操作符
func (c *Channel) UnWatch(accepts ...int32) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	for _, op := range accepts {
		delete(c.watchOps, op)
	}
	return
}

// NeedPush 判断操作符是否监听
func (c *Channel) NeedPush(op int32) bool {
	c.mutex.RLock()
	if _, ok := c.watchOps[op]; ok {
		c.mutex.RUnlock()
		return true
	}
	c.mutex.RUnlock()
	return false
}

// Push 将消息推送给连接
func (c *Channel) Push(proto *pb.Proto) {
	log.Infof("push proto:%v", proto)
	c.signal <- proto
	return
}

// Ready 获取连接推送消息
func (c *Channel) Ready() *pb.Proto {
	return <-c.signal
}

// Close 关闭连接
func (c *Channel) Close() {
	c.signal <- pb.ProtoFinish
}
