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

	for {
		msg := proto.NewMsg(comm.UDP)
		ok := msg.Decode(nil, conn, nil)
		if !ok {
			continue
		}

		clog.Debug("Bserver() msg.DecodeBytes %+v", msg)
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
