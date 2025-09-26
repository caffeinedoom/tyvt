package rotator

import (
	"net/http"
	"net/url"
	"sync"
)

type IPRotator struct {
	mu           sync.RWMutex
	proxies      []string
	currentIndex int
}

func NewIPRotator(proxies []string) *IPRotator {
	return &IPRotator{
		proxies:      proxies,
		currentIndex: 0,
	}
}

func (ir *IPRotator) ProxyFunc() func(*http.Request) (*url.URL, error) {
	if len(ir.proxies) == 0 {
		return http.ProxyFromEnvironment
	}

	return func(req *http.Request) (*url.URL, error) {
		ir.mu.Lock()
		defer ir.mu.Unlock()

		if len(ir.proxies) == 0 {
			return nil, nil
		}

		proxy := ir.proxies[ir.currentIndex]
		ir.currentIndex = (ir.currentIndex + 1) % len(ir.proxies)

		return url.Parse(proxy)
	}
}

func (ir *IPRotator) CurrentProxy() string {
	ir.mu.RLock()
	defer ir.mu.RUnlock()

	if len(ir.proxies) == 0 {
		return ""
	}

	return ir.proxies[ir.currentIndex]
}

func (ir *IPRotator) AddProxy(proxy string) {
	ir.mu.Lock()
	defer ir.mu.Unlock()
	ir.proxies = append(ir.proxies, proxy)
}

func (ir *IPRotator) GetProxyCount() int {
	ir.mu.RLock()
	defer ir.mu.RUnlock()
	return len(ir.proxies)
}