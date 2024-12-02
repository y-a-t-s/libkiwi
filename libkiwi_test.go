package libkiwi

import (
	"context"
	"log"
	"net/http"
	"os"
	"testing"

	"github.com/y-a-t-s/kiwijar"
)

const TEST_HOST = "kiwifarms.net"

func TestGetPage(t *testing.T) {
	cookies := os.Getenv("TEST_COOKIES")
	kf, err := NewKF(http.Client{}, TEST_HOST, cookies)
	if err != nil {
		t.Error(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log.Println("Getting homepage")
	resp, err := kf.GetPage(ctx, kf.domain)
	if err != nil {
		t.Error(err)
	}
	defer resp.Body.Close()

	log.Printf("Response status code: %d\n\n", resp.StatusCode)
	for k, v := range resp.Header {
		if len(v) > 0 {
			log.Printf("%s: %s\n", k, v[0])
		}
	}
	log.Printf("Response host: %s\n\n", kf.domain)
	log.Printf("Cookies: %s\n", kf.Client.Jar.(*kiwijar.KiwiJar).CookieString(kf.domain))
}

func TestRefreshSession(t *testing.T) {
	cookies := os.Getenv("TEST_COOKIES")
	kf, err := NewKF(http.Client{}, TEST_HOST, cookies)
	if err != nil {
		t.Error(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log.Println("Refreshing xf_session")
	tk, err := kf.RefreshSession(ctx)
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

	log.Println("Cookies from jar: " + kf.Client.Jar.(*kiwijar.KiwiJar).CookieString(kf.domain))
}
