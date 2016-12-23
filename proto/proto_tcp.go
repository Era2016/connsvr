package proto

import (
	"bufio"
	"encoding/binary"
	"runtime/debug"

	"github.com/simplejia/clog"
	"github.com/simplejia/connsvr/comm"
)

const (
	SBYTE = 0xfa
	EBYTE = 0xfb
)

type MsgTcp struct {
	MsgComm
}

func (msg *MsgTcp) Decode(buf *bufio.Reader) (ok bool) {
	defer func() {
		if err := recover(); err != nil {
			clog.Error("MsgTcp:Decode() recover err: %v, stack: %s", err, debug.Stack())
		}
	}()

	for {
		sbyte, err := buf.ReadByte()
		if err != nil {
			return false
		}
		if sbyte == SBYTE {
			break
		}
	}

	length := [2]byte{}
	for m := 0; m < len(length); {
		n, err := buf.Read(length[m:])
		if err != nil || n <= 0 {
			return false
		}
		m += n
	}

	msg.length = int(binary.BigEndian.Uint16(length[:]))

	data := make([]byte, msg.length-3)
	for m := 0; m < len(data); {
		n, err := buf.Read(data[m:])
		if err != nil || n <= 0 {
			return false
		}
		m += n
	}

	pos := 0
	msg.cmd = comm.CMD(data[pos])
	pos += 1
	msg.subcmd = data[pos]
	pos += 1
	uid_len := int(data[pos])
	pos += 1
	msg.uid = string(data[pos : pos+uid_len])
	pos += uid_len
	sid_len := int(data[pos])
	pos += 1
	msg.sid = string(data[pos : pos+sid_len])
	pos += sid_len
	rid_len := int(data[pos])
	pos += 1
	msg.rid = string(data[pos : pos+rid_len])
	pos += rid_len
	body_len := int(binary.BigEndian.Uint16(data[pos : pos+2]))
	pos += 2
	msg.body = string(data[pos : body_len+pos])
	pos += body_len
	ext_len := int(binary.BigEndian.Uint16(data[pos : pos+2]))
	pos += 2
	msg.ext = string(data[pos : ext_len+pos])
	pos += ext_len
	ebyte := data[pos]
	if ebyte != EBYTE {
		return false
	}

	return true
}

func (msg *MsgTcp) Encode() ([]byte, bool) {
	data := []byte{}
	data = append(data, SBYTE)
	data = append(data, make([]byte, 2)...)
	data = append(data, byte(msg.cmd))
	data = append(data, msg.subcmd)
	data = append(data, byte(len(msg.uid)))
	data = append(data, msg.uid...)
	data = append(data, byte(len(msg.sid)))
	data = append(data, msg.sid...)
	data = append(data, byte(len(msg.rid)))
	data = append(data, msg.rid...)
	data = append(data, make([]byte, 2)...)
	binary.BigEndian.PutUint16(data[len(data)-2:len(data)], uint16(len(msg.body)))
	data = append(data, msg.body...)
	data = append(data, make([]byte, 2)...)
	binary.BigEndian.PutUint16(data[len(data)-2:len(data)], uint16(len(msg.ext)))
	data = append(data, msg.ext...)
	data = append(data, EBYTE)
	binary.BigEndian.PutUint16(data[1:3], uint16(len(data)))

	return data, true
}
