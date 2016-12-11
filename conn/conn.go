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

	// 这里需要设置写超时，因为缓冲区会满，有了超时就不至于导致房间处理goroutine阻塞
	connWrap.C.SetWriteDeadline(time.Now().Add(time.Millisecond))
	for m := 0; m < len(data); {
		n, err := connWrap.C.Write(data[m:])
		if err != nil || n <= 0 {
			return false
		}
		m += n
	}
	return true
}

func (connWrap *ConnWrap) Close() {
	// net.Conn可以多次关闭
	connWrap.C.Close()
}
