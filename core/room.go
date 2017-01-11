package core

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/simplejia/clog"
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
	body interface{}
}

type RoomMap struct {
	n   int
	chs []chan *roomMsg
}

func (roomMap *RoomMap) init() {
	roomMap.n = conf.C.Cons.U_MAP_NUM
	for i := 0; i < roomMap.n; i++ {
		roomMap.chs = append(roomMap.chs, make(chan *roomMsg, conf.C.Cons.ROOM_MSG_LEN))
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
			connWrap := msg.body.(*ConnWrap)
			rids_m, ok := data[rid]
			if !ok {
				rids_m = map[[2]string]*ConnWrap{}
				data[rid] = rids_m
			}
			connWrap.Rids = append(connWrap.Rids, rid)
			for _, _rid := range connWrap.Rids[:len(connWrap.Rids)-1] {
				if _rid == rid {
					connWrap.Rids = connWrap.Rids[:len(connWrap.Rids)-1]
					break
				}
			}

			ukey := [2]string{connWrap.Uid, connWrap.Sid}
			rids_m[ukey] = connWrap
		case DEL:
			rid := msg.rid
			connWrap := msg.body.(*ConnWrap)
			rids_m, ok := data[rid]
			if !ok {
				break
			}

			ukey := [2]string{connWrap.Uid, connWrap.Sid}
			if v, ok := rids_m[ukey]; !ok || v.C != connWrap.C {
				break
			}

			delete(rids_m, ukey)
			if len(rids_m) == 0 {
				delete(data, rid)
			}
			for i, _rid := range connWrap.Rids {
				if _rid == rid {
					connWrap.Rids = append(connWrap.Rids[:i], connWrap.Rids[i+1:]...)
					break
				}
			}
		case PUSH:
			rid := msg.rid
			m := msg.body.(proto.Msg)
			rids_m, ok := data[rid]
			if !ok || len(rids_m) == 0 {
				break
			}

			var pushExt *comm.PushExt
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

			servExt_bs, _ := json.Marshal(servExt)

			btime := time.Now()
			ukey_ex := [2]string{m.Uid(), m.Sid()}
			for ukey, connWrap := range rids_m {
				if ukey_ex[1] == "" { // 当后端没有传入sid时，只匹配uid
					if ukey_ex[0] == ukey[0] {
						continue
					}
				} else {
					if ukey_ex == ukey {
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

				msg.SetExt(string(servExt_bs))
				connWrap.Write(msg)
			}
			etime := time.Now()
			stat, _ := json.Marshal(&comm.Stat{
				Ip:    utils.LocalIp,
				N:     i,
				Rid:   rid,
				Msg:   fmt.Sprintf("%+v", m),
				Num:   len(rids_m),
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
	case roomMap.chs[i] <- &roomMsg{cmd: ADD, rid: rid, body: connWrap}:
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
	case roomMap.chs[i] <- &roomMsg{cmd: DEL, rid: rid, body: connWrap}:
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
