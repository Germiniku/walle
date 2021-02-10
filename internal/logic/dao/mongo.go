/**
* @Time: 2021/2/5 上午11:58
* @Author: miku
* @File: mongo
* @Version: 1.0.0
* @Description:
 */

package dao

import (
	"context"
	log "github.com/golang/glog"
	"walle/internal/logic/model"
)

func (d *Dao) InsertChatRecords(c context.Context, records []*model.ChatRecord) {
	collection := d.DB.Collection("chatRecord")
	tmp := make([]interface{}, len(records))
	for _, record := range records {
		tmp = append(tmp, record)
	}
	if _, err := collection.InsertMany(c, tmp); err != nil {
		log.Errorf("insert chat records error:%v", err)
	}
}

// 批量发送
//func (d *Dao) SaveChatRecord(c context.Context, kind uint8, keys []string, msg []byte, mids []int64, Type string) {
//	record := &chatRecord{
//		Kind: kind,
//		Keys: keys,
//		Msg:  string(msg),
//		Mids: mids,
//		Type: Type,
//	}
//	record = record
//}
