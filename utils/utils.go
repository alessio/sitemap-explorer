package utils

import (
	"net/url"
	"strings"
)

func IsAllowedDomain(allowedDomain string, absoluteURL *url.URL) bool {
	if absoluteURL.Host != allowedDomain {
		return false
	}
	return true
}

func BuildAbsoluteURL(a, b string) (*url.URL, error) {
	aParsed, err := url.Parse(a)
	if err != nil {
		return nil, nil
	}
	bParsed, _ := url.Parse(b)
	if err != nil {
		return nil, nil
	}
	if bParsed.IsAbs() {
		return bParsed, nil
	}
	ret, err := url.Parse(
		aParsed.Scheme +
			"://" + aParsed.Host + "/" +
			strings.TrimLeft(b, "/"))
	if err != nil {
		return bParsed, nil
	}
	return ret, nil
}
