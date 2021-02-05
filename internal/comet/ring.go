/**
* @Time: 2021/1/31 上午11:10
* @Author: miku
* @File: ring
* @Version: 1.0.0
* @Description:
 */

package comet

import (
	"walle/api/protocol"
	"walle/internal/comet/errors"
)

// Ring 用于存储客户端的消息
type Ring struct {
	// read
	rp   uint64 // read pointer
	num  uint64
	mask uint64
	// write
	wp   uint64
	data []protocol.Proto
}

// NewRing new a ring buffer
func NewRing(num int) *Ring {
	r := new(Ring)
	r.init(uint64(num))
	return r
}

// Init init ring
func (r *Ring) Init(num int) {
	r.init(uint64(num))
}

func (r *Ring) init(num uint64) {
	// 2 ^ num
	if num&(num-1) != 0 {
		for num&(num-1) != 0 {
			num &= (num - 1)
		}
		num = num << 1
	}
	r.data = make([]protocol.Proto, num)
	r.num = num
	r.mask = r.num - 1
}

// Get get a proto from ring
func (r *Ring) Get() (proto *protocol.Proto, err error) {
	if r.rp == r.wp {
		return nil, errors.ErrRingEmpty
	}
	proto = &r.data[r.rp&r.mask]
	return
}

// GetAdv incr read index
func (r *Ring) GetAdv() {
	r.rp++
}

// Set get a proto to write
func (r *Ring) Set() (proto *protocol.Proto, err error) {
	if r.wp-r.rp >= r.num {
		return nil, errors.ErrRingFull
	}
	proto = &r.data[r.wp&r.mask]
	return
}

// SetAdv incr write index
func (r *Ring) SetAdv() {
	r.wp++
}

// Reset reset ring
func (r *Ring) Reset() {
	r.rp = 0
	r.wp = 0
}
