package libkiwi

import (
	"context"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/y-a-t-s/firebird"
	"github.com/y-a-t-s/kiwijar"
)

type KF struct {
	Client http.Client
	domain *url.URL
}

// Supply your own http.Client to route through any proxies.
func NewKF(hc http.Client, host string, cookies string) (kf *KF, err error) {
	u, err := parseHost(host)
	if err != nil {
		return
	}

	jar := kiwijar.KiwiJar{}
	jar.ParseString(u, cookies)
	hc.Jar = &jar

	kf = &KF{
		Client: hc,
		domain: u,
	}

	// Update host url in case we get redirected across domains.
	hc.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		hn := req.URL.Hostname()
		if hn != u.Hostname() {
			// Deliberately set to Hostname() and not Host.
			// This excludes any extra shit like ports.
			u.Host = hn
		}

		return nil
	}

	return
}

func (kf *KF) GetPage(ctx context.Context, u *url.URL) (resp *http.Response, err error) {
	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return
	}

	resp, err = kf.Client.Do(req)
	if err != nil {
		return
	}
	hn := resp.Request.URL.Hostname()
	if hn != kf.domain.Hostname() {
		jar := kf.Client.Jar.(*kiwijar.KiwiJar)
		jar.SetCookies(resp.Request.URL, jar.Cookies(kf.domain))
		kf.domain.Host = hn
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
	kf.Client.Jar.(*kiwijar.KiwiJar).SetCookie(kf.domain, &http.Cookie{
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
	_, err = firebird.Submit(kf.Client, s)
	if err != nil {
		return err
	}

	return nil
}

func parseHost(host string) (*url.URL, error) {
	// Try prepending protocol if it seems to be missing.
	if !strings.Contains(host, "://") {
		host = "https://" + host
	}

	return url.Parse(host)
}
