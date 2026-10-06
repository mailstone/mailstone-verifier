package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"mailstone-verifier/internal/evidence"
	"mailstone-verifier/internal/hasher"
	"mailstone-verifier/internal/merkle"
	"mailstone-verifier/internal/signature"
	"mailstone-verifier/internal/timestamp"
	"os"
	"path/filepath"
	goruntime "runtime"
	"strings"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// SelectFile opens a file dialog and returns the selected file path
func (a *App) SelectFile() (string, error) {
	filePath, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select file to hash",
		Filters: []runtime.FileFilter{
			{
				DisplayName: "All Files",
				Pattern:     "*.*",
			},
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to open file dialog: %w", err)
	}
	return filePath, nil
}

// CalculateHash calculates SHA-256 hash of a file
func (a *App) CalculateHash(filePath string) (string, error) {
	hash, err := hasher.CalculateSHA256(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to calculate hash: %w", err)
	}
	return hash, nil
}

// DecodeTimestamp decodes a Base64 RFC 3161 timestamp token and checks its
// signature. expectedHash (optional) is the hash the token should cover — a
// Merkle root from a proof block, a file hash — and yields CoversHash.
func (a *App) DecodeTimestamp(base64Token string, expectedHash string) (*timestamp.DecodeTimestampResult, error) {
	// Remove whitespace and newlines
	base64Token = timestamp.CleanBase64(base64Token)

	// Decode base64
	tsrBytes, err := base64.StdEncoding.DecodeString(base64Token)
	if err != nil {
		return &timestamp.DecodeTimestampResult{
			Error: "Invalid Base64 encoding: " + err.Error(),
		}, nil
	}

	// Parse timestamp
	tsInfo, err := timestamp.ParseTimestamp(tsrBytes)
	if err != nil {
		return &timestamp.DecodeTimestampResult{
			Error: "Failed to parse timestamp: " + err.Error(),
		}, nil
	}
	tsInfo.Covers(expectedHash)
	return tsInfo, nil
}

// MerkleVerifyRequest represents the input for Merkle verification
type MerkleVerifyRequest struct {
	MerkleJSON string `json:"merkleJson"`
	Hash       string `json:"hash"`
}

