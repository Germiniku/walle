/**
* @Time: 2021/2/5 下午12:13
* @Author: miku
* @File: chat
* @Version: 1.0.0
* @Description:
 */

package model

type ChatRecord struct {
	Kind uint8
	Keys []string
	Msg  string
	Mids []int64
	Type string
}
