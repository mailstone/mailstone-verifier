package timestamp

import (
	"crypto"
	"crypto/x509"
	"encoding/asn1"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
	"time"

	ts "github.com/digitorus/timestamp"
	smallstepPKCS7 "github.com/smallstep/pkcs7"
)

// DecodeTimestampResult represents the decoded timestamp information
type DecodeTimestampResult struct {
	Provider        string `json:"provider"`
	DateTime        string `json:"dateTime"`
	SerialNumber    string `json:"serialNumber"`
	HashAlgo        string `json:"hashAlgo"`
	TimestampedHash string `json:"timestampedHash"`
	Status          string `json:"status"`
	// SignatureVerified is true when the token's PKCS#7 signature checks
	// out against the TSA certificate embedded in the token. It says the
	// token is intact and was issued by the holder of that certificate; it
	// does NOT, by itself, say that certificate belongs to a trusted
	// authority (see ChainNote).
	SignatureVerified bool   `json:"signatureVerified"`
	SignerSubject     string `json:"signerSubject"`
	SignerIssuer      string `json:"signerIssuer"`
	SignerValidFrom   string `json:"signerValidFrom"`
	SignerValidTo     string `json:"signerValidTo"`
	ChainNote         string `json:"chainNote"`
	// CoversHash is set when the caller passed the hash the token is
	// expected to cover (a Merkle root, a file hash): "yes", "no", or ""
	// when nothing was asked.
	CoversHash string `json:"coversHash"`
	Error      string `json:"error,omitempty"`
}

// chainNote is printed with every decoded token: this tool checks the
// token's own signature; trusting the signer is a separate, policy question.
const chainNote = "Signature checked against the certificate embedded in the token. Whether that certificate is a trusted timestamping authority is for you to confirm (its subject and issuer are shown)."

// Covers reports whether the token's message imprint equals hashHex. The
// comparison is what links a timestamp to a batch: a token over a Merkle
// root dates every leaf of that batch.
func (r *DecodeTimestampResult) Covers(hashHex string) {
	want := strings.ToLower(strings.TrimSpace(hashHex))
	if want == "" {
		r.CoversHash = ""
		return
	}
	if strings.ToLower(r.TimestampedHash) == want {
		r.CoversHash = "yes"
	} else {
		r.CoversHash = "no"
	}
}

// CleanBase64 removes whitespace and newlines from base64 string
func CleanBase64(base64Str string) string {
	base64Str = strings.ReplaceAll(base64Str, "\n", "")
	base64Str = strings.ReplaceAll(base64Str, "\r", "")
	base64Str = strings.ReplaceAll(base64Str, " ", "")
	base64Str = strings.ReplaceAll(base64Str, "\t", "")
	return base64Str
}

// ParseTimestamp parses a TSA timestamp response (RFC 3161)
// Supports both standard RSA signatures and RSASSA-PSS (used by MailStone TimeStamp)
func ParseTimestamp(tsrBytes []byte) (*DecodeTimestampResult, error) {
	// Try digitorus/timestamp first (handles FreeTSA, InternalTSA, Unataca)
	tsResp, err := ts.ParseResponse(tsrBytes)
	if err == nil {
		return buildResultFromDigitorus(tsResp), nil
	}

	// Fallback: use smallstep/pkcs7 for RSASSA-PSS responses (MailStone TimeStamp)
	result, fallbackErr := parseWithSmallstep(tsrBytes)
	if fallbackErr != nil {
		return nil, fmt.Errorf("failed to parse timestamp response: %w (fallback: %v)", err, fallbackErr)
	}

	return result, nil
}

// buildResultFromDigitorus builds the result from a digitorus/timestamp parsed response
func buildResultFromDigitorus(tsResp *ts.Timestamp) *DecodeTimestampResult {
	serialNumber := "N/A"
	if tsResp.SerialNumber != nil {
		serialNumber = tsResp.SerialNumber.String()
	}

	timestampedHash := "N/A"
	if len(tsResp.HashedMessage) > 0 {
		timestampedHash = hex.EncodeToString(tsResp.HashedMessage)
	}

	out := &DecodeTimestampResult{
		Provider:        extractTSAProvider(tsResp),
		DateTime:        tsResp.Time.Format("2006-01-02 15:04:05 MST"),
		SerialNumber:    serialNumber,
		HashAlgo:        getHashAlgoName(tsResp.HashAlgorithm),
		TimestampedHash: timestampedHash,
		Status:          "GRANTED",
		// digitorus/timestamp verifies the PKCS#7 signature when the token
		// embeds its certificate (Parse fails otherwise); a token without
		// certificate is parsed but cannot be checked.
		SignatureVerified: tsResp.AddTSACertificate,
		ChainNote:         chainNote,
	}
	if !tsResp.AddTSACertificate {
		out.ChainNote = "The token embeds no certificate: its signature could not be checked."
	}
	if len(tsResp.Certificates) > 0 {
		fillSigner(out, tsResp.Certificates[0])
	}
	return out
}

// fillSigner copies the signer certificate's identity and validity window.
func fillSigner(out *DecodeTimestampResult, cert *x509.Certificate) {
	out.SignerSubject = cert.Subject.String()
	out.SignerIssuer = cert.Issuer.String()
	out.SignerValidFrom = cert.NotBefore.UTC().Format("2006-01-02")
	out.SignerValidTo = cert.NotAfter.UTC().Format("2006-01-02")
}

// parseWithSmallstep parses a TSA response using smallstep/pkcs7 (supports RSASSA-PSS)
func parseWithSmallstep(tsrBytes []byte) (*DecodeTimestampResult, error) {
	// Parse outer TimeStampResp: SEQUENCE { PKIStatusInfo, TimeStampToken }
	var outer asn1.RawValue
	_, err := asn1.Unmarshal(tsrBytes, &outer)
	if err != nil {
		return nil, fmt.Errorf("failed to parse TimeStampResp: %w", err)
	}

	// Skip PKIStatusInfo, get TimeStampToken bytes
	var statusSeq asn1.RawValue
	tokenBytes, err := asn1.Unmarshal(outer.Bytes, &statusSeq)
	if err != nil {
		return nil, fmt.Errorf("failed to parse PKIStatusInfo: %w", err)
	}

	// Check status
	var status int
	if _, err := asn1.Unmarshal(statusSeq.Bytes, &status); err != nil {
		return nil, fmt.Errorf("failed to parse PKIStatus: %w", err)
	}
	if status > 1 {
		return nil, fmt.Errorf("TSA returned non-granted status: %d", status)
	}

	if len(tokenBytes) == 0 {
		return nil, fmt.Errorf("TimeStampToken is missing")
	}

	// Parse PKCS7 with smallstep (supports RSASSA-PSS)
	p7, err := smallstepPKCS7.Parse(tokenBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse PKCS7 token: %w", err)
	}

	if len(p7.Content) == 0 {
		return nil, fmt.Errorf("PKCS7 content (TSTInfo) is empty")
	}

	// Parse TSTInfo from PKCS7 content
	var tst tstInfo
	if _, err := asn1.Unmarshal(p7.Content, &tst); err != nil {
		return nil, fmt.Errorf("failed to parse TSTInfo: %w", err)
	}

	// Extract serial number
	serialNumber := "N/A"
	if tst.SerialNumber != nil {
		serialNumber = tst.SerialNumber.String()
	}

	// Extract timestamped hash from MessageImprint
	hashAlgo, timestampedHash := parseMessageImprint(tst.MessageImprint)

	// Extract generation time
	dateTime := tst.GenTime.Format("2006-01-02 15:04:05 MST")

	// Extract provider from certificate
	provider := "Unknown TSA"
	if len(p7.Certificates) > 0 {
		cert := p7.Certificates[0]
		if len(cert.Subject.Organization) > 0 {
			provider = cert.Subject.Organization[0]
		} else if cert.Subject.CommonName != "" {
			provider = cert.Subject.CommonName
		}
	}

	out := &DecodeTimestampResult{
		Provider:        provider,
		DateTime:        dateTime,
		SerialNumber:    serialNumber,
		HashAlgo:        hashAlgo,
		TimestampedHash: timestampedHash,
		Status:          "GRANTED",
		ChainNote:       chainNote,
	}
	// This path used to decode without checking anything: an RSASSA-PSS
	// token with a forged signature decoded as cleanly as a genuine one.
	if len(p7.Certificates) == 0 {
		out.ChainNote = "The token embeds no certificate: its signature could not be checked."
	} else if verr := p7.Verify(); verr != nil {
		out.ChainNote = "Signature check FAILED: " + verr.Error()
	} else {
		out.SignatureVerified = true
	}
	if len(p7.Certificates) > 0 {
		fillSigner(out, p7.Certificates[0])
	}
	return out, nil
}

// tstInfo is the RFC 3161 TSTInfo structure
type tstInfo struct {
	Version        int
	Policy         asn1.ObjectIdentifier
	MessageImprint asn1.RawValue
	SerialNumber   *big.Int
	GenTime        time.Time `asn1:"generalized"`
}

// messageImprint is the MessageImprint structure in TSTInfo
type messageImprint struct {
	HashAlgorithm struct {
		Algorithm asn1.ObjectIdentifier
	}
	HashedMessage []byte
}

// parseMessageImprint extracts the hash algorithm name and hash value from a MessageImprint
func parseMessageImprint(raw asn1.RawValue) (string, string) {
	var mi messageImprint
	if _, err := asn1.Unmarshal(raw.FullBytes, &mi); err != nil {
		return "Unknown", "N/A"
	}

	// Map OID to algorithm name
	algoName := mi.HashAlgorithm.Algorithm.String()
	knownOIDs := map[string]string{
		"2.16.840.1.101.3.4.2.1": "SHA-256",
		"2.16.840.1.101.3.4.2.2": "SHA-384",
		"2.16.840.1.101.3.4.2.3": "SHA-512",
		"1.3.14.3.2.26":          "SHA-1",
	}
	if name, ok := knownOIDs[algoName]; ok {
		algoName = name
	}

	return algoName, hex.EncodeToString(mi.HashedMessage)
}

// getHashAlgoName returns the name of the hash algorithm from crypto.Hash
func getHashAlgoName(algo crypto.Hash) string {
	names := map[crypto.Hash]string{
		crypto.SHA1:   "SHA-1",
		crypto.SHA256: "SHA-256",
		crypto.SHA384: "SHA-384",
		crypto.SHA512: "SHA-512",
	}
	if name, ok := names[algo]; ok {
		return name
	}
	return fmt.Sprintf("%v", algo)
}

// extractTSAProvider extracts the TSA provider name from a digitorus/timestamp response
func extractTSAProvider(tsResp *ts.Timestamp) string {
	if len(tsResp.Certificates) > 0 {
		cert := tsResp.Certificates[0]
		if len(cert.Subject.Organization) > 0 {
			return cert.Subject.Organization[0]
		}
		if len(cert.Subject.OrganizationalUnit) > 0 {
			return cert.Subject.OrganizationalUnit[0]
		}
		if cert.Subject.CommonName != "" {
			return cert.Subject.CommonName
		}
	}

	if tsResp.Policy.String() != "" {
		knownPolicies := map[string]string{
			"1.3.6.1.4.1.6449.1.2.1.3.1":     "FreeTSA",
			"1.3.6.1.4.1.57916.1.1.1.1.1":     "MailStone TimeStamp",
			"0.4.0.2023.1.1":                   "Unataca",
		}
		if provider, ok := knownPolicies[tsResp.Policy.String()]; ok {
			return provider
		}
	}

	return "Unknown TSA"
}
