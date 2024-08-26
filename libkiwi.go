package libkiwi

import (
	"errors"
	"net/http"
	"net/url"
	"regexp"

	"github.com/y-a-t-s/firebird"
)

type KF struct {
	client http.Client
	domain *url.URL
}

// Supply your own http.Client to route through any proxies.
func NewKF(hc http.Client, host string, cookies string) (kf *KF, err error) {
	_, host, err = splitProtocol(host)
	if err != nil {
		return
	}
	u, err := url.Parse("https://" + host)
	if err != nil {
		return
	}

	jar, err := NewKiwiJar(u, cookies)
	if err != nil {
		return
	}
	hc.Jar = jar

	kf = &KF{
		client: hc,
		domain: u,
	}

	return
}

func (kf *KF) GetPage(u *url.URL) (resp *http.Response, err error) {
	if u == nil {
		err = errors.New("Received nil URL.")
		return
	}

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return
	}
	// req.Header.Set("Cookie", kf.cookies)

	resp, err = kf.client.Do(req)
	if err != nil {
		return
	}

	// KiwiFlare redirect is signaled by 203 status.
	if resp.StatusCode == 203 {
		err = kf.solveKiwiFlare()
		if err != nil {
			return
		}
		// Try fetching the page again now that we're authed.
		return kf.GetPage(u)
	}

	return
}

func (kf *KF) RefreshSession() (tk string, err error) {
	// Clear any existing session token to request a new one.
	kf.client.Jar.(*KiwiJar).SetCookie(kf.domain, &http.Cookie{
		Name:  "xf_session",
		Value: "",
	})

	resp, err := kf.GetPage(kf.domain)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	tk = regexp.MustCompile(`xf_session=([^;]*)`).FindString(resp.Header.Get("Set-Cookie"))
	return
}

func (kf *KF) solveKiwiFlare() error {
	c, err := firebird.NewChallenge(kf.client, kf.domain.String())
	if err != nil {
		return err
	}
	s, err := firebird.Solve(c)
	if err != nil {
		return err
	}
	_, err = firebird.Submit(kf.client, kf.domain.String(), s)
	if err != nil {
		return err
	}

	return nil
}
