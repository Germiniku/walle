/**
* @Time: 2021/1/31 上午11:29
* @Author: miku
* @File: error
* @Version: 1.0.0
* @Description:
 */

package errors

import "errors"

var (
	// server
	ErrHandshake = errors.New("handshake failed")
	ErrOperation = errors.New("request operation not valid")
	// ring
	ErrRingEmpty = errors.New("ring buffer empty")
	ErrRingFull  = errors.New("ring buffer full")
	// timer
	ErrTimerFull   = errors.New("timer full")
	ErrTimerEmpty  = errors.New("timer empty")
	ErrTimerNoItem = errors.New("timer item not exist")
	// channel
	ErrPushMsgArg   = errors.New("rpcx pushmsg arg error")
	ErrPushMsgsArg  = errors.New("rpcx pushmsgs arg error")
	ErrMPushMsgArg  = errors.New("rpcx mpushmsg arg error")
	ErrMPushMsgsArg = errors.New("rpcx mpushmsgs arg error")
	// bucket
	ErrBroadCastArg     = errors.New("rpcx broadcast arg error")
	ErrBroadCastRoomArg = errors.New("rpcx broadcast  room arg error")

	// room
	ErrRoomDroped = errors.New("room droped")
	// rpcx
	ErrLogic = errors.New("logic rpcx is not available")
)
