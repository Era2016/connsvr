package core

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/simplejia/clog/api"
	"github.com/simplejia/connsvr/comm"
	"github.com/simplejia/connsvr/conf"
	"github.com/simplejia/connsvr/proto"
	"github.com/simplejia/utils"
)

const (
	ADD = iota + 1 // 1
	DEL
	PUSH
)

type roomMsg struct {
	cmd  int
	rid  string
	uid  string
	sid  string
	body interface{}
}

type RoomMap struct {
	n   int
	chs []chan *roomMsg
}

func (roomMap *RoomMap) init() {
	roomMap.n = conf.C.Cons.U_MAP_NUM
	roomMap.chs = make([]chan *roomMsg, roomMap.n)
	for i := 0; i < roomMap.n; i++ {
		roomMap.chs[i] = make(chan *roomMsg, conf.C.Cons.ROOM_MSG_LEN)
		go roomMap.proc(i)
	}
}

func (roomMap *RoomMap) proc(i int) {
	ch := roomMap.chs[i]
	data := map[string]map[[2]string]*ConnWrap{}

	for msg := range ch {
		switch msg.cmd {
		case ADD:
			rid := msg.rid
			users, ok := data[rid]
			if !ok {
				users = map[[2]string]*ConnWrap{}
				data[rid] = users
			}
			ukey := [2]string{msg.uid, msg.sid}
			if v, ok := users[ukey]; ok && v.C != msg.body.(*ConnWrap).C {
				// TODO
				// multi add, you can close old conn
			}
			users[ukey] = msg.body.(*ConnWrap)
		case DEL:
			rid := msg.rid
			users, ok := data[rid]
			if !ok {
				break
			}
			ukey := [2]string{msg.uid, msg.sid}
			if v, ok := users[ukey]; !ok || v.C != msg.body.(*ConnWrap).C {
				break
			}

			delete(users, ukey)
			if len(users) == 0 {
				delete(data, rid)
			}
		case PUSH:
			rid := msg.rid
			users, _ := data[rid]
			if len(users) == 0 {
				break
			}

			var pushExt *comm.PushExt
			m := msg.body.(proto.Msg)
			if ext := m.Ext(); ext != "" {
				err := json.Unmarshal([]byte(ext), &pushExt)
				if err != nil {
					clog.Error("RoomMap:proc() json.Unmarshal error: %v", err)
					break
				}
			}

			servExt := &comm.ServExt{}
			if pushExt != nil && pushExt.PushKind != 0 {
				servExt.PushKind = pushExt.PushKind
			} else if kind, ok := conf.V.Get().RoomWithPushKind[rid]; ok {
				servExt.PushKind = kind
			} else {
				servExt.PushKind = conf.V.Get().PushKind
			}

			servExtBs, _ := json.Marshal(servExt)
			ukeyEx := [2]string{m.Uid(), m.Sid()}
			btime := time.Now()
			for ukey, connWrap := range users {
				if ukeyEx[1] == "" { // 当后端没有传入sid时，只匹配uid
					if ukeyEx[0] == ukey[0] {
						continue
					}
				} else {
					if ukeyEx == ukey {
						continue
					}
				}
				msg := proto.NewMsg(connWrap.T)
				msg.SetCmd(m.Cmd())
				msg.SetSubcmd(m.Subcmd())
				msg.SetUid(connWrap.Uid)
				msg.SetSid(connWrap.Sid)
				msg.SetRid(m.Rid())

				if kind := servExt.PushKind; kind == comm.DISPLAY {
					msg.SetBody(m.Body())
				}

				msg.SetExt(string(servExtBs))
				connWrap.Write(msg)
			}
			etime := time.Now()
			stat, _ := json.Marshal(&comm.Stat{
				Ip:    utils.LocalIp,
				N:     i,
				Rid:   rid,
				Msg:   fmt.Sprintf("%+v", m),
				Num:   len(users),
				Btime: btime,
				Etime: etime,
			})
			clog.Busi(comm.BUSI_STAT, "%s", stat)
		default:
			clog.Error("RoomMap:proc() unexpected cmd %v", msg.cmd)
		}
	}
}

func (roomMap *RoomMap) Add(rid string, connWrap *ConnWrap) {
	clog.Info("RoomMap:Add() %s, %+v", rid, connWrap)

	if rid == "" || connWrap.Uid == "" {
		return
	}

	i := utils.Hash33(connWrap.Uid) % roomMap.n
	select {
	case roomMap.chs[i] <- &roomMsg{cmd: ADD, rid: rid, uid: connWrap.Uid, sid: connWrap.Sid, body: connWrap}:
	default:
		clog.Error("RoomMap:Add() chan full")
	}
}

func (roomMap *RoomMap) Del(rid string, connWrap *ConnWrap) {
	clog.Info("RoomMap:Del() %s, %+v", rid, connWrap)

	if rid == "" || connWrap.Uid == "" {
		return
	}

	i := utils.Hash33(connWrap.Uid) % roomMap.n
	select {
	case roomMap.chs[i] <- &roomMsg{cmd: DEL, rid: rid, uid: connWrap.Uid, sid: connWrap.Sid, body: connWrap}:
	default:
		clog.Error("RoomMap:Del() chan full")
	}
}

func (roomMap *RoomMap) Push(msg proto.Msg) {
	clog.Info("RoomMap:Push() %+v", msg)

	var pushExt *comm.PushExt
	if ext := msg.Ext(); ext != "" {
		err := json.Unmarshal([]byte(ext), &pushExt)
		if err != nil {
			clog.Error("RoomMap:Push() json.Unmarshal error: %v", err)
			return
		}
	}
	if pushExt != nil {
		ML.Append(pushExt.MsgId, msg)
	}

	for _, ch := range roomMap.chs {
		select {
		case ch <- &roomMsg{cmd: PUSH, rid: msg.Rid(), body: msg}:
		default:
			clog.Error("RoomMap:Push() chan full")
		}
	}
}

var RM RoomMap

func init() {
	RM.init()
}
