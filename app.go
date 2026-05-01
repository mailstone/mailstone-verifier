package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"mailstone-verifier/internal/hasher"
	"mailstone-verifier/internal/merkle"
	"mailstone-verifier/internal/timestamp"

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

// DecodeTimestamp decodes a Base64 RFC 3161 timestamp token
func (a *App) DecodeTimestamp(base64Token string) (*timestamp.DecodeTimestampResult, error) {
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

	return tsInfo, nil
}

// MerkleVerifyRequest represents the input for Merkle verification
type MerkleVerifyRequest struct {
	MerkleJSON string `json:"merkleJson"`
	Hash       string `json:"hash"`
}

// VerifyMerkle verifies that a hash is in the Merkle tree and reconstructs the root
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

	// Verify the hash
	result, err := merkle.VerifyHash(request.Hash, &merkleData)
	if err != nil {
		return &merkle.MerkleVerifyResult{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return result, nil
}
