package main

import (
	"strings"
	"testing"
)

// About shows the licence and the third-party notices embedded at build time.
func TestLegalIsEmbedded(t *testing.T) {
	a := &App{}
	if !strings.Contains(a.GetLicense(), "MIT License") {
		t.Error("licence not embedded")
	}
	for _, want := range []string{"# Third-party components — MailStone Verifier", "github.com/wailsapp/wails/v2", "### Go standard library"} {
		if !strings.Contains(a.GetThirdPartyNotices(), want) {
			t.Errorf("third-party notices without %q", want)
		}
	}
}
