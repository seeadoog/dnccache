package dnscache

import (
	"fmt"
	"github.com/gorilla/websocket"
	"io/ioutil"
	"net/http"
	"testing"
)

func TestNewDnsCache(t *testing.T) {
	//使用http 请求
	req, err := http.NewRequest("GET", "https://ws-api.xfyun.cn:80/v2/tts", nil)
	if err != nil {
		panic(err)
	}

	rsp, err := DefaultDnsCache.DoHttpRequest(req)
	if err != nil {
		panic(err)
	}

	b, _ := ioutil.ReadAll(rsp.Body)
	fmt.Println(string(b))

	// 使用websocket
	d := websocket.Dialer{
		NetDial: DefaultDnsCache.DialFunc(),
	}

	d.Dial("wss://ws-api.xfyun.cn/v2/iat", nil)

}
