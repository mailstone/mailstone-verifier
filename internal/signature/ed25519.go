// Package signature verifies the recipient's decision on a MailStone
// Registered Electronic Delivery (ERE).
//
// When the recipient accepts or refuses a delivery, their browser signs a
// small canonical text with their Ed25519 private key — a key that never
// leaves their device. The proof document prints the public key, the
// signature and that exact text. Anyone can then check, without MailStone,
// that the holder of that key signed exactly that decision for exactly that
// delivery: Ed25519 is a signature scheme, not a hash — the verifier gets
// the message, the public key and the signature, and answers true or false.
package signature

import (
	"crypto/ed25519"
	"encoding/base64"
	"fmt"
	"strings"
	"time"
)

// DecisionMessage builds the canonical bytes the recipient signed. It must
// match, byte for byte, the API's buildDecisionPayload and the portal's
// Decide screen: four lines, each ending with a line feed, decided_at in
// RFC 3339 UTC with no fractional seconds.
func DecisionMessage(ereID, decision string, decidedAt time.Time) string {
	return "ere/decision/v1\n" +
		"ere_id=" + strings.TrimSpace(ereID) + "\n" +
		"decision=" + strings.TrimSpace(decision) + "\n" +
		"decided_at=" + decidedAt.UTC().Format(time.RFC3339) + "\n"
}

// NormalizeMessage takes the block as pasted from the proof document — any
// line endings, surrounding whitespace, trailing LF present or not — and
// returns the exact signed bytes. The document says every line ends with a
// line feed, the last one included; a reader who copies the four lines
// without the final one must not get "false" for that reason alone.
func NormalizeMessage(pasted string) string {
	s := strings.ReplaceAll(pasted, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	lines := strings.Split(strings.TrimSpace(s), "\n")
	for i := range lines {
		lines[i] = strings.TrimSpace(lines[i])
	}
	return strings.Join(lines, "\n") + "\n"
}

// cleanBase64 removes every whitespace character: a key or signature copied
// from a PDF arrives with the line break (or a space) the page inserted
// where the long string wrapped.
func cleanBase64(s string) string {
	return strings.Map(func(r rune) rune {
		if r == ' ' || r == '\n' || r == '\r' || r == '\t' || r == '\u00a0' {
			return -1
		}
		return r
	}, s)
}

// Result is what the verifier reports.
type Result struct {
	Valid     bool   `json:"valid"`
	Message   string `json:"message"`   // the exact bytes verified, for display
	PublicKey string `json:"publicKey"` // base64, as given
	Error     string `json:"error,omitempty"`
}

// Verify checks sig (base64, 64 bytes) over message with publicKey (base64,
// 32 bytes). A malformed key or signature is an error; a well-formed
// signature that does not match is simply Valid=false.
func Verify(publicKeyB64, signatureB64, message string) (*Result, error) {
	pk, err := base64.StdEncoding.DecodeString(cleanBase64(publicKeyB64))
	if err != nil {
		return nil, fmt.Errorf("public key is not valid base64: %w", err)
	}
	if len(pk) != ed25519.PublicKeySize {
		return nil, fmt.Errorf("public key must be %d bytes, got %d", ed25519.PublicKeySize, len(pk))
	}
	sig, err := base64.StdEncoding.DecodeString(cleanBase64(signatureB64))
	if err != nil {
		return nil, fmt.Errorf("signature is not valid base64: %w", err)
	}
	if len(sig) != ed25519.SignatureSize {
		return nil, fmt.Errorf("signature must be %d bytes, got %d", ed25519.SignatureSize, len(sig))
	}
	return &Result{
		Valid:     ed25519.Verify(ed25519.PublicKey(pk), []byte(message), sig),
		Message:   message,
		PublicKey: cleanBase64(publicKeyB64),
	}, nil
}
