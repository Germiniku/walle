/**
* @Time: 2021/1/28 上午10:53
* @Author: miku
* @File: logic
* @Version: 1.0.0
* @Description:
 */

package logic

import (
	"context"
	"encoding/json"
	log "github.com/golang/glog"
	"github.com/google/uuid"
	"time"
	"walle/internal/logic/conf"
	"walle/internal/logic/dao"
	"walle/internal/logic/model"
)

/**
logic
负责处理comet的连接 映射key --> serverID userID --> RoomID
负责发送msg到kafka
*/

type Logic struct {
	c   *conf.Config
	dao *dao.Dao
	// online
	totalCoons int64
	roomCount  map[string]int32
	// collect chat records
	chatRecordCh []chan *model.ChatRecord
}

func New(c *conf.Config) *Logic {
	logic := new(Logic)
	logic.c = c
	logic.dao = dao.New(c)
	if err := logic.Ping(); err != nil {
		log.Errorf("logic Ping error:%s", err)
		panic(err)
	}
	collect := false
	if collect {
		// 创建收集聊天记录的channel
		logic.chatRecordCh = make([]chan *model.ChatRecord, c.Server.ChatRecordRoutine)
		for _, ch := range logic.chatRecordCh {
			ch = make(chan *model.ChatRecord, c.Server.ChatRecordBatch)
			go logic.collectChatRecord(ch, time.Second*2)
		}
	}
	return logic
}

func (l *Logic) Ping() (err error) {
	return l.dao.Ping()
}

func (l *Logic) Close() {
	l.dao.Close()
}

// Connect
func (l *Logic) Connect(c context.Context, server, cookie string, token []byte) (mid int64, key, roomID string, accepts []int32, err error) {
	var params struct {
		Mid      int64   `json:"mid"`
		Key      string  `json:"key"`
		RoomID   string  `json:"room_id"`
		Platform string  `json:"platform"`
		Accepts  []int32 `json:"accepts"`
	}
	if err = json.Unmarshal(token, &params); err != nil {
		log.Errorf("json.Unmarshal(%s) error(%v)", token, err)
		return
	}
	mid, key, roomID, accepts = params.Mid, params.Key, params.RoomID, params.Accepts
	if key = params.Key; key == "" {
		key = uuid.New().String()
	}
	if err = l.dao.AddMapping(c, mid, key, server); err != nil {
		log.Errorf("l.dao.AddMapping(c,%d,%s,%s) error:%v", mid, key, server, err)
	}
	log.Infof("conn connected key:%s mid:%d server:%s", key, mid, server)
	return
}

// Disconnect
func (l *Logic) Disconnect(c context.Context, mid int64, key, server string) (err error) {
	if err = l.dao.DelMapping(c, mid, key, server); err != nil {
		log.Errorf("l.dao.DelMappint(c,%d,%s,%s) error:%v", mid, key, server)
	}
	return err
}

// Heartbeat 心跳保活
func (l *Logic) Heartbeat(c context.Context, mid int64, key, server string) (err error) {
	/*
		给会话添加过期时间
		如果key不存在则添加会话信息
	*/
	has, err := l.dao.AddExpire(c, mid, key, server)
	if err != nil {
		log.Errorf("l.dao.AddExpire(c,%d,%s,%s) error:%v", mid, key, server)
	}
	if !has {
		err = l.dao.AddMapping(c, mid, key, server)
	}
	return
}
