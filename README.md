# [connsvr](http://github.com/simplejia/connsvr) 长连接服务
## 功能
* 支持tcp自定义协议长连接（见下文）
* 支持http协议长连接（long poll机制，ajax挂上去后等待数据返回）
* 每个用户建立一个连接，每个连接唯一对应一个用户，用户可以同时加入多个房间
* 推送数据时，可以不给房间内特定的一个用户推数据，适用于：前端假写数据，长连接服务帮过滤掉这条消息
* 接收到上行数据后，同步转发给相应业务处理服务，可通过conf/conf.json配置pubs节点，connsvr将数据通过http方式路由到后端业务处理服务，然后透传结果到客户端
* 支持定期拉取远程服务器配置数据, 远程服务器的配置见conf/conf.json varhost节点, 配置数据格式见comm.Var:
```
{
	PushKind         comm.PUSH_KIND            // 消息推送方式
	RoomWithPushKind map[string]comm.PUSH_KIND // 消息推送方式, 房间单独配置
	MsgNum           int                       // connsvr服务缓存消息最大长度
}
```

* 根据远程配置信息，可给客户端下发消息拉取方式的指令，目前支持:
  * 推送整条消息，客户端不用拉，适用于：对消息一致性要求不高，丢一两条没关系，这种方式，基本只需要有connsvr就够了，而且对connsvr的要求不高，比如只推送比赛即时比分信息（默认）
  * 推送通知，然后客户端来connsvr或者后端服务拉消息，适用于：当从connsvr拉消息，对消息一致性上有要求，但允许在瞬间消息量比较大的情况下充掉connsvr缓冲消息列表而丢掉部分老的消息，这种方式，基本只需要有connsvr就够了，对connsvr要求较高；当从后端服务拉消息，对消息一致性要求高，这种方式，可以保证用户拉到完整的消息列表，不会由于推送失败而丢消息，后端服务需要提供客户端数据拉取的完整功能
  * 注1：可以配置特定房间单独的拉取方式，见comm.Var.RoomWithPushKind，需要远程配置服务器对每个房间设定自动失效时间
  * 注2：每条推送消息也可单独设置拉取方式，此优先级最高，其次是远程配置服务器单独设置的房间消息拉取方式，最后才是全局设定的配置，此方式适用于某些直播场景送的道具，只在收到此条道具消息推送时才展示全屏效果，过后拉取出来的历史消息列表不需要展示全屏效果

## 安装
> go get -u github.com/simplejia/connsvr

## 实现
* 启用一个协程用于接收后端push数据，启用若干个协程用于管理房间用户，用户被hash到对应协程
* 每个协程需要通过管道接收数据，包括：加入房间，退出房间，推送消息
* 每个用户连接启一个读协程
* 无锁 

## 特点
* 通信协议足够简单高效
* 服务设计尽量简化，通用性好

## 协议
* http长连接
```
** 加入房间 **
http://xxx.xxx.com?cmd=2&subcmd=xxx&rid=xxx&uid=xxx&sid=xxx&body=xxx&callback=xxx
Request Method: get or post
请求参数说明:
cmd: 固定为2
subcmd: 用于区分不同业务，有效数据：0
rid: 房间号
uid: 用户id
sid: session_id，区分同一uid不同连接，[可选]
body: 支持如下：(可以传空串）
{
	MsgIds map[byte]string // 混合业务命令字, key: subcmd, value: msgid
}
callback: jsonp回调函数，[可选]

返回数据说明：
[callback(][json body][)]
示例如下: cb({"body":{"1":["hello world"]},"cmd":"2","rid":"r1","sid":"","subcmd":"0","uid":"r2"})

注1：如果服务端有用户未读消息，并且传入合适的subcmd，立马返回消息列表
```

```
** 退出房间 **
http://xxx.xxx.com?cmd=3&rid=xxx&uid=xxx&sid=xxx
Request Method: get or post
请求参数说明:
cmd: 固定为3
rid: 房间号
uid: 用户id
sid: session_id，区分同一uid不同连接，[可选]

返回数据说明：
无
```

