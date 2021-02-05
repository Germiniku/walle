/**
* @Time: 2021/1/28 下午1:30
* @Author: miku
* @File: server
* @Version: 1.0.0
* @Description:
 */

package pipe

import (
	"encoding/json"
	"fmt"
	log "github.com/golang/glog"
	"github.com/golang/protobuf/proto"
	"github.com/nats-io/nats.go"
	"github.com/rcrowley/go-metrics"
	etcd_client "github.com/rpcxio/rpcx-etcd/client"
	"github.com/rpcxio/rpcx-etcd/serverplugin"
	"github.com/smallnest/rpcx/client"
	"github.com/smallnest/rpcx/server"
	"net/url"
	"strings"
	"sync"
	"time"
	pb "walle/api/logic"
	"walle/internal/pipe/conf"
	"walle/internal/pipe/dao"
	"walle/protocol"
)

type Pipe struct {
	conf        *conf.Config
	dao         *dao.Dao
	recvMsg     chan *nats.Msg
	cometServer map[string]*Comet
	rooms       map[string]*Room
	server      *server.Server
	mutex       sync.RWMutex
}

// svcInfo 服务发现服务信息
type svcInfo struct {
	ServerID string `json:"server_id"`
	Ip       string `json:"ip"`
}

func New(conf *conf.Config) *Pipe {
	pipe := &Pipe{
		dao:         dao.New(conf),
		conf:        conf,
		recvMsg:     make(chan *nats.Msg, 1024),
		cometServer: make(map[string]*Comet),
	}
	//go pipe.Register()
	go pipe.watchComet()
	return pipe
}

// Register 服务注册
func (p *Pipe) Register() {
	p.server = server.NewServer()
	p.addRegistryPlugin(p.conf)
	p.server.RegisterName("pipe", p, "")
}

func (p *Pipe) addRegistryPlugin(conf *conf.Config) {
	r := &serverplugin.EtcdV3RegisterPlugin{
		EtcdServers:    conf.Discovery.Endpoints,
		BasePath:       p.getDiscoveryBasePath(),
		Metrics:        metrics.NewRegistry(),
		UpdateInterval: time.Minute,
	}
	err := r.Start()
	if err != nil {
		log.Fatal(err)
	}
	p.server.Plugins.Add(r)
}

// getDiscoveryBasePath 获取注册中心key前缀
func (p *Pipe) getDiscoveryBasePath() string {
	return fmt.Sprintf("/%s/%s", p.conf.Discovery.Env, p.conf.Discovery.AppId)
}

// watchComet 监听注册中心中的comet服务节点更新Pipe中 的cometServer
func (p *Pipe) watchComet() {
	var (
		discovery client.ServiceDiscovery
		err       error
	)
	if discovery, err = etcd_client.NewEtcdV3Discovery(p.getDiscoveryBasePath(), "comet", p.conf.Discovery.Endpoints, nil); err != nil {
		log.Errorf("client.NewEtcdV3Discovery error:%v", err)
		return
	}
	servicesChan := discovery.WatchService()
	for {
		select {
		case services := <-servicesChan:
			for _, service := range services {
				serverInfo := getServerInfo(service.Value)
				if serverInfo == nil {
					continue
				}
				log.Infof("serverID:%s ip:%s", serverInfo.ServerID, serverInfo.Ip)
				comet, err := NewComet(p.conf.Comet, serverInfo.ServerID, serverInfo.Ip)
				if err != nil {
					p.mutex.Unlock()
					continue
				}
				p.cometServer[serverInfo.ServerID] = comet
			}
		default:
			time.Sleep(time.Second * 10)
		}
	}

}

func (p *Pipe) Run() {
	p.worker()
}

// worker 订阅消息并消费消息
func (p *Pipe) worker() {
	subject := protocol.PackagePublishKey(p.conf.Nats.Topic, "*")
	p.dao.RecvMessage(subject, p.recvMsg)
	go p.consume()
}

// consume 消费消息
func (p *Pipe) consume() {
	for {
		msg := <-p.recvMsg
		var m pb.PushMsg
		if err := proto.Unmarshal(msg.Data, &m); err != nil {
			log.Errorf("[worker] proto.Unmarshal err:%v", err)
			continue
		}
		log.Infof("keys:%v operation:%d msg:%s type:%d", m.Keys, m.Operation, string(m.Msg), m.Type)
		// TODO: 这里应该使用线程池
		p.push(&m)
	}
}

func (p *Pipe) Close() {
	close(p.recvMsg)
	p.dao.Close()
	return
}

// getServerInfo 根据服务发现信息获取服务所需信息
func getServerInfo(data string) *svcInfo {
	var svcInfo svcInfo
	item, err := url.PathUnescape(data)
	if err != nil {
		log.Errorf("url.PathUnescape(%s) error:%v", data, err)
		return nil
	}
	left := strings.Index(item, "{")
	right := strings.Index(item, "}")
	item = item[left : right+1]
	if err = json.Unmarshal([]byte(item), &svcInfo); err != nil {
		log.Errorf("json.Unmarshal(%s) error:%v", item, err)
		return nil
	}
	log.Infof("serverId:%s ip:%s", svcInfo.ServerID, svcInfo.Ip)
	return &svcInfo
}
