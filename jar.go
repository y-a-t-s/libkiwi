package libkiwi

import (
	"fmt"
	"net/http"
	"net/url"
	"sync"
)

type cookieMap map[string]map[string]*http.Cookie

type KiwiJar struct {
	cookieMap
	mutex *sync.Mutex
}

func newCookieMap() cookieMap {
	return make(cookieMap, 2)
}

func NewKiwiJar(domain *url.URL, cookies string) (kj *KiwiJar, err error) {
	kj = &KiwiJar{
		cookieMap: newCookieMap(),
		mutex:     &sync.Mutex{},
	}
	kj.newDomain(domain)

	if cookies == "" {
		return
	}

	cs, err := parseCookieString(cookies)
	if err != nil {
		return
	}
	kj.SetCookies(domain, cs)

	return
}

func (kj *KiwiJar) checkAlloc(u *url.URL) {
	if kj.cookieMap == nil || kj.cookieMap[u.Host] == nil {
		kj.newDomain(u)
	}
}

func (kj *KiwiJar) SetCookies(u *url.URL, cookies []*http.Cookie) {
	kj.checkAlloc(u)

	for _, c := range cookies {
		kj.SetCookie(u, c)
	}
}

func (kj *KiwiJar) Cookies(u *url.URL) []*http.Cookie {
	kj.checkAlloc(u)

	res := make(chan []*http.Cookie, 1)

	go func() {
		kj.mutex.Lock()
		defer kj.mutex.Unlock()

		cl := len(kj.cookieMap[u.Host])
		cs := make([]*http.Cookie, cl)
		i := 0
		for _, c := range kj.cookieMap[u.Host] {
			if i >= cl {
				break
			}

			cs[i] = c
			i++
		}

		res <- cs
	}()

	return <-res
}

func (kj *KiwiJar) CookieString(u *url.URL) (cookies string) {
	cs := kj.Cookies(u)
	for _, c := range cs {
		cookies += fmt.Sprintf("; %s=%s", c.Name, c.Value)
	}
	// Remove leading semicolon+space.
	cookies = cookies[2:]

	return
}

func (kj *KiwiJar) GetCookie(u *url.URL, name string) *http.Cookie {
	kj.checkAlloc(u)

	res := make(chan *http.Cookie, 1)

	go func() {
		kj.mutex.Lock()
		defer kj.mutex.Unlock()

		res <- kj.cookieMap[u.Host][name]
	}()

	return <-res
}

func (kj *KiwiJar) SetCookie(u *url.URL, cookie *http.Cookie) {
	kj.checkAlloc(u)

	done := make(chan bool, 1)

	go func() {
		kj.mutex.Lock()
		defer kj.mutex.Unlock()

		kj.cookieMap[u.Host][cookie.Name] = cookie
		done <- true
	}()

	<-done
}

func (kj *KiwiJar) newDomain(domain *url.URL) {
	if kj.cookieMap == nil {
		kj.cookieMap = newCookieMap()
	}

	done := make(chan bool, 1)

	go func() {
		kj.mutex.Lock()
		defer kj.mutex.Unlock()

		kj.cookieMap[domain.Host] = make(map[string]*http.Cookie, 16)
		done <- true
	}()

	<-done
}
