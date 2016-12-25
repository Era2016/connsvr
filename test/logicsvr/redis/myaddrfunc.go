package redis

import "fmt"

func init() {
	// 请赋值成自己的根据addrType, addr返回ip:port的函数
	addrFunc = func(addrType, addr string) (string, error) {
		switch addrType {
		case "ip":
			return addr, nil
		default:
			panic(fmt.Sprintf("Unexpected addrtype: %v", addrType))
		}
	}
}
