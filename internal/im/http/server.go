/**
* @Time: 2021/1/28 上午10:30
* @Author: miku
* @File: server
* @Version: 1.0.0
* @Description:
 */

package http

import (
	"github.com/gin-gonic/gin"
	"walle/internal/im/conf"
	"walle/internal/im/dao"
)

type Server struct {
	engine *gin.Engine
	dao *dao.Dao
}

func New(conf *conf.Config)*Server{
	engine := gin.Default()
	d := dao.New(conf)
	go func() {
		engine.Run(conf.HTTPServer.Addr)
	}()
	server := &Server{
		engine: engine,
		dao: d,
	}
	server.initRouter()
	return server
}

func (s *Server)initRouter(){
	api := s.engine.Group("/api")
	api.POST("/auth/join",s.authJoin)

}

func (s *Server)Close(){
	return
}