/**
* @Time: 2021/1/30 下午12:40
* @Author: miku
* @File: user
* @Version: 1.0.0
* @Description:
 */

package model

type User struct {
	Username string
	Password string
	Sex uint8  // 0: 男 1:女 2:未知
	Token string
	Status uint8 // 1:冻结 2: 正常
}