```
** 拉取消息 **
http://xxx.xxx.com?cmd=5&subcmd=xxx&rid=xxx&uid=xxx&sid=xxx&body=xxx&callback=xxx
Request Method: get or post
请求参数说明:
cmd: 固定为5
subcmd: 用于区分不同业务，有效数据：0
rid: 房间号
uid: 用户id
sid: session_id，区分同一uid不同连接，[可选]
body: 支持如下：(可以传空串）
{
	MsgIds map[byte]string // 混合业务命令字, key: subcmd, value: msgid
}
callback: jsonp回调函数，[可选]

返回数据说明：
[callback(][json body][)]
示例如下: cb({"body":{"1":["hello world"]},"cmd":"5","rid":"r1","sid":"","subcmd":"0","uid":"r2"})
```

```
** 上行消息 **
http://xxx.xxx.com?cmd=4&subcmd=xxx&rid=xxx&uid=xxx&sid=xxx&body=xxx&callback=xxx
Request Method: get or post
请求参数说明:
cmd: 固定为4
subcmd: 用于区分不同业务，有效数据：1~255之间
rid: 房间号
uid: 用户id
sid: session_id，区分同一uid不同连接，[可选]
body: 客户端上传内容
callback: jsonp回调函数，[可选]

返回数据说明：
[callback(][json body][)]
示例如下: cb({"body":"","cmd":"4","rid":"r1","sid":"","subcmd":"1","uid":"r2"})

注1：当connsvr服务处理异常，比如调用后端服务失败，返回如下：cb({"body":"","cmd":"-1","rid":"r1","sid":"","subcmd":"1","uid":"r2"})
```

> test文件夹有个ajax长轮询示例：ajax.html，使用方式如下：
  1. 首先配置host: 127.0.0.1 connsvr.com [ip换成connsvr服务对应的ip]
  2. 启动redis-server: redis-server，监听端口用默认的6379
  3. 启动connsvr: ./connsvr -env dev
  4. 启动clog: cd test/clog/server; ./server -env dev 
  5. 启动logicsvr: cd test/logicsvr; ./logicsvr -env dev 
  6. 浏览器里打开ajax.html，可以在url里跟上参数：rid=xxx&uid=xxx&sid=xxx，分别代表房间号和用户id，用户session_id，可以同时开两个tab，然后人别传入不同的uid，rid可以一样
  7. 在文本框内输入字符，点"发送“
  
> 注1：步骤3启动clog，是为了做消息分发，test/clog目录提供了一个定制的clog服务，clog服务来自：https://github.com/simplejia/clog

> 注2：步骤4启动logicsvr，是为了提供一个业务服务demo，用来做发消息后转发消息的，test/logicsvr目录提供了一个定制的demo服务，服务规范来自：https://github.com/simplejia/wsp

> 注3：由于connsvr的ip上报是通过redis存储，所以需要启一个默认的redis-server

> 注4：也可简单测试，这种方式就不能用到ajax.html提供的发送消息功能了，不用执行2，3，4，5步骤，仅运行包含push消息的测试用例：go test -env dev -v -run=TestTcp$



