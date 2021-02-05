/**
* @Time: 2021/2/1 下午4:53
* @Author: miku
* @File: websocket
* @Version: 1.0.0
* @Description:
 */

package ws

import (
	"context"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/gogo/protobuf/proto"
	log "github.com/golang/glog"
	"github.com/gorilla/websocket"
	"io"
	"net/http"
	"time"
	"walle/api/protocol"
	pb "walle/api/protocol"
	"walle/internal/comet"
	"walle/internal/comet/conf"
)

var (
	upgrader = websocket.Upgrader{}
)

// Server websocket server
type Server struct {
	conf   *conf.Config
	engine *gin.Engine
	comet  *comet.Server
}

// New 返回websocket server
func New(c *conf.Config, comet *comet.Server) *Server {
	server := new(Server)
	server.conf = c
	server.comet = comet
	engine := gin.Default()
	server.engine = engine
	server.initRouter()
	go server.Run()
	return server
}

func (s *Server) Run() {
	s.engine.Run(s.conf.Websocket.Bind)
}

// initRouter 初始化路由
func (s *Server) initRouter() {
	api := s.engine.Group("/api")
	api.GET("/sub", s.sub)
}

// sub 订阅websocket
func (s *Server) sub(c *gin.Context) {
	var (
		conn      *websocket.Conn
		b         *comet.Bucket
		mid       int64
		key       string
		rid       string
		accepts   []int32
		heartbeat time.Duration
		err       error
	)
	defer conn.Close()
	// 1.升级websocket连接
	if conn, err = upgrader.Upgrade(c.Writer, c.Request, nil); err != nil {
		if conn != nil {
			conn.Close()
		}
		log.Errorf("addr:%s upgrade websocket error(%v)", c.Request.RemoteAddr, err)
		c.JSON(http.StatusOK, nil)
		return
	}
	cookie := c.GetHeader("cookie")
	// . 获得连接所需身份信息
	if mid, key, rid, accepts, heartbeat, err = s.authWebsocket(c, conn, cookie); err != nil {
		conn.Close()
		if err != io.EOF && err != websocket.ErrCloseSent {
			log.Errorf("key: %s remoteIP %s ws failed,error:%v", key, conn.RemoteAddr().String(), err)
		}
		return
	}
	if conf.Conf.Debug {
		log.Infof("mid:%d key:%s rid:%s", mid, key, rid)
	}
	// 创建连接
	channel := comet.NewChannel(s.conf.Protocol.SvrProto, conn)
	channel.Heartbeat = heartbeat
	channel.Mid, channel.Key = mid, key
	channel.Watch(accepts...)
	// 将连接放进bucket中管理
	b = s.comet.Bucket(channel.Key)
	b.Put(rid, channel)
	if conf.Conf.Debug {
		log.Infof("websocket connected key:%s mid:%d rid:%s", channel.Key, channel.Mid, rid)
	}
	// 处理派发给websocket的信息
	go s.dispatchWebsocket(conn, channel)
	// 处理心跳事件
	//serverHearbeat := s.comet.RandServerHearbeat()
	channel.SetServerHeartboat(time.Minute)
	timer := time.NewTimer(time.Second * 40)
	channel.Timer = timer
	channel.SetHeartBeatEvent()
	go s.heartbeatCheck(channel.Timer, func() {
		channel.HeartBeatFailed()
	})
	for {
		if err = s.watchWebsocketOperate(c, channel, conn, b); err != nil {
			log.Errorf("watchWebsocketOperate error:%v", err)
			break
		}
		// 	// 通知连接已经准备好
		channel.Signal()
	}
	channel.CloseConn()
	channel.Close()
	return
}

// watchWebsocketOperate 监听websocket操作
func (s *Server) watchWebsocketOperate(c context.Context, channel *comet.Channel, conn *websocket.Conn, b *comet.Bucket) (err error) {
	var (
		data []byte
		p    protocol.Proto
	)
	if _, data, err = conn.ReadMessage(); err != nil {
		return
	}
	if err = proto.Unmarshal(data, &p); err != nil {
		return
	}
	// 处理websocket client端传过来的操作
	if err = s.comet.Operate(c, channel, &p, b); err != nil {
		return
	}
	return
}

