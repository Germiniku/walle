/**
* @Time: 2021/1/28 上午10:35
* @Author: miku
* @File: result
* @Version: 1.0.0
* @Description:
 */

package http

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

const (
	OK = 200
	REQUEST_ERR = 400
	SERVER_ERROR = 500
	INVAILD_PARAMS = 401

)

type Response struct {
	Code int32 `json:"code"`
	Message string `json:"message"`
	Data interface{} `json:"data"`
}

func errors(c *gin.Context,code int32,message string){
	c.JSON(http.StatusOK,&Response{
		Code: code,
		Message: message,
	})
	return
}

func result(c *gin.Context,code int32,data interface{}){
	c.JSON(http.StatusOK, &Response{
		Code: code,
		Data: data,
	})
	return
}

