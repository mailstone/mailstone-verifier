package timestamp

import (
	"encoding/base64"
	"os"
	"testing"
)

// A real root token issued by the MailStone internal TSA over the Merkle
// root of an anchored batch (decision stage of delivery 0fce8041…, 2026-09-29
// 13:01:23 UTC). The token embeds its certificate, so the signature is
// checked, and its imprint must equal the batch root.
const batchRoot = "75cb922f2c3f5c88a0bc6230c21ba757b2f855bbdd83cb0f4d9a84728b29e8eb"

func loadToken(t *testing.T) []byte {
	t.Helper()
	b64, err := os.ReadFile("testdata/internaltsa-root-token.b64")
	if err != nil {
		t.Fatalf("fixture: %v", err)
	}
	raw, err := base64.StdEncoding.DecodeString(CleanBase64(string(b64)))
	if err != nil {
		t.Fatalf("fixture base64: %v", err)
	}
	return raw
}

func TestRealRootTokenDecodesAndVerifies(t *testing.T) {
	r, err := ParseTimestamp(loadToken(t))
	if err != nil {
		t.Fatalf("ParseTimestamp: %v", err)
	}
	if r.Status != "GRANTED" || r.HashAlgo != "SHA-256" {
		t.Fatalf("status=%s algo=%s", r.Status, r.HashAlgo)
	}
	if !r.SignatureVerified {
		t.Fatalf("the signature of a genuine token must verify (%s)", r.ChainNote)
	}
	if r.SignerSubject == "" || r.SignerValidTo == "" {
		t.Fatal("signer identity must be reported")
	}
	if r.DateTime[:10] != "2026-09-29" || r.SerialNumber != "42" {
		t.Fatalf("decoded %s serial %s", r.DateTime, r.SerialNumber)
	}
}

func TestTokenCoversTheBatchRoot(t *testing.T) {
	r, err := ParseTimestamp(loadToken(t))
	if err != nil {
		t.Fatalf("ParseTimestamp: %v", err)
	}
	r.Covers(batchRoot)
	if r.CoversHash != "yes" {
		t.Fatalf("imprint %s must equal the batch root", r.TimestampedHash)
	}
	r.Covers("0000" + batchRoot[4:])
	if r.CoversHash != "no" {
		t.Fatal("a different hash must not be reported as covered")
	}
	r.Covers("")
	if r.CoversHash != "" {
		t.Fatal("no expected hash, no verdict")
	}
}

func TestTamperedTokenIsRejected(t *testing.T) {
	raw := loadToken(t)
	// Flip a byte deep inside the signed content.
	raw[len(raw)/2] ^= 0x01
	r, err := ParseTimestamp(raw)
	if err == nil && r.SignatureVerified {
		t.Fatal("a tampered token must not verify")
	}
}
