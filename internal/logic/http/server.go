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
	"walle/internal/logic"
	"walle/internal/logic/conf"
)

type Server struct {
	engine *gin.Engine
	logic  *logic.Logic
}

func New(logic *logic.Logic, conf *conf.HTTPServer) *Server {
	engine := gin.Default()
	go func() {
		engine.Run(conf.Addr)
	}()
	server := &Server{
		engine: engine,
		logic:  logic,
	}
	server.initRouter()
	return server
}

func (s *Server) initRouter() {
	api := s.engine.Group("/api")
	api.POST("/push/keys", s.pushKeys)
	api.POST("/push/room", s.pushRoom)
	api.POST("/push/all", s.pushAll)
	api.POST("/push/mids", s.pushMids)
}

func (s *Server) Close() {
	return
}
