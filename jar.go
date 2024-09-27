package libkiwi

import (
	"fmt"
	"net/http"
	"net/url"
	"sync"
)

type cookieMap map[string]map[string]*http.Cookie

// An http cookiejar implementation that doesn't suck ass.
type KiwiJar struct {
	cookieMap
	mutex sync.Mutex

	init func()
}

func NewKiwiJar() *KiwiJar {
	kj := new(KiwiJar)
	kj.init = sync.OnceFunc(func() {
		kj.cookieMap = make(cookieMap, 2)
	})

	return kj
}

func (kj *KiwiJar) Cookies(u *url.URL) []*http.Cookie {
	kj.newDomain(u)

	hn := u.Hostname()
	res := make(chan []*http.Cookie, 1)

	go func() {
		kj.mutex.Lock()
		defer kj.mutex.Unlock()

		cs := make([]*http.Cookie, 0, len(kj.cookieMap[hn]))
		for _, c := range kj.cookieMap[hn] {
			cs = append(cs, c)
		}

		res <- cs
	}()

	return <-res
}

func (kj *KiwiJar) ParseString(u *url.URL, cookies string) error {
	if cookies == "" {
		return nil
	}

	cs, err := parseCookieString(cookies)
	if err != nil {
		return err
	}

	kj.init()
	kj.SetCookies(u, cs)

	return nil
}

func (kj *KiwiJar) CookieString(u *url.URL) (cookies string) {
	cs := kj.Cookies(u)
	for _, c := range cs {
		cookies += fmt.Sprintf("; %s=%s", c.Name, c.Value)
	}
	if len(cookies) > 2 {
		// Remove leading semicolon+space.
		cookies = cookies[2:]
	}

	return
}

func (kj *KiwiJar) GetCookie(u *url.URL, name string) *http.Cookie {
	kj.newDomain(u)

	res := make(chan *http.Cookie, 1)

	go func() {
		kj.mutex.Lock()
		defer kj.mutex.Unlock()

		res <- kj.cookieMap[u.Hostname()][name]
	}()

	return <-res
}

func (kj *KiwiJar) set(u *url.URL, cookie *http.Cookie) {
	kj.mutex.Lock()
	defer kj.mutex.Unlock()

	kj.cookieMap[u.Hostname()][cookie.Name] = cookie
}

func (kj *KiwiJar) SetCookie(u *url.URL, cookie *http.Cookie) {
	kj.newDomain(u)

	done := make(chan bool, 1)

	go func() {
		defer close(done)
		kj.set(u, cookie)
	}()

	<-done
}

func (kj *KiwiJar) SetCookies(u *url.URL, cookies []*http.Cookie) {
	kj.newDomain(u)

	var wg sync.WaitGroup

	for _, c := range cookies {
		wg.Add(1)
		go func() {
			defer wg.Done()
			kj.set(u, c)
		}()
	}

	wg.Wait()
}

func (kj *KiwiJar) newDomain(u *url.URL) {
	kj.init()
	if kj.cookieMap[u.Hostname()] != nil {
		return
	}

	done := make(chan bool, 1)

	go func() {
		defer close(done)

		kj.mutex.Lock()
		defer kj.mutex.Unlock()

		kj.cookieMap[u.Hostname()] = make(map[string]*http.Cookie, 16)
	}()

	<-done
}
