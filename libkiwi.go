package libkiwi

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"regexp"

	"github.com/y-a-t-s/firebird"
)

type KF struct {
	Client http.Client
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
		Client: hc,
		domain: u,
	}

	return
}

func (kf *KF) GetPage(ctx context.Context, u *url.URL) (resp *http.Response, err error) {
	if u == nil {
		err = errors.New("Received nil URL.")
		return
	}

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return
	}

	resp, err = kf.Client.Do(req)
	if err != nil {
		return
	}

	// KiwiFlare redirect is signaled by 203 status.
	if resp.StatusCode == 203 {
		err = kf.solveKiwiFlare(ctx)
		if err != nil {
			return
		}
		// Try fetching the page again now that we're authed.
		return kf.GetPage(ctx, u)
	}

	return
}

func (kf *KF) RefreshSession(ctx context.Context) (tk string, err error) {
	// Clear any existing session token to request a new one.
	kf.Client.Jar.(*KiwiJar).SetCookie(kf.domain, &http.Cookie{
		Name:  "xf_session",
		Value: "",
	})

	resp, err := kf.GetPage(ctx, kf.domain)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	tk = regexp.MustCompile(`xf_session=([^;]*)`).FindString(resp.Header.Get("Set-Cookie"))
	return
}

func (kf *KF) solveKiwiFlare(ctx context.Context) error {
	c, err := firebird.NewChallenge(kf.Client, kf.domain.String())
	if err != nil {
		return err
	}
	s, err := firebird.Solve(ctx, c)
	if err != nil {
		return err
	}
	_, err = firebird.Submit(kf.Client, kf.domain.String(), s)
	if err != nil {
		return err
	}

	return nil
}
