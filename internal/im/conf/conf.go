/**
* @Time: 2021/1/30 下午2:15
* @Author: miku
* @File: conf
* @Version: 1.0.0
* @Description:
 */

package conf

type Config struct {
	HTTPServer *HTTPServer
	DB *DB
}


type HTTPServer struct {
	Addr string
}

type DB struct {
	Addr string
	Port int
	ConnectTimeout int
	MinPoolSize uint64
	MaxPoolSize uint64
	HeartbeatInterval int
	DB string
}