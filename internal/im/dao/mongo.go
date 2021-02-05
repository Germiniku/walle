/**
* @Time: 2021/1/30 下午12:33
* @Author: miku
* @File: mysql
* @Version: 1.0.0
* @Description:
 */

package dao

import (
	"context"
	log "github.com/golang/glog"
	"go.mongodb.org/mongo-driver/bson"
	model2 "walle/internal/im/model"
)

// InsertUser 插入新用户
func (d *Dao)InsertUser(c context.Context,username,password string)(user *model2.User,err error){
	newUser := model2.User{Username: username,Password: password}
	result, err := d.DB.User.InsertOne(c, newUser)
	if err != nil{
		log.Errorf("insert user:%s error:%v",username,err)
		return
	}
	filter := bson.D{{
		"id", result.InsertedID,
	}}
	if err = d.DB.User.FindOne(c,filter).Decode(user);err != nil{
		return
	}
	return
}