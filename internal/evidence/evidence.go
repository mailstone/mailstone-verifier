// Package evidence recomputes the leaf hashes MailStone anchors for each
// event of a Registered Electronic Delivery (ERE).
//
// Every event of a delivery — deposit, content, dispatch, each presentation
// to the recipient, the decision — is folded into a Merkle batch as its own
// leaf. The leaf is not an arbitrary value: it is the SHA-256 of a small
// canonical text that commits to the facts of the event (which delivery,
// who, when, which provider message…). The proof document prints those
// facts; this package rebuilds the exact text and hashes it, so a reader
// can confirm that the anchored leaf really is the one committing to the
// facts shown — and then verify that leaf's inclusion in the Merkle tab.
//
// The canonical formats below mirror the platform byte for byte
// (shared/pkg/ereoutbox and the API's receipt hashes). Changing one here
// without the other would make every proof unverifiable; the versioned
// domain prefixes ("ere-deposit:v1", …) exist so a future format can live
// next to the old one.
package evidence

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// uuidRe is the textual UUID the platform prints (lowercase hex, four dashes).
var uuidRe = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

// Stage names, as the proof document labels its event cards.
const (
	StageDeposit      = "deposit"      // récépissé de dépôt
	StageContent      = "content"      // the Email PDF itself (leaf = its SHA-256)
	StageEmission     = "emission"     // hand-over to the sending service
	StageDelivery     = "delivery"     // first presentation
	StagePresentation = "presentation" // a later presentation (reminder)
	StageDecision     = "decision"     // recipient accepted / refused
	StageAbort        = "abort"        // sender cancelled
	StageExpiry       = "expiry"       // no decision in time
)

// Field names accepted by Compute, per stage. Times are RFC 3339 (UTC
// recommended; any zone is normalised), with fractional seconds where the
// platform recorded them.
const (
	FieldEREID          = "ere_id"
	FieldSenderEmail    = "sender_email"
	FieldRecipientEmail = "recipient_email"
	FieldSubject        = "subject"
	FieldContentHash    = "content_hash" // hex SHA-256 of the Email PDF
	FieldMessageID      = "provider_message_id"
	FieldSubmittedAt    = "submitted_at"
	FieldDeliveredAt    = "delivered_at"
	FieldPresentedAt    = "presented_at"
	FieldOrdinal        = "ordinal"
	FieldSignature      = "signature" // base64 Ed25519 signature of the decision
	FieldReceivedAt     = "received_at"
	FieldSenderUserID   = "sender_user_id"
	FieldAbortedAt      = "aborted_at"
	FieldExpiresAt      = "expires_at"
)

// Result carries the recomputed leaf and the exact text that was hashed, so
// the reader sees what the leaf commits to.
type Result struct {
	Stage     string `json:"stage"`
	LeafHash  string `json:"leafHash"`  // hex SHA-256
	Canonical string `json:"canonical"` // the hashed text, printable (signature shown as base64)
	Error     string `json:"error,omitempty"`
}

// Compute rebuilds the leaf hash of one stage from its fields.
func Compute(stage string, fields map[string]string) (*Result, error) {
	// Every value but the subject is an identifier, a digest, a message id or
	// an instant: whitespace inside it can only be the line wrap the proof
	// document inserted where the string was too long for the column. The
	// subject is the one free-text fact and is hashed exactly as printed.
	get := func(name string) string {
		if name == FieldSubject {
			return fields[name]
		}
		return strings.Join(strings.Fields(fields[name]), "")
	}
	need := func(names ...string) error {
		for _, n := range names {
			if get(n) == "" {
				return fmt.Errorf("field %q is required for stage %q", n, stage)
			}
		}
		return nil
	}
	// Identifiers and digests have a fixed shape: a value of the wrong length
	// is a copy that lost a character, and hashing it would only produce a
	// mismatch the reader cannot explain. Refuse it with the reason instead.
	// UUIDs are hashed lowercase, as the platform prints them.
	uuidOf := func(name string) (string, error) {
		v := strings.ToLower(get(name))
		if !uuidRe.MatchString(v) {
			return "", fmt.Errorf("%s must be a UUID of 36 characters (8-4-4-4-12 hex digits), got %d: %q", name, len(v), v)
		}
		return v, nil
	}
	hexOf := func(name string) (string, error) {
		v := strings.ToLower(get(name))
		if len(v) != 64 {
			return "", fmt.Errorf("%s must be 64 hex characters (a SHA-256), got %d", name, len(v))
		}
		if _, err := hex.DecodeString(v); err != nil {
			return "", fmt.Errorf("%s is not hexadecimal", name)
		}
		return v, nil
	}
	var (
		preimage  []byte
		printable string
	)
	switch stage {
	case StageDeposit:
		if err := need(FieldEREID, FieldSenderEmail, FieldRecipientEmail, FieldContentHash); err != nil {
			return nil, err
		}
		id, err := uuidOf(FieldEREID)
		if err != nil {
			return nil, err
		}
		ch, err := hexOf(FieldContentHash)
		if err != nil {
			return nil, err
		}
		printable = strings.Join([]string{"ere-deposit:v1", id, get(FieldSenderEmail), get(FieldRecipientEmail), fields[FieldSubject], ch}, "\n")
		preimage = []byte(printable)

	case StageEmission:
		if err := need(FieldEREID, FieldMessageID, FieldSubmittedAt); err != nil {
			return nil, err
		}
		id, err := uuidOf(FieldEREID)
		if err != nil {
			return nil, err
		}
		at, err := parseTime(get(FieldSubmittedAt))
		if err != nil {
			return nil, err
		}
		printable = strings.Join([]string{"ere-emission:v1", id, get(FieldMessageID), at.Format(time.RFC3339Nano)}, "\n")
		preimage = []byte(printable)

	case StageDelivery:
		if err := need(FieldEREID, FieldDeliveredAt, FieldMessageID); err != nil {
			return nil, err
		}
		id, err := uuidOf(FieldEREID)
		if err != nil {
			return nil, err
		}
		at, err := parseTime(get(FieldDeliveredAt))
		if err != nil {
			return nil, err
		}
		printable = strings.Join([]string{"ere-delivery:v1", id, at.Format(time.RFC3339Nano), get(FieldMessageID)}, "\n")
		preimage = []byte(printable)

	case StagePresentation:
		if err := need(FieldEREID, FieldOrdinal, FieldPresentedAt, FieldMessageID); err != nil {
			return nil, err
		}
		n, err := strconv.Atoi(get(FieldOrdinal))
		if err != nil || n < 2 {
			return nil, fmt.Errorf("ordinal must be an integer ≥ 2 (the first presentation is the 'delivery' stage)")
		}
		id, err := uuidOf(FieldEREID)
		if err != nil {
			return nil, err
		}
		at, err := parseTime(get(FieldPresentedAt))
		if err != nil {
			return nil, err
		}
		printable = strings.Join([]string{"ere-presentation:v1", id, strconv.Itoa(n), at.Format(time.RFC3339Nano), get(FieldMessageID)}, "\n")
		preimage = []byte(printable)

	case StageDecision:
		// The leaf binds the recipient's raw signature AND the server's
		// receipt time, so a signature cannot be back-dated by signing an
		// earlier decided_at.
		if err := need(FieldSignature, FieldReceivedAt); err != nil {
			return nil, err
		}
		sig, err := base64.StdEncoding.DecodeString(get(FieldSignature))
		if err != nil {
			return nil, fmt.Errorf("signature is not valid base64: %w", err)
		}
		at, err := parseTime(get(FieldReceivedAt))
		if err != nil {
			return nil, err
		}
		suffix := "\nreceived_at=" + at.Format(time.RFC3339)
		preimage = append(append([]byte{}, sig...), []byte(suffix)...)
		printable = "<signature bytes, base64: " + get(FieldSignature) + ">" + suffix

	case StageAbort:
		if err := need(FieldEREID, FieldSenderUserID, FieldAbortedAt); err != nil {
			return nil, err
		}
		id, err := uuidOf(FieldEREID)
		if err != nil {
			return nil, err
		}
		user, err := uuidOf(FieldSenderUserID)
		if err != nil {
			return nil, err
		}
		at, err := parseTime(get(FieldAbortedAt))
		if err != nil {
			return nil, err
		}
		printable = "ere/abort/v1\nere_id=" + id + "\nsender_user=" + user + "\naborted_at=" + at.Format(time.RFC3339) + "\n"
		preimage = []byte(printable)

	case StageExpiry:
		if err := need(FieldEREID, FieldExpiresAt); err != nil {
			return nil, err
		}
		id, err := uuidOf(FieldEREID)
		if err != nil {
			return nil, err
		}
		at, err := parseTime(get(FieldExpiresAt))
		if err != nil {
			return nil, err
		}
		printable = id + "|expired|" + at.Format(time.RFC3339)
		preimage = []byte(printable)

	case StageContent:
		// The leaf IS the SHA-256 of the Email PDF: no canonical text, the
		// file's digest is what the batch committed to.
		if err := need(FieldContentHash); err != nil {
			return nil, err
		}
		h, err := hexOf(FieldContentHash)
		if err != nil {
			return nil, err
		}
		return &Result{Stage: stage, LeafHash: h, Canonical: "<the SHA-256 of the Email PDF bytes, no canonical text>"}, nil

	default:
		return nil, fmt.Errorf("unknown stage %q", stage)
	}

	sum := sha256.Sum256(preimage)
	return &Result{Stage: stage, LeafHash: hex.EncodeToString(sum[:]), Canonical: printable}, nil
}

// parseTime accepts RFC 3339 with or without fractional seconds, in any
// zone, and returns the instant in UTC — the platform hashes UTC.
func parseTime(s string) (time.Time, error) {
	for _, layout := range []string{time.RFC3339Nano, time.RFC3339, "2006-01-02 15:04:05.999999999Z07:00", "2006-01-02 15:04:05Z07:00"} {
		if t, err := time.Parse(layout, s); err == nil {
			return t.UTC(), nil
		}
	}
	return time.Time{}, fmt.Errorf("time %q is not RFC 3339 (expected e.g. 2026-10-05T09:17:29Z)", s)
}
