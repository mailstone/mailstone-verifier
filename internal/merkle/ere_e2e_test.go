package merkle

import "testing"

// TestVerifyRealEREProof feeds a REAL ERE proof (captured end-to-end from the
// unified pipeline: the 'deposit' leaf of an anchored envelope) through the
// Verifier's own VerifyHash, in the exact MerkleProofData shape the API now
// emits. It proves the Verifier reconstructs the Merkle root of an ERE
// delivery — i.e. the offline "copy-paste this JSON into MailStone Verifier"
// workflow works for ERE, not just V2.
func TestVerifyRealEREProof(t *testing.T) {
	const (
		leaf = "e9f74b2745ffe15a17b07b77dce175849c9d0b131a883ac93b87472aaed558aa"
		root = "9ac6efcd8118a23e5b5deda51198da617764e52128263ceb61e390c5a670364f"
	)
	data := &MerkleProofData{
		RootHash:  root,
		LeafCount: 5,
		Leaves: []LeafData{{
			EntityType: "ere",
			LeafHash:   leaf,
			LeafIndex:  0,
			MerklePath: []MerklePathNode{
				{Level: 0, Position: "RIGHT", SiblingHash: "af531c8592e0918b81a8ed11fae8db3f77147c034040f67b13ecebcdfd086d73"},
				{Level: 1, Position: "RIGHT", SiblingHash: "054280f3fa276b1648d93f6e74745077c7770e49086c36230e6d9bab12297235"},
			},
		}},
	}
	res, err := VerifyHash(leaf, data)
	if err != nil {
		t.Fatalf("VerifyHash: %v", err)
	}
	if !res.Success {
		t.Fatalf("verification failed: calculated=%s expected=%s (%s)", res.CalculatedRoot, res.ExpectedRoot, res.Error)
	}
	if res.CalculatedRoot != root {
		t.Fatalf("reconstructed root mismatch: got %s want %s", res.CalculatedRoot, root)
	}
}
