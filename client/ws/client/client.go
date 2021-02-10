/**
* @Time: 2021/2/3 下午1:49
* @Author: miku
* @File: client
* @Version: 1.0.0
* @Description:
 */

package client

import "C"
import (
	"encoding/json"
	log "github.com/golang/glog"
	"github.com/gorilla/websocket"
	"net/url"
	"time"
	pb "walle/api/protocol"
)

type Client struct {
	Key  string
	Mid  int64
	Conn *websocket.Conn
}

type params struct {
	Mid      int64   `json:"mid"`
	Key      string  `json:"key"`
	RoomID   string  `json:"room_id"`
	Platform string  `json:"platform"`
	Accepts  []int32 `json:"accepts"`
}

func (c *Client) newAuthProtoJson() []byte {
	accepts := []int32{
		pb.OpUnsub, pb.OpChangeRoom, pb.OpRaw, pb.OpHeartbeat, pb.OpHeartbeatReply,
	}
	param := params{
		Mid:      c.Mid,
		Key:      c.Key,
		RoomID:   "",
		Platform: "ios",
		Accepts:  accepts,
	}
	body, _ := json.Marshal(&param)
	p := pb.Proto{Op: pb.OpAuth, Ver: 1, Seq: 1, Body: body}
	data, err := json.Marshal(&p)
	if err != nil {
		return nil
	}
	return data
}

func NewHeartbeatProto() []byte {
	p := pb.Proto{Op: pb.OpHeartbeat, Ver: 1, Seq: 1}
	data, err := json.Marshal(&p)
	if err != nil {
		return nil
	}
	return data
}

func NewClient(mid int64, id, addr string) *Client {
	client := new(Client)
	client.Key = id
	client.Mid = mid
	u := url.URL{Scheme: "ws", Host: addr, Path: "/api/sub"}
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Errorf("client:%s dial server error:%v", id, err)
	}
	client.Conn = c
	log.Infof("client key:%s mid:%d connected", id, mid)
	return client
}

func (c *Client) ChangeRoom(roomID string) {
	var (
		data []byte
		err  error
	)
	p := pb.Proto{Op: pb.OpChangeRoom, Ver: 1, Seq: 1, Body: []byte(roomID)}
	if data, err = json.Marshal(&p); err != nil {
		log.Errorf("proto.Marshal(%v) error:%v", p, err)
	}
	log.Infof("change room msg:%s", string(data))
	if err = c.Conn.WriteMessage(websocket.BinaryMessage, data); err != nil {
		log.Infof("write change room error:%v", err)
	}
	return
}

// Authentification client发送身份认证
func (c *Client) Authentification() {
	p := c.newAuthProtoJson()
	c.WriteTextMessage(p)
}

// KeepHeartbeat 保持心跳连接
func (c *Client) KeepHeartbeat() {
	c.Conn.SetPongHandler(func(appData string) error {
		log.Infof("接收到了pong")
		return nil
	})
	for {
		proto := NewHeartbeatProto()
		c.WritePingMessage(proto)
		time.Sleep(30 * time.Second)
	}
}

// ReadMessage 循环读取websocket发送来的消息
func (c *Client) ReadMessage() {
	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			log.Errorf("client key:%s mid:%d read message error:%v", c.Key, c.Mid, err)
		}
		var proto pb.Proto
		if err := json.Unmarshal(message, &proto); err != nil {
			log.Errorf("json.Unmarshal error:%v", err)
		}
		log.Infof("client key:%s mid:%d op:%d body:%s ", c.Key, c.Mid, proto.Op, string(proto.Body))
		time.Sleep(time.Second * 1)
	}
}

func (c *Client) WritePingMessage(data []byte) {
	if err := c.Conn.WriteMessage(websocket.PingMessage, data); err != nil {
		log.Errorf("client key:%s mid:%d write ping message:%s error:%v", c.Key, c.Mid, string(data), err)
	}
	return
}

func (c *Client) WriteTextMessage(data []byte) {
	if err := c.Conn.WriteMessage(websocket.TextMessage, data); err != nil {
		log.Errorf("client key:%s mid:%d write text message:%s error:%v", c.Key, c.Mid, string(data), err)
	}
	return
}

func (c *Client) Close() {
	log.Infof("client key:%s mid:%d close connect", c.Key, c.Mid)
	c.Conn.Close()
}
