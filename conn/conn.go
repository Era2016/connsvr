package conn

import (
	"bufio"
	"net"
	"time"

	"github.com/simplejia/connsvr/comm"
	"github.com/simplejia/connsvr/proto"
)

type ConnWrap struct {
	T    comm.PROTO    // 消息类型
	C    net.Conn      // socket
	BR   *bufio.Reader // bufio.NewReader(C)
	Uid  string        // 用户
	Sid  string        // session id，区分同一个用户不同连接
	Rids []string      // 房间列表
	Misc interface{}   // 额外参数
}

// return false will close the conn
func (connWrap *ConnWrap) Read() (proto.Msg, bool) {
	connWrap.C.SetReadDeadline(time.Now().Add(time.Hour))

	msg := proto.NewMsg(connWrap.T)
	ok := msg.Decode(connWrap.BR, connWrap.C, &connWrap.Misc)
	if !ok {
		return nil, false
	}

	return msg, true
}

// when return false, close the connection
func (connWrap *ConnWrap) Write(msg proto.Msg) bool {
	connWrap.C.SetWriteDeadline(time.Now().Add(time.Millisecond))

	ok := msg.Encode(connWrap.C, connWrap.Misc)
	if !ok {
		return false
	}

	return true
}

func (connWrap *ConnWrap) Close() {
	// net.Conn可以多次关闭
	connWrap.C.Close()
}
