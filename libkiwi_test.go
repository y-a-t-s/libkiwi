package libkiwi

import (
	"log"
	"net/http"
	"net/url"
	"os"
	"testing"
)

const TEST_HOST = "kiwifarms.st"

func TestGetPage(t *testing.T) {
	cookies := os.Getenv("TEST_COOKIES")
	kf, err := NewKF(http.Client{}, TEST_HOST, cookies)
	if err != nil {
		t.Error(err)
	}
	log.Println("Getting homepage")
	resp, err := kf.GetPage(kf.domain)
	if err != nil {
		t.Error(err)
	}
	defer resp.Body.Close()

	log.Printf("Response status code: %d\n", resp.StatusCode)
	for k, v := range resp.Header {
		if len(v) > 0 {
			log.Printf("%s: %s\n", k, v[0])
		}
	}
}

func TestRefreshSession(t *testing.T) {
	cookies := os.Getenv("TEST_COOKIES")
	kf, err := NewKF(http.Client{}, TEST_HOST, cookies)
	if err != nil {
		t.Error(err)
	}
	log.Println("Refreshing xf_session")
	tk, err := kf.RefreshSession()
	if err != nil {
		t.Error(err)
	}
	log.Println("New xf_session token: " + tk)
}

func TestCookieString(t *testing.T) {
	cookies := os.Getenv("TEST_COOKIES")
	kf, err := NewKF(http.Client{}, TEST_HOST, cookies)
	if err != nil {
		t.Error(err)
	}
	u, err := url.Parse("https://" + TEST_HOST)
	if err != nil {
		t.Error(err)
	}
	log.Println("Cookies from jar: " + kf.Client.Jar.(*KiwiJar).CookieString(u))
}
