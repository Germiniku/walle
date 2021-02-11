/**
* @Time: 2021/1/28 上午11:51
* @Author: miku
* @File: push
* @Version: 1.0.0
* @Description:
 */

package logic

import (
	"context"
	log "github.com/golang/glog"
	"time"
	"walle/api/protocol"
	"walle/internal/logic/model"
)

const (
	pushUser      = 1
	broadcastRoom = 2
	broadcast     = 3
)

/**
PushKeys: 根据keys推送消息
@Params
op: 操作
*/
func (l *Logic) PushKeys(c context.Context, op int32, keys []string, msg []byte) {
	// 通过keys --> serverID
	servers, err := l.dao.ServersByKeys(keys)
	log.Infof("servers:%v", servers)
	if err != nil {
		return
	}
	/**
	{
		serverId: [key1,key2,key3]
	}
	*/
	pushKeys := make(map[string][]string)
	for index, key := range keys {
		server := servers[index]
		if server != "" && key != "" {
			pushKeys[server] = append(pushKeys[server], key)
		}
	}
	for server := range pushKeys {
		log.Infof("op:%d server:%s keys:%v msg:%s", op, server, keys, string(msg))
		if err = l.dao.MQ.PushMsg(c, op, server, keys, msg); err != nil {
			return
		}
	}
}

// PushMids: 根据用户mid推送消息
func (l *Logic) PushMids(c context.Context, op int32, mids []int64, msg []byte) (err error) {
	keyServers, _, err := l.dao.ServersByMids(mids)
	if err != nil {
		return
	}
	keys := make(map[string][]string)
	for key, server := range keyServers {
		if key == "" || server == "" {
			log.Warningf("push key:%s server:%s is empty", key, server)
			continue
		}
		keys[server] = append(keys[server], key)
	}
	for server, keys := range keys {
		if err = l.dao.MQ.PushMsg(c, op, server, keys, msg); err != nil {
			return
		}
	}
	return
}

func (l *Logic) PushAll(c context.Context, op int32, msg []byte) (err error) {
	return l.dao.MQ.BroadcastMsg(c, op, msg)
}

func (l *Logic) PushRoom(c context.Context, op int32, Type, roomId string, msg []byte) (err error) {
	return l.dao.MQ.BroadcastRoomMsg(c, op, Type, roomId, msg)
}

// chatRecord 存储聊天记录
func (l *Logic) chatRecord(c context.Context, kind uint8, op int32, keys []string, msg []byte, mids []int64, Type string) {
	// 判断是聊天记录
	if op == protocol.OpRaw {

	}
}

// collectChatRecord 存储聊天记录
func (l *Logic) collectChatRecord(ch chan *model.ChatRecord, signal time.Duration) {
	batch := l.c.Server.ChatRecordBatch
	sign := make(chan struct{}, 1)
	chatRecords := make([]*model.ChatRecord, batch*2)
	td := time.AfterFunc(signal, func() {
		sign <- struct{}{}
	})
	for {
		var chatRecord *model.ChatRecord
		select {
		case chatRecord = <-ch:
		default:
			time.Sleep(time.Millisecond * 100)
		}
		if chatRecord != nil {
			chatRecords = append(chatRecords, chatRecord)
			if int32(len(chatRecords)) >= batch {
				l.dao.InsertChatRecords(context.Background(), chatRecords)
				chatRecords = make([]*model.ChatRecord, batch*2)
				td.Reset(signal)
			}
			select {
			case <-sign:
				l.dao.InsertChatRecords(context.Background(), chatRecords)
				chatRecords = make([]*model.ChatRecord, batch*2)
				td.Reset(signal)
			default:
			}
		}
	}
}
