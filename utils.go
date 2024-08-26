package libkiwi

import (
	"errors"
	"net/http"
	"regexp"
	"strings"
)

func parseCookieString(cookies string) ([]*http.Cookie, error) {
	sp := strings.Split(cookies, "; ")
	cs := make([]*http.Cookie, len(sp))

	for i, c := range sp {
		kv := strings.Split(c, "=")
		if len(kv) != 2 {
			return nil, errors.New("Invalid cookie string: " + cookies)
		}
		cs[i] = &http.Cookie{
			Name:  kv[0],
			Value: kv[1],
		}
	}

	return cs, nil
}

func splitProtocol(addr string) (proto string, host string, err error) {
	// FindStringSubmatch is used to capture the groups.
	// Index 0 is the full matching string with all groups.
	// The rest are numbered by the order of the opening parens.
	// Here, we want the last 2 groups (indexes 1 and 2, requiring length 3).
	tmp := regexp.MustCompile(`^([\w-]+://)?([^/]+)`).FindStringSubmatch(addr)
	// At the very least, we need the hostname part (index 2).
	if len(tmp) < 3 || tmp[2] == "" {
		err = errors.New("Failed to parse address: " + addr)
		return
	}

	proto = tmp[1]
	host = tmp[2]

	return
}
