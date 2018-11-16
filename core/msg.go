package core

import (
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/simplejia/clog/api"
	"github.com/simplejia/connsvr/comm"
	"github.com/simplejia/connsvr/conf"
	"github.com/simplejia/connsvr/proto"
	"github.com/simplejia/lc"
	"github.com/simplejia/utils"
)

type MsgElem struct {
	id   string
	body string
	uid  string
	sid  string
}

type MsgList []*MsgElem

func (a MsgList) Len() int           { return len(a) }
func (a MsgList) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a MsgList) Less(i, j int) bool { return a[i].id < a[j].id }

func (a MsgList) Key4Lc(msg proto.Msg) string {
	return fmt.Sprintf("conn:msgs:%v:%v", msg.Rid(), msg.Subcmd())
}

func (a MsgList) GetLc(msg proto.Msg) (MsgList, bool) {
	key_lc := a.Key4Lc(msg)
	a_lc, ok := lc.Get(key_lc)
	if a_lc != nil {
		a = a_lc.(MsgList)
	} else {
		a = nil
	}
	return a, ok
}

func (a MsgList) SetLc(msg proto.Msg) {
	key_lc := a.Key4Lc(msg)
	lc.Set(key_lc, a, time.Minute)
}

func (a MsgList) Append(id string, msg proto.Msg) MsgList {
	a, _ = a.GetLc(msg)

	x := &MsgElem{
		id:   id,
		body: msg.Body(),
		uid:  msg.Uid(),
		sid:  msg.Sid(),
	}
	i := sort.Search(len(a), func(i int) bool { return a[i].id >= id })
	if i == len(a) {
		a = append(a, x)
	} else if a[i].id == id {
		a[i] = x
	} else {
		a = append(a[:i], append([]*MsgElem{x}, a[i:]...)...)
	}

	n := conf.V.Get().MsgNum
	if len(a) > n {
		copy(a, a[len(a)-n:])
		a = a[:n]
	}

	sort.Sort(a)
	a.SetLc(msg)

	return a
}

// 请赋值成自己的根据addrType, addr返回ip:port的函数
var MsgAddrFunc = func(addrType, addr string) (string, error) {
	return addr, nil
}

func (a MsgList) Bodys(id string, msg proto.Msg) (bodys []string) {
	a, ok := a.GetLc(msg)

	// 当connsvr缓存消息为空时，路由到后端服务拉取数据
	if len(a) == 0 && !ok {
		subcmd := strconv.Itoa(int(msg.Subcmd()))
		c := conf.C.Msgs[subcmd]
		if c == nil {
			clog.Error("MsgList:Bodys() no expected subcmd: %s", subcmd)
			return
		}
		addr, err := MsgAddrFunc(c.AddrType, c.Addr)
		if err != nil {
			clog.Error("MsgList:Bodys() MsgAddrFunc error: %v", err)
			return
		}
		arrs := []string{
			strconv.Itoa(int(msg.Cmd())),
			subcmd,
			msg.Uid(),
			msg.Sid(),
			msg.Rid(),
		}
		ps := map[string]string{}
		values, _ := url.ParseQuery(fmt.Sprintf(c.Params, utils.Slice2Interface(arrs)...))
		for k, vs := range values {
			ps[k] = vs[0]
		}

		timeout, _ := time.ParseDuration(c.Timeout)

		headers := map[string]string{
			"Host": c.Host,
		}

		uri := fmt.Sprintf("http://%s/%s", addr, strings.TrimPrefix(c.Cgi, "/"))

		gpp := &utils.GPP{
			Uri:     uri,
			Timeout: timeout,
			Headers: headers,
			Params:  ps,
		}

		body, err := utils.Get(gpp)
		if err != nil {
			clog.Error("MsgList:utils.Get() http error, err: %v, body: %s, gpp: %v", err, body, gpp)
			return
		}
		clog.Debug("MsgList:utils.Get() http success, body: %s, gpp: %v", body, gpp)

		var ms []*comm.Msg
		err = json.Unmarshal(body, &ms)
		if err != nil {
			clog.Error("MsgList:json.Unmarshal() error, err: %v, body: %s", err, body)
			return
		}

		for _, m := range ms {
			_msg := proto.NewMsg(comm.SVR)
			_msg.SetSubcmd(msg.Subcmd())
			_msg.SetRid(msg.Rid())
			_msg.SetUid(m.Uid)
			_msg.SetSid(m.Sid)
			_msg.SetBody(m.Body)
			a = a.Append(m.MsgId, _msg)
		}

		// 当后端也没有数据时，放一个空数据，避免下次再次拉取
		if len(a) == 0 {
			_msg := proto.NewMsg(comm.SVR)
			_msg.SetSubcmd(msg.Subcmd())
			_msg.SetRid(msg.Rid())
			a = a.Append("", _msg)
		}
	}

	i := sort.Search(len(a), func(i int) bool { return a[i].id > id })
	for _, e := range a[i:] {
		// 过滤掉自己的消息，但当客户端传入id为空时（客户端无缓存消息），不用过滤
		if id != "" {
			if e.sid == "" { // 当后端没有传入sid时，只匹配uid
				if e.uid == msg.Uid() {
					continue
				}
			} else {
				if e.uid == msg.Uid() && e.sid == msg.Sid() {
					continue
				}
			}
		}
		bodys = append(bodys, e.body)
	}
	return
}

var ML MsgList
