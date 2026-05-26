package client

import "testing"

func TestIsCloudflareBlock(t *testing.T) {
	if !isCloudflareBlock([]byte("Just a moment...")) {
		t.Error("expected true")
	}
	if !isCloudflareBlock([]byte("cf-browser-verify")) {
		t.Error("expected true")
	}
	if !isCloudflareBlock([]byte("Cloudflare")) {
		t.Error("expected true")
	}
	if isCloudflareBlock([]byte("<html><body>HLTV</body></html>")) {
		t.Error("expected false")
	}
}
