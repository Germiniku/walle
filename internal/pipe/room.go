/**
* @Time: 2021/2/4 下午4:40
* @Author: miku
* @File: room
* @Version: 1.0.0
* @Description:
 */

package pipe

import (
	"errors"
	log "github.com/golang/glog"
	"time"
	"walle/api/protocol"
	"walle/internal/pipe/conf"
	"walle/pkg/bytes"
)

var (
	ErrRoomFull    = errors.New("房间推送消息已满")
	roomReadyProto = new(protocol.Proto)
)

type Room struct {
	c      *conf.Room
	pipe   *Pipe
	roomID string
	proto  chan *protocol.Proto
}

func newRoom(p *Pipe, roomID string, c *conf.Room) *Room {
	r := &Room{
		c:      c,
		pipe:   p,
		roomID: roomID,
		proto:  make(chan *protocol.Proto, c.Batch*2),
	}
	// 推送消息
	go r.pushproc(c.Batch, time.Duration(c.Signal))
	return r
}

// pushproc 向comet推送消息
func (r *Room) pushproc(batch int, sigTime time.Duration) {
	var (
		p    *protocol.Proto
		n    int
		last time.Time
		buf  = bytes.NewWriterSize(int(protocol.MaxBodySize))
	)
	log.Infof("start room:%s goroutine", r.roomID)
	td := time.AfterFunc(sigTime, func() {
		select {
		case r.proto <- roomReadyProto:
		default:
		}
	})
	defer td.Stop()
	for {
		if p = <-r.proto; p == nil {
			break
		} else if p != roomReadyProto {
			// 批量推送消息
			p.WriteTo(buf)
			if n++; n == 1 {
				last = time.Now()
				td.Reset(sigTime)
				continue
			} else if n < batch {
				if sigTime > time.Since(last) {
					continue
				}
			}
		} else {
			if n == 0 {
				break
			}
		}
		// 向comet推送房间广播消息
		r.pipe.broadcastRoomByBatch(r.roomID, buf.Buffer())
		buf = bytes.NewWriterSize(buf.Size())
		n = 0
		if r.c.Idle != 0 {
			td.Reset(time.Duration(r.c.Idle))
		} else {
			td.Reset(time.Minute)
		}
	}
	r.pipe.delRoom(r.roomID)
	log.Infof("room:%s goroutine exit", r.roomID)
}

func (r *Room) Push(op int32, msg []byte) (err error) {
	p := &protocol.Proto{
		Op:   op,
		Ver:  1,
		Body: msg,
	}
	select {
	case r.proto <- p:
	default:
		err = ErrRoomFull
	}
	return
}

// delRoom 删除房间
func (p *Pipe) delRoom(roomID string) {
	p.mutex.Lock()
	delete(p.rooms, roomID)
	p.mutex.Unlock()
	return
}

// getRoom 获取房间
func (p *Pipe) getRoom(roomID string) *Room {
	p.mutex.RLock()
	room, ok := p.rooms[roomID]
	p.mutex.RUnlock()
	if !ok {
		p.mutex.Lock()
		if room, ok = p.rooms[roomID]; !ok {
			room = newRoom(p, roomID, p.conf.Room)
			p.rooms[roomID] = room
		}
		p.mutex.Unlock()
	}
	return room
}
