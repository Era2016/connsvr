package bsvr

import (
	"github.com/simplejia/clog"
	"github.com/simplejia/connsvr/comm"
	"github.com/simplejia/connsvr/proto"
	"github.com/simplejia/connsvr/room"

	"net"
)

func Bserver(host string) {
	udpAddr, err := net.ResolveUDPAddr("udp", host)
	if err != nil {
		panic(err)
	}
	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	msg := new(proto.MsgUdp)
	request := make([]byte, 1024*50)
	for {
		readLen, err := conn.Read(request)
		if err != nil || readLen <= 0 {
			continue
		}

		ok := msg.DecodeBytes(request[:readLen])
		clog.Debug("Bserver() msg.DecodeBytes %+v, %v", msg, ok)
		if !ok {
			clog.Error("Bserver:DecodeBytes() %v", request[:readLen])
			continue
		}

		dispatchCmd(msg)
	}
}

func dispatchCmd(msg proto.Msg) {
	switch msg.Cmd() {
	case comm.PUSH:
		go room.RM.Push(msg)
	default:
		clog.Error("bsvr:dispatchCmd() unexpected cmd: %v", msg.Cmd())
	}
}
