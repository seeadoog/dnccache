package dnscache

import (
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)
var DefaultDnsCache = NewDnsCache(20*time.Second)


type DnsCache struct {
	cache         sync.Map   //ip 地址缓存
	freshInterval time.Duration // dns 刷新时间
	timer         *time.Ticker
	httpClient    *http.Client
}

func (d *DnsCache)DialFunc()func(network, addr string) (net.Conn, error){
	return d.dialFunc
}

func (d *DnsCache) dialFunc(network, addr string) (net.Conn, error) {
	ips := strings.Split(addr, ":")
	if len(ips) != 2 {
		return nil, fmt.Errorf("invaild addr:%s", addr)
	}
	ip, err := d.Get(ips[0])
	if err != nil {
		return nil, fmt.Errorf("resolve ip error:%w", err)
	}
	return net.Dial("tcp4", fmt.Sprintf("%s:%s", ip, ips[1]))
}

func NewDnsCache(freshInterval time.Duration) *DnsCache {
	c := &DnsCache{
		freshInterval: freshInterval,
	}
	c.timer = time.NewTicker(freshInterval)
	c.httpClient = &http.Client{
		Transport: &http.Transport{
			Dial:    c.dialFunc,
			DialTLS: c.dialFunc,
		},
	}
	go c.freshDns()
	return c
}

func (d *DnsCache) resolveIp(host string) (string, error) {
	ip, err := net.ResolveIPAddr("ip4", host)
	if err != nil {
		return "", fmt.Errorf("resolve ip of %s error:%w", host, err)
	}
	ips := ip.IP.String()
	d.cache.Store(host, ips)
	return ips, nil
}

func (d *DnsCache) Get(host string) (string, error) {
	r, ok := d.cache.Load(host)
	if ok {
		return r.(string), nil
	}
	return d.resolveIp(host)
}

func (d *DnsCache) Clear(host string) {
	d.cache.Delete(host)
}

func (d *DnsCache) Destroy() {
	d.timer.Stop()
}

func (d *DnsCache) freshDns() {
	for _ = range d.timer.C {
		// 定时刷新
		d.cache.Range(func(key, value interface{}) bool {
			host := key.(string)
			ip, err := d.resolveIp(host)
			if err != nil {
				return true
			}
			d.cache.Store(host, ip)
			return true
		})
	}
}

func (d *DnsCache) DoHttpRequest(req *http.Request) (*http.Response,error){
	return d.httpClient.Do(req)
}

