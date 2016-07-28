package utils

import (
	"net/url"
	"testing"
)

func TestIsAllowedDomain(t *testing.T) {
	u, _ := url.Parse("https://golang.org/pkg/")
	ret := IsAllowedDomain("golang.org", u)
	if !ret {
		t.Error("Expected 'true', got %q", ret)
	}
}

func TestIsAllowedDomainFail(t *testing.T) {
	v, _ := url.Parse("golang.org/pkg/")
	ret := IsAllowedDomain("golang.org", v)
	if ret {
		t.Error("Expected 'false', got %q", ret)
	}

}

func TestBuildAbsoluteURL(t *testing.T) {
	v, _ := BuildAbsoluteURL("http://golang.org", "file.txt")
	if !v.IsAbs() {
		t.Error("Expected 'true', got %q", v.IsAbs())
	}
	if v.String() != "http://golang.org/file.txt" {
		t.Error("Expected 'http://golang.org/file.txt', got %q", v.String())
	}
}
