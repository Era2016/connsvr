package conf

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/simplejia/connsvr/comm"
	"github.com/simplejia/utils"
)

type Conf struct {
	App *struct {
		Name  string
		Tport int
		Hport int
		Wport int
		Bport int
	}
	Cons *struct {
		MAX_ROOM_NUM int // 一个用户最多加入的房间数
		U_MAP_NUM    int // 用户分组hash
		ROOM_MSG_LEN int // 房间消息队列长度
		C_RBUF       int // 读缓冲区大小
		C_WBUF       int // 写缓冲区大小，越小，越可能写超时
	}
	Pubs map[string]*struct {
		Addr     string
		AddrType string
		Retry    int
		Host     string
		Cgi      string
		Params   string
		Method   string
		Timeout  string
	}
	Msgs map[string]*struct {
		Addr     string
		AddrType string
		Host     string
		Cgi      string
		Params   string
		Timeout  string
	}
	Vars *struct {
		PushKind         comm.PUSH_KIND            // 消息推送方式
		RoomWithPushKind map[string]comm.PUSH_KIND // 消息推送方式, 房间单独配置
		MsgNum           int                       // connsvr服务缓存消息最大长度
	}
	Clog *struct {
		Mode  int
		Level int
	}
}

var (
	Envs map[string]*Conf
	Env  string
	C    *Conf
)

func init() {
	flag.StringVar(&Env, "env", "prod", "set env")
	flag.Parse()

	dir := "conf"
	for i := 0; i < 3; i++ {
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			break
		}
		dir = filepath.Join("..", dir)
	}
	fcontent, err := ioutil.ReadFile(filepath.Join(dir, "conf.json"))
	if err != nil {
		panic(err)
	}

	fcontent = utils.RemoveAnnotation(fcontent)
	if err := json.Unmarshal(fcontent, &Envs); err != nil {
		panic(err)
	}

	C = Envs[Env]
	if C == nil {
		fmt.Println("env not right:", Env)
		os.Exit(-1)
	}

	fmt.Printf("Env: %s\nC: %s\n", Env, utils.Iprint(C))
	return
}
