/**
* @Time: 2021/1/30 下午2:19
* @Author: miku
* @File: auth
* @Version: 1.0.0
* @Description:
 */

package http

import (
	"github.com/gin-gonic/gin"
)

func (s *Server)authJoin(c *gin.Context){
	var args struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBind(&args);err != nil{
		errors(c,INVAILD_PARAMS,"缺少参数")
		return
	}
	user, err := s.dao.InsertUser(c, args.Username, args.Password)
	if err != nil{
		errors(c,SERVER_ERROR,"数据库异常")
		return
	}
	result(c,OK,user)
	return
}