// VerifyMerkle verifies that a hash is in the Merkle tree and reconstructs the root.
// With an empty Hash, the block's own leaf is verified — the ERE proof prints
// one single-leaf block per event, and the reader has nothing else to type.
func (a *App) VerifyMerkle(request MerkleVerifyRequest) (*merkle.MerkleVerifyResult, error) {
	// Parse JSON
	var merkleData merkle.MerkleProofData
	err := json.Unmarshal([]byte(request.MerkleJSON), &merkleData)
	if err != nil {
		return &merkle.MerkleVerifyResult{
			Success: false,
			Error:   "Invalid JSON format: " + err.Error(),
		}, nil
	}
	hash := strings.TrimSpace(request.Hash)
	if hash == "" {
		if len(merkleData.Leaves) != 1 {
			return &merkle.MerkleVerifyResult{
				Success: false,
				Error:   "This block lists several leaves: enter the hash you want to verify.",
			}, nil
		}
		hash = merkleData.Leaves[0].LeafHash
	}

	// Verify the hash
	result, err := merkle.VerifyHash(hash, &merkleData)
	if err != nil {
		return &merkle.MerkleVerifyResult{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return result, nil
}

// EreDecisionRequest is the input for the recipient-decision check. Either
// the pasted canonical block (Message) or its three fields are given; the
// fields win when both are present.
type EreDecisionRequest struct {
	PublicKey string `json:"publicKey"` // base64 Ed25519
	Signature string `json:"signature"` // base64
	Message   string `json:"message"`   // the four-line block as printed
	EreID     string `json:"ereId"`
	Decision  string `json:"decision"`  // accepted | refused
	DecidedAt string `json:"decidedAt"` // RFC 3339
}

// VerifyEreDecision checks the recipient's Ed25519 signature on a delivery
// decision. Valid means: the holder of this public key signed exactly this
// decision, for exactly this delivery, at exactly this time.
func (a *App) VerifyEreDecision(req EreDecisionRequest) (*signature.Result, error) {
	msg := ""
	switch {
	case strings.TrimSpace(req.EreID) != "" && strings.TrimSpace(req.Decision) != "" && strings.TrimSpace(req.DecidedAt) != "":
		at, err := time.Parse(time.RFC3339, strings.TrimSpace(req.DecidedAt))
		if err != nil {
			return &signature.Result{Error: "decided_at must be RFC 3339, e.g. 2026-10-05T09:36:36Z"}, nil
		}
		msg = signature.DecisionMessage(req.EreID, req.Decision, at)
	case strings.TrimSpace(req.Message) != "":
		msg = signature.NormalizeMessage(req.Message)
	default:
		return &signature.Result{Error: "Paste the signed block, or fill in ere_id, decision and decided_at."}, nil
	}
	res, err := signature.Verify(req.PublicKey, req.Signature, msg)
	if err != nil {
		return &signature.Result{Error: err.Error(), Message: msg}, nil
	}
	return res, nil
}

// EreEvidenceRequest is the input for an evidence-hash recomputation: the
// stage and the facts the proof document prints for it.
type EreEvidenceRequest struct {
	Stage  string            `json:"stage"`
	Fields map[string]string `json:"fields"`
}

// ComputeEreEvidenceHash rebuilds the leaf hash anchored for one event of a
// delivery from the facts shown in the proof. The result is the hash to
// verify in the Merkle tab against that event's proof block.
func (a *App) ComputeEreEvidenceHash(req EreEvidenceRequest) (*evidence.Result, error) {
	res, err := evidence.Compute(req.Stage, req.Fields)
	if err != nil {
		return &evidence.Result{Stage: req.Stage, Error: err.Error()}, nil
	}
	return res, nil
}

// appVersion is printed in the report header (keep in step with wails.json).
const appVersion = "2.1.0"

// AppInfo heads an exported report.
type AppInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	OS      string `json:"os"`
	Arch    string `json:"arch"`
}

// GetAppInfo returns the tool's name, version and platform.
func (a *App) GetAppInfo() AppInfo {
	return AppInfo{Name: "MailStone Verifier", Version: appVersion, OS: goruntime.GOOS, Arch: goruntime.GOARCH}
}

// FileSize is the size of a file, for the report (0 when unknown).
func (a *App) FileSize(path string) (int64, error) {
	st, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return st.Size(), nil
}

// ExportRequest is a report to save: the page renders it (text in the
// interface language, or JSON), Go only asks where and writes it.
type ExportRequest struct {
	Format      string `json:"format"` // "txt" | "json"
	Content     string `json:"content"`
	DefaultName string `json:"defaultName"`
}

// ExportReport opens the save dialog and writes the report. Returns the
// path written, or "" when the reader cancelled.
func (a *App) ExportReport(req ExportRequest) (string, error) {
	format := req.Format
	if format != "json" {
		format = "txt"
	}
	filter := runtime.FileFilter{DisplayName: "Text report (*.txt)", Pattern: "*.txt"}
	if format == "json" {
		filter = runtime.FileFilter{DisplayName: "JSON data (*.json)", Pattern: "*.json"}
	}
	path, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "Save the verification report",
		DefaultFilename: req.DefaultName,
		Filters:         []runtime.FileFilter{filter},
	})
	if err != nil {
		return "", fmt.Errorf("save dialog: %w", err)
	}
	if path == "" {
		return "", nil
	}
	if filepath.Ext(path) == "" {
		path += "." + format
	}
	if err := os.WriteFile(path, []byte(req.Content), 0o644); err != nil {
		return "", fmt.Errorf("write report: %w", err)
	}
	return path, nil
}

// ConfirmDialog asks a yes/no question with the platform's native dialog.
func (a *App) ConfirmDialog(title, message string) (bool, error) {
	res, err := runtime.MessageDialog(a.ctx, runtime.MessageDialogOptions{
		Type:          runtime.QuestionDialog,
		Title:         title,
		Message:       message,
		Buttons:       []string{"Yes", "No"},
		DefaultButton: "No",
		CancelButton:  "No",
	})
	if err != nil {
		return false, err
	}
	return res == "Yes" || res == "Ok", nil
}
