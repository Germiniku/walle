/**
* @Time: 2021/1/31 上午11:38
* @Author: miku
* @File: room
* @Version: 1.0.0
* @Description:
 */

package comet

import (
	"sync"
	pb "walle/api/protocol"
	"walle/internal/comet/errors"
)

// Room 聊天室房间
type Room struct {
	ID        string
	next      *Channel
	drop      bool
	Online    int32
	AllOnline int32
	mutex     sync.RWMutex
}

// NewRoom 返回一个新的房间
func NewRoom(id string) *Room {
	r := new(Room)
	r.ID = id
	r.drop = false
	r.next = nil
	r.Online = 0
	return r
}

// Put 创建房间连接
func (r *Room) Put(ch *Channel) (err error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if !r.drop {
		if r.next != nil {
			r.next.Prev = ch
		}
		ch.Next = r.next
		r.next = ch
		r.Online++
	} else {
		err = errors.ErrRoomDroped
	}
	return
}

// Del 删除房间的连接 返回房间是否废弃
func (r *Room) Del(ch *Channel) bool {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if ch.Next != nil {
		// 不是链表尾
		ch.Next.Prev = ch.Prev
	}
	if ch.Prev != nil {
		// 不是链表头
		ch.Prev.Next = ch.Next
	} else {
		r.next = ch.Next
	}
	r.Online--
	r.drop = (r.Online == 0)
	return r.drop
}

// Push 将消息推送到房间里的每个连接
func (r *Room) Push(proto *pb.Proto) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	for ch := r.next; ch != nil; ch = ch.Next {
		ch.Push(proto)
	}
	return
}

func (r *Room) OnlineNum() int32 {
	if r.AllOnline > 0 {
		return r.AllOnline
	}
	return r.Online
}

// Close 关闭房间
func (r *Room) Close() {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	for ch := r.next; ch != nil; ch = ch.Next {
		ch.Close()
	}
	return
}
