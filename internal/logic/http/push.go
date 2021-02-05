/**
* @Time: 2021/1/28 上午10:33
* @Author: miku
* @File: push
* @Version: 1.0.0
* @Description:
 */

package http

import (
	"github.com/gin-gonic/gin"
	log "github.com/golang/glog"
	"io/ioutil"
)

// pushKeys 根据keys 推送消息
func (s *Server) pushKeys(c *gin.Context) {
	var args struct {
		Operation int32    `form:"operation" binding:"required"`
		Keys      []string `form:"keys" binding:"required"`
	}
	if err := c.BindQuery(&args); err != nil {
		errors(c, REQUEST_ERR, err.Error())
		return
	}
	msg, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		errors(c, REQUEST_ERR, err.Error())
		return
	}
	log.Infof("receive push msg keys:%v operation:%d msg:%s", args.Keys, args.Operation, string(msg))
	s.logic.PushKeys(c, args.Operation, args.Keys, msg)
	result(c, OK, nil)
}

// pushRoom 向房间推送消息
func (s *Server) pushRoom(c *gin.Context) {
	var args struct {
		Op   int32  `json:"op" binding:"required"`
		Type string `json:"type" binding:"required"`
		Room string `json:"room" binding:"required"`
	}
	if err := c.BindQuery(&args); err != nil {
		errors(c, REQUEST_ERR, err.Error())
		return
	}
	msg, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		errors(c, REQUEST_ERR, err.Error())
		return
	}
	if err := s.logic.PushRoom(c, args.Op, args.Type, args.Room, msg); err != nil {
		result(c, REQUEST_ERR, err.Error())
		return
	}
	result(c, OK, nil)
}

// pushAll 全频道推送消息
func (s *Server) pushAll(c *gin.Context) {
	var args struct {
		Op int32 `json:"operation" binding:"required"`
	}
	if err := c.BindQuery(&args); err != nil {
		errors(c, REQUEST_ERR, err.Error())
		return
	}
	msg, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		errors(c, REQUEST_ERR, err.Error())
		return
	}
	if err := s.logic.PushAll(c, args.Op, msg); err != nil {
		errors(c, REQUEST_ERR, err.Error())
		return
	}
	result(c, OK, nil)
	return
}

// pushMids 根据mid推送消息
func (s *Server) pushMids(c *gin.Context) {
	var args struct {
		Op   int32   `json:"opartion" binding:"required"`
		Mids []int64 `json:"mids" binding:"required"`
	}
	if err := c.BindQuery(&args); err != nil {
		errors(c, REQUEST_ERR, err.Error())
		return
	}
	// read message
	msg, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		errors(c, REQUEST_ERR, err.Error())
		return
	}
	if err := s.logic.PushMids(c, args.Op, args.Mids, msg); err != nil {
		errors(c, REQUEST_ERR, err.Error())
		return
	}
	result(c, OK, nil)
	return
}
