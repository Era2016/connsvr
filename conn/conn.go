package conn

import (
	"bufio"
	"net"

	"github.com/simplejia/connsvr/comm"
	"github.com/simplejia/connsvr/proto"
)

type ConnWrap struct {
	T    comm.PROTO    // 消息类型
	C    net.Conn      // socket
	Buf  *bufio.Reader // bufio.NewReader(C)
	Uid  string        // 用户
	Sid  string        // session id，区分同一个用户不同连接
	Rids []string      // 房间列表
	Misc interface{}   // 额外参数
}

// return false will close the conn
func (connWrap *ConnWrap) Read() (proto.Msg, bool) {
	msg := proto.NewMsg(connWrap.T)
	if !msg.Decode(connWrap.Buf) {
		return nil, false
	}
	return msg, true
}

// when return false, close the connection
func (connWrap *ConnWrap) Write(msg proto.Msg) bool {
	data, ok := msg.Encode()
	if !ok {
		return true
	}
	for wlen := 0; wlen < len(data); {
		_wlen, err := connWrap.C.Write(data[wlen:])
		if err != nil || _wlen <= 0 {
			return false
		}
		wlen += _wlen
	}
	return true
}

func (connWrap *ConnWrap) Close() {
	// net.Conn可以多次关闭
	connWrap.C.Close()
}
