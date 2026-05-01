package hasher

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
)

// CalculateSHA256 calculates the SHA-256 hash of a file and returns it as a hex string
func CalculateSHA256(filePath string) (string, error) {
	// Open file
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// Create SHA-256 hasher
	hasher := sha256.New()

	// Copy file content to hasher
	_, err = io.Copy(hasher, file)
	if err != nil {
		return "", err
	}

	// Get hash sum and convert to hex
	hashBytes := hasher.Sum(nil)
	hashHex := hex.EncodeToString(hashBytes)

	return hashHex, nil
}