* tcp自定义协议长连接（包括收包，回包）
```
Sbyte+Length+Cmd+Subcmd+UidLen+Uid+SidLen+Sid+RidLen+Rid+BodyLen+Body+ExtLen+Ext+Ebyte

Sbyte: 1个字节，固定值：0xfa，标识数据包开始
Length: 2个字节(网络字节序)，包括自身在内整个数据包的长度
Cmd: 1个字节，
  * 0x01：心跳 // 现在的技术方案用不到心跳
  * 0x02：加入房间 
  * 0x03：退出房间 
  * 0x04：上行消息 
  * 0x05：拉取消息列表 
  * 0x06：推送消息
  * 0xff：标识服务异常
Subcmd: 1个字节，路由不同的后端接口，见conf/conf.json pubs和msgs节点，
  * pubs代表上行消息配置，中转给业务方数据示例如下：uid=u1&sid=s1&rid=r1&cmd=4&subcmd=1&body=xxx，直接把后端返回作为body内容传回给client
  * msgs代表拉消息列表配置，中转给业务方数据示例如下：uid=u1&sid=s1&rid=r1&cmd=5&subcmd=1，返回给client的body内容示例如下：{"1":["hello world"]}
UidLen: 1个字节，代表Uid长度
Uid: 用户id，对于app，可以是设备id，对于浏览器，可以是登陆用户id
SidLen: 1个字节，代表Sid长度
Sid: session_id，区分同一uid不同连接，对于浏览器，可以是生成的随机串，浏览器多窗口，多标签需单独生成随机串
RidLen: 1个字节，代表Rid长度
Rid: 房间id
BodyLen: 2个字节(网络字节序)，代表Body长度
Body: 和业务方对接，connsvr会中转给业务方
ExtLen: 2个字节(网络字节序)，代表Ext长度
Ext: 扩展字段:
1. 当来自于connsvr时，目前支持如下：
{    
    "PushKind": 1 // 1: 推送通知，然后客户端主动拉后端服务 2: 推送整条消息，客户端不用拉（默认） 3: 推送通知，然后客户端来connsvr拉消息   
}
2. 当来自于client时，目前支持如下：
{    
    "Cookie": "xx=x;yy=y" // 传入client的cookie值
}
Ebyte: 1个字节，固定值：0xfb，标识数据包结束

注1：当connsvr服务处理异常，比如调用后端服务失败，返回给client的数据包，Cmd：0xff
注2：当Cmd为0x05时，客户端到connsvr拉取消息列表，当connsvr消息为空时，connsvr为根据conf/conf.json msgs节点配置路由到后端服务拉取消息列表，body支持如下：
{
	MsgIds map[byte]string // 混合业务命令字, key: subcmd, value: msgid
}
注3：当Cmd为0x02时，如果服务端有用户未读消息，并且传入合适的body，立马返回消息列表, body支持如下：
{
	MsgIds map[byte]string // 混合业务命令字, key: subcmd, value: msgid
}
```

* 后端push协议格式(udp)
```
Cmd+Subcmd+UidLen+Uid+SidLen+Sid+RidLen+Rid+BodyLen+Body+ExtLen+Ext:

Cmd: 1个字节，经由connsvr直接转发给client
Subcmd: 1个字节，经由connsvr直接转发给client，有效范围: 1~255
UidLen: 1个字节，代表Uid长度
Uid: 指定排除的用户uid
SidLen: 1个字节，代表Sid长度
Sid: 指定排除的用户session_id，当没有传入Sid时，只匹配uid
RidLen: 1个字节，代表Rid长度
Rid: 房间id
BodyLen: 2个字节(网络字节序)，代表Body长度
Body: 和业务方对接，connsvr会中转给client
ExtLen: 2个字节(网络字节序)，代表Ext长度
Ext: 扩展字段，目前支持如下：
{    
    "MsgId": "1234" // 标识本条消息id      
    "PushKind": 1 // 1: 推送通知，然后客户端主动拉connsvr或者后端服务 2: 推送整条消息，客户端不用拉（默认）
}
注1：数据包长度限制50k内
注2：当Ext.MsgId每次传相同id，比如“1”，connsvr消息列表只会缓存最新的唯一的一条消息
注3：当Ext.MsgId每次传空id: “”，connsvr消息列表不会缓存任何消息
```

## 使用方法
* 配置文件：[conf.json](http://github.com/simplejia/connsvr/tree/master/conf/conf.json) (json格式，支持注释)，可以通过传入自定义的env及conf参数来重定义配置文件里的参数，如：./connsvr -env dev -conf='app.hport=80;clog.mode=1'，多个参数用`;`分隔
* 建议用[cmonitor](http://github.com/simplejia/cmonitor)做进程启动管理
* api文件夹提供的代码用于后端服务给connsvr推送消息的，实际是通过[clog](http://github.com/simplejia/clog)服务分发的
* connsvr的上报数据，比如本机ip定期上报（用于更新待推送服务器列表），连接数、推送用时上报，等等，这些均是通过clog服务中转实现，所以我提供了clog的handler，可以在test/clog目录找到相关代码，具体用到clog配置，部分如下：(见test/clog/server/conf/conf.json)
```
"connsvr/logbusi_report": [
    {
        "handler": "connreporthandler",
        "params": {
            "redis": {"addrtype": "ip", "addr": ":6379"}
        }
    }
],
"connsvr/logbusi_stat": [
    {
        "handler": "connstathandler",
        "params": {}
    }
],
"logicsvr/logbusi_push": [
    {
        "handler": "connpushhandler",
        "params": {
            "redis": {"addrtype": "ip", "addr": ":6379"}
        }
    }
]
```
