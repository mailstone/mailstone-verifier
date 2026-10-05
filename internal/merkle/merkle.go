package merkle

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

// MerkleProofData represents the JSON structure from the proof
type MerkleProofData struct {
	Blockchain struct {
		Ledger   string `json:"ledger"`
		Network  string `json:"network"`
		Provider string `json:"provider"`
		TxHash   string `json:"tx_hash"`
	} `json:"blockchain"`
	LeafCount int        `json:"leaf_count"`
	Leaves    []LeafData `json:"leaves"`
	RootHash  string     `json:"root_hash"`
}

// LeafData represents a single leaf in the Merkle tree
type LeafData struct {
	EntityID   string           `json:"entity_id"`
	EntityType string           `json:"entity_type"`
	LeafHash   string           `json:"leaf_hash"`
	LeafIndex  int              `json:"leaf_index"`
	MerklePath []MerklePathNode `json:"merkle_path"`
	Filename   string           `json:"filename,omitempty"`
}

// MerklePathNode represents one step in the Merkle path
type MerklePathNode struct {
	Level       int    `json:"level"`
	Position    string `json:"position"`
	SiblingHash string `json:"sibling_hash"`
}

// MerkleVerifyResult represents the result of Merkle verification
type MerkleVerifyResult struct {
	Success         bool     `json:"success"`
	CalculatedRoot  string   `json:"calculatedRoot"`
	ExpectedRoot    string   `json:"expectedRoot"`
	LeafIndex       int      `json:"leafIndex"`
	TotalLeaves     int      `json:"totalLeaves"`
	LeafType        string   `json:"leafType"`
	MerklePathSteps []string `json:"merklePathSteps"`
	Error           string   `json:"error,omitempty"`
}

// VerifyHash verifies that a hash exists in the Merkle tree and reconstructs the root
func VerifyHash(hashToVerify string, data *MerkleProofData) (*MerkleVerifyResult, error) {
	// Normalize hash (lowercase, remove whitespace)
	hashToVerify = strings.ToLower(strings.TrimSpace(hashToVerify))

	// Find the leaf with this hash
	var foundLeaf *LeafData
	for i := range data.Leaves {
		if strings.ToLower(data.Leaves[i].LeafHash) == hashToVerify {
			foundLeaf = &data.Leaves[i]
			break
		}
	}

	if foundLeaf == nil {
		return &MerkleVerifyResult{
			Success: false,
			Error:   "Hash not found in any leaf of the Merkle tree",
		}, nil
	}

	// Reconstruct Merkle root from this leaf
	calculatedRoot, steps, err := reconstructMerkleRoot(hashToVerify, foundLeaf.MerklePath)
	if err != nil {
		return &MerkleVerifyResult{
			Success: false,
			Error:   fmt.Sprintf("Failed to reconstruct Merkle root: %v", err),
		}, nil
	}

	// Compare with expected root
	expectedRoot := strings.ToLower(strings.TrimSpace(data.RootHash))
	success := calculatedRoot == expectedRoot

	return &MerkleVerifyResult{
		Success:         success,
		CalculatedRoot:  calculatedRoot,
		ExpectedRoot:    expectedRoot,
		LeafIndex:       foundLeaf.LeafIndex,
		TotalLeaves:     data.LeafCount,
		LeafType:        foundLeaf.EntityType,
		MerklePathSteps: steps,
	}, nil
}

// reconstructMerkleRoot reconstructs the Merkle root by following the path from leaf to root
func reconstructMerkleRoot(leafHash string, path []MerklePathNode) (string, []string, error) {
	currentHash := strings.ToLower(leafHash)
	steps := []string{fmt.Sprintf("Step 0: Leaf hash = %s", currentHash)}

	// Follow the path upwards
	// Position indicates where the SIBLING is located (LEFT or RIGHT of current node)
	for i, node := range path {
		siblingHash := strings.ToLower(node.SiblingHash)

		var newHash string
		var err error

		if node.Position == "LEFT" {
			// Sibling is on the left, current is on the right
			newHash, err = hashPair(siblingHash, currentHash)
			steps = append(steps, fmt.Sprintf("Step %d: SHA256(%s || %s) = %s (sibling LEFT)", i+1, siblingHash[:16]+"...", currentHash[:16]+"...", newHash))
		} else {
			// Sibling is on the right, current is on the left
			newHash, err = hashPair(currentHash, siblingHash)
			steps = append(steps, fmt.Sprintf("Step %d: SHA256(%s || %s) = %s (sibling RIGHT)", i+1, currentHash[:16]+"...", siblingHash[:16]+"...", newHash))
		}

		if err != nil {
			return "", nil, err
		}

		currentHash = newHash
	}

	steps = append(steps, fmt.Sprintf("Final root: %s", currentHash))

	return currentHash, steps, nil
}

// hashPair hashes two hex-encoded hashes together using SHA-256
func hashPair(leftHash, rightHash string) (string, error) {
	// Decode hex strings to bytes
	leftBytes, err := hex.DecodeString(leftHash)
	if err != nil {
		return "", fmt.Errorf("failed to decode left hash: %w", err)
	}

	rightBytes, err := hex.DecodeString(rightHash)
	if err != nil {
		return "", fmt.Errorf("failed to decode right hash: %w", err)
	}

	// Concatenate and hash
	hasher := sha256.New()
	hasher.Write(leftBytes)
	hasher.Write(rightBytes)

	result := hasher.Sum(nil)
	return hex.EncodeToString(result), nil
}
