/**
* @Time: 2021/1/31 下午6:47
* @Author: miku
* @File: bucket
* @Version: 1.0.0
* @Description:
 */

package comet

import (
	log "github.com/golang/glog"
	"sync"
	"sync/atomic"
	pb "walle/api/comet"
	proto "walle/api/protocol"
	"walle/internal/comet/conf"
)

type Bucket struct {
	c     *conf.Bucket
	mutex sync.RWMutex
	chs   map[string]*Channel // map key to channel
	// rooms
	rooms  map[string]*Room // map
	ipCnts map[string]uint
	// room
	routines    []chan *pb.BroadcastRoomReq // 节点房间的总信道数
	routinesNum uint64                      // 处理routines信道的goroutine个数，用于消费房播的信道消息
}

func NewBucket(c *conf.Bucket) *Bucket {
	bucket := new(Bucket)
	bucket.c = c
	bucket.chs = make(map[string]*Channel, c.Size)
	bucket.rooms = make(map[string]*Room, c.Room)
	bucket.routines = make([]chan *pb.BroadcastRoomReq, c.RoutineAmount)
	bucket.ipCnts = make(map[string]uint)
	for i := uint64(0); i < c.RoutineAmount; i++ {
		ch := make(chan *pb.BroadcastRoomReq, c.RoutineSize)
		bucket.routines[i] = ch
		go bucket.roomprocess(ch)
	}
	return bucket
}

func (b *Bucket) Rooms() map[string]*Room {
	return b.rooms
}

func (b *Bucket) Put(rid string, ch *Channel) (err error) {
	var (
		room *Room
		ok   bool
	)
	b.mutex.Lock()
	if oldCh := b.chs[ch.Key]; oldCh != nil {
		oldCh.Close()
	}
	b.chs[ch.Key] = ch
	if rid != "" {
		if room, ok = b.rooms[rid]; !ok {
			room = NewRoom(rid)
			b.rooms[rid] = room
		}
		ch.Room = room
	}
	b.ipCnts[ch.IP]++
	b.mutex.Unlock()
	if room != nil {
		room.Put(ch)
	}
	return
}

// roomprocess 处理房间通道
func (b *Bucket) roomprocess(ch chan *pb.BroadcastRoomReq) {
	for {
		req := <-ch
		if room := b.Room(req.RoomID); room != nil {
			room.Push(req.Proto)
		}
	}
}

// BroadcastRoom 房间广播消息
func (b *Bucket) BroadcastRoom(arg *pb.BroadcastRoomReq) {
	num := atomic.AddUint64(&b.routinesNum, 1) % b.c.RoutineAmount
	b.routines[num] <- arg
}

// Broadcast 广播消息
func (b *Bucket) Broadcast(proto *proto.Proto, op int32) {
	for _, ch := range b.chs {
		if ok := ch.NeedPush(op); ok {
			ch.Push(proto)
		}
	}
	return
}

func (b *Bucket) Channel(key string) *Channel {
	log.Infof("now bucket channel chs:%v", b.chs)
	if channel, ok := b.chs[key]; ok {
		return channel
	}
	return nil
}

func (b *Bucket) Room(roomID string) *Room {
	if room, ok := b.rooms[roomID]; ok {
		return room
	}
	return nil
}

// ChannelCount chanel总数
func (b *Bucket) ChannelCount() int {
	return len(b.chs)
}

// DelRoom 从bukcet中删除房间
func (b *Bucket) DelRoom(room *Room) {
	b.mutex.Lock()
	delete(b.rooms, room.ID)
	b.mutex.Unlock()
	room.Close()
}

func (b *Bucket) ChangeRoom(nrid string, ch *Channel) (err error) {
	var (
		nroom *Room
		ok    bool
		oroom = ch.Room
	)
	// 退出房间
	if nrid == "" {
		if oroom != nil {
			oroom.Online--
			oroom.Del(ch)
			if oroom.Online == 0 {
				b.DelRoom(oroom)
			}
		}
		ch.Room = nil
		return
	}
	b.mutex.Lock()
	// 房间不存在则创建房间
	if nroom, ok = b.rooms[nrid]; !ok {
		nroom = NewRoom(nrid)
		b.rooms[nrid] = nroom
	}
	b.mutex.Unlock()
	if err = nroom.Put(ch); err != nil {
		return
	}
	ch.Room = nroom
	return
}