func (s *Server) heartbeatCheck(timer *time.Timer, fn func()) {
	for {
		select {
		case <-timer.C:
			fn()
		default:
			time.Sleep(time.Second)
		}
	}
}

func (s *Server) dispatchWebsocket(conn *websocket.Conn, ch *comet.Channel) {
	var (
		online int32
		finish bool
		data   []byte
		err    error
	)
	if conf.Conf.Debug {
		log.Infof("key:%s start dispatch websocket goroutine", ch.Key)
	}
	for {
		p := ch.Ready()
		if conf.Conf.Debug {
			log.Infof("key:%s dispatch msg:%s", ch.Key, string(p.Body))
		}
		switch p {
		case protocol.ProtoReady:
			if conf.Conf.Debug {
				log.Infof("key:%s ready dispatch goroutine", ch.Key)
			}
			for {
				// 获取logic发送过来的心跳回复
				p = ch.Ready()
				if err = proto.Unmarshal(data, p); err != nil {
					log.Errorf("proto.Unmarshal error:%v", err)
					continue
				}
				data, err = proto.Marshal(p)
				if err != nil {
					continue
				}
				// 收到心跳回复则返回心跳信息
				if p.Op == protocol.OpHeartbeatReply {
					if ch.Room != nil {
						online = ch.Room.OnlineNum()
					}
					var heartPayload struct {
						Online int32 `json:"online"`
					}
					heartPayload.Online = online
					body, _ := json.Marshal(&heartPayload)
					p.Body = body
					if err = ch.WriteWebsocketHeart(p); err != nil {
						goto failed
					}
				} else {
					if err = ch.WriteWebsocketMessage(p); err != nil {
						goto failed
					}
				}
			}
		case protocol.ProtoFinish:
			if conf.Conf.Debug {
				log.Infof("key:%s wakeup exit dispatch goroutine", ch.Key)
			}
			finish = true
			goto failed
		default:
			if err = ch.WriteWebsocketMessage(p); err != nil {
				goto failed
			}
		}
	}
failed:
	if err != nil && err != io.EOF && err != websocket.ErrCloseSent {
		log.Errorf("key: %s dispatch ws error(%v)", ch.Key, err)
	}
	for !finish {
		finish = ch.Ready() == protocol.ProtoFinish
	}
	if conf.Conf.Debug {
		log.Infof("key: %s dispatch goroutine exit", ch.Key)
	}
}

// authWebsocket 确认身份返回连接信息
func (s *Server) authWebsocket(c context.Context, conn *websocket.Conn, cookie string) (mid int64, key, rid string, accepts []int32, heartbeat time.Duration, err error) {
	var (
		readCount   = 0
		msg         []byte
		p           pb.Proto
		messageType int
	)
	for {
		log.Infof("这里在执行第%d次", readCount)
		if readCount >= 5 {
			log.Infof("client:%s conn close", conn.RemoteAddr().String())
			conn.Close()
			return
		}
		if messageType, msg, err = conn.ReadMessage(); err != nil {
			log.Errorf("conn.ReadMessage error:%v", err)
			conn.Close()
			return
		}
		if err = json.Unmarshal(msg, &p); err != nil {
			log.Errorf("proto.Unmarshal error:%v", err)
			readCount++
		}
		if conf.Conf.Debug {
			log.Infof("read one message messageType:%d op:%d ver:%d", messageType, p.Op, p.Ver)
		}
		if p.Op == protocol.OpAuth {
			break
		}
	}
	if mid, key, rid, accepts, heartbeat, err = s.comet.Connect(c, &p, cookie); err != nil {
		log.Errorf("s.comet.Connect error:%v", err)
		return
	}
	if s.conf.Debug {
		log.Infof("mid:%d key:%s rid:%s accepts:%v heartbeat:%v", mid, key, rid, accepts, heartbeat)
	}
	p.Op = protocol.OpAuthReply
	p.Body = nil
	data, err := json.Marshal(&p)
	if err != nil {
		log.Errorf("proto.Marshal error:%v", err)
	}
	if err = conn.WriteMessage(websocket.TextMessage, data); err != nil {
		log.Errorf("conn.WriteMessage(%d,%s)", websocket.TextMessage, string(data))
	}
	return
}
