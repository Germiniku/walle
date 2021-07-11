# walle

Walle来自于机器人总动员

![avatar](https://ss2.bdstatic.com/70cFvnSh_Q1YnxGkpoWK1HF6hhy/it/u=3370454460,1552705626&fm=26&gp=0.jpg)

golang实现的消息推送系统(目前只实现了websocket)，场景参考: 直播的弹幕系统

> rpc框架使用rpcx
>
> 注册中心使用etcd
>
> 消息队列支持
>> kafka (生产推荐)
>>
>> nats (部署方便 轻量 测试推荐)

> 目前实现内容:
> 1. 单独私聊消息
> 2. 房间消息
> 3. 全频道广播消息

> 计划内容:
> 1. 实现消息推送过程中的中间件

![avatar](https://github.com/Germiniku/walle/blob/main/assets/process.png)

```
推送消息流程:
客户端首先与comet建立websocket连接(通过websocket发送协议信息实现切换房间等等操作),与websocket建立连接以后
comet调用logic rpc服务获取用户信息(key)。给用户连接分配到bucket中进行管理连接。
用户调用logic的http请求发送私聊/房间群聊/全频道广播消息。logic将消息推送到消息队列(这里可用做白名单、过滤🈲消息)
pipe服务订阅消息队列中的消息，获取消息后通过rpc调用将消息转发到对应comet服务上，comet将消息推送到目标的连接。 
```
```
logic通过redis管理在线会话(session),在redis中每个会话结构为
key: logic随机生成的每个连接唯一id
mid: 用户的id
server: 对应comet服务的服务id
{key: server}
mid: {
    key: server 
}
在创建websocket连接的时候comet会调用logic，logic在redis中生成session，后续客户端websocket心跳保活的时候触发
rpc调用logic给在线会话信息添加过期时间
```

本项目目前仅供参考，仅供测试学习。借鉴于bilibili毛剑的goim

部署方法cmd目录下的comet、logic、pipe文件编译运行，环境依赖etcd
