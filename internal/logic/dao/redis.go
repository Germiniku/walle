/**
* @Time: 2021/1/28 上午10:58
* @Author: miku
* @File: redis
* @Version: 1.0.0
* @Description:
 */

package dao

import (
	"context"
	"fmt"
	log "github.com/golang/glog"
	"github.com/gomodule/redigo/redis"
)

const (
	_prefixMidServer    = "mid_%d" // mid -> key:server
	_prefixKeyServer    = "key_%s" // key -> server
	_prefixServerOnline = "ol_%s"  // server -> online
)

func keyMidServer(mid int64) string {
	return fmt.Sprintf(_prefixMidServer, mid)
}

func keyKeyServer(key string) string {
	return fmt.Sprintf(_prefixKeyServer, key)
}

func keyServerOnline(key string) string {
	return fmt.Sprintf(_prefixServerOnline, key)
}

func (d *Dao) pingRedis() (err error) {
	conn := d.Redis.Get()
	_, err = conn.Do("SET", "PING", "PONG")
	conn.Close()
	return
}

// AddExpire 会话状态增加过期时间
func (d *Dao) AddExpire(c context.Context, mid int64, key, server string) (has bool, err error) {
	conn := d.Redis.Get()
	defer conn.Close()
	receiveCount := 2
	if err = conn.Send("EXPIRE", keyMidServer(mid), d.redisExpire); err != nil {
		log.Errorf("[AddExpire] conn.Send(EXPIRE,%s,%s)", keyMidServer(mid), d.redisExpire)

	}
	if err = conn.Send("EXPIRE", keyKeyServer(key), d.redisExpire); err != nil {
		log.Errorf("[AddExpire] conn.Send(EXPIRE,%s,%s)", keyKeyServer(key), d.redisExpire)
	}
	if err = conn.Flush(); err != nil {
		log.Errorf("conn.Flush() error:%v", err)
	}
	for i := 0; i < receiveCount; i++ {
		if has, err = redis.Bool(conn.Receive()); err != nil {
			log.Errorf("conn.Receive() error:%v", err)
			return
		}
	}
	return
}

// AddMapping mid-key-server映射关系
func (d *Dao) AddMapping(c context.Context, mid int64, key, server string) (err error) {
	conn := d.Redis.Get()
	defer conn.Close()
	var n = 2
	if mid > 0 {
		if err = conn.Send("HSET", keyMidServer(mid), key, server); err != nil {
			log.Errorf("conn.Send(HSET %d,%s,%s) error(%v)", mid, server, key, err)
			return
		}
		if err = conn.Send("EXPIRE", keyMidServer(mid), d.redisExpire); err != nil {
			log.Errorf("conn.Send(EXPIRE %d,%s,%s) error(%v)", mid, key, server, err)
			return
		}
		n += 2
	}
	if err = conn.Send("SET", keyKeyServer(key), server); err != nil {
		log.Errorf("conn.Send(HSET %d,%s,%s) error(%v)", mid, server, key, err)
		return
	}
	if err = conn.Send("EXPIRE", keyKeyServer(key), d.redisExpire); err != nil {
		log.Errorf("conn.Send(EXPIRE %d,%s,%s) error(%v)", mid, key, server, err)
		return
	}
	if err = conn.Flush(); err != nil {
		log.Errorf("conn.Flush() error(%v)", err)
		return
	}
	for i := 0; i < n; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Errorf("conn.Receive() error(%v)", err)
			return
		}
	}
	return
}

func (d *Dao) DelMapping(c context.Context, mid int64, key, server string) (err error) {
	conn := d.Redis.Get()
	defer conn.Close()
	n := 2
	if err = conn.Send("DEL", keyKeyServer(key)); err != nil {
		log.Errorf("conn.Send(DEL,%s)", keyKeyServer(key))
	}
	if err = conn.Send("HDEL", keyMidServer(mid)); err != nil {
		log.Errorf("conn.Send(DEL,%s)", keyKeyServer(key))
	}
	if err = conn.Flush(); err != nil {
		log.Errorf("conn.Flush() error:%v", err)
	}
	for i := 0; i < n; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Errorf("conn.Receive() error:%v", err)
			return
		}
	}
	return
}

// ServersByKeys  keys ---> servers
func (d *Dao) ServersByKeys(keys []string) (result []string, err error) {
	conn := d.Redis.Get()
	defer conn.Close()
	var args []interface{}
	for _, key := range keys {
		args = append(args, "key_"+key)
	}
	if result, err = redis.Strings(conn.Do("MGET", args...)); err != nil {
		log.Errorf(`conn.DO("MGET",%v)`, args)
	}
	return
}

// ServersByMids mids --> servers
func (d *Dao) ServersByMids(mids []int64) (result map[string]string, olMids []int64, err error) {
	conn := d.Redis.Get()
	defer conn.Close()
	for _, mid := range mids {
		if err = conn.Send("HGETALL", keyMidServer(mid)); err != nil {
			log.Errorf("conn.Do(HGETALL,%s) error:%v", keyMidServer(mid), err)
			return
		}
	}
	if err = conn.Flush(); err != nil {
		log.Errorf("conn.Flush() error:%v", err)
	}
	for i := 0; i < len(mids); i++ {
		var res map[string]string
		if res, err = redis.StringMap(conn.Receive()); err != nil {
			log.Errorf("conn.Receive() error:%v", err)
			return
		}
		if len(res) > 0 {
			olMids = append(olMids, mids[i])
		}
		for k, v := range res {
			result[k] = v
		}
	}
	return
}
