package merkle

import "testing"

// TestVerifyLegacyFullTreeProof proves the Verifier still reconstructs a proof
// carried in the LEGACY V2 "full tree" shape — where data.leaves holds EVERY
// leaf of the batch, not just the one being verified. Clients who already hold
// a full-tree V2 Proof PDF must keep verifying it forever, even after new
// proofs move to the leaner single-leaf inclusion-path shape (like ERE).
//
// VerifyHash searches data.leaves for the hash under test, then walks THAT
// leaf's merkle_path — so the same code handles 1 leaf (new) or N leaves
// (legacy). This locks that contract: never assume leaves[0].
func TestVerifyLegacyFullTreeProof(t *testing.T) {
	const root = "9ac6efcd8118a23e5b5deda51198da617764e52128263ceb61e390c5a670364f"
	// A real 4-leaf batch (captured end-to-end), rendered as a legacy full tree.
	data := &MerkleProofData{
		RootHash:  root,
		LeafCount: 4,
		Leaves: []LeafData{
			{LeafIndex: 0, LeafHash: "e9f74b2745ffe15a17b07b77dce175849c9d0b131a883ac93b87472aaed558aa", MerklePath: []MerklePathNode{
				{Level: 0, Position: "RIGHT", SiblingHash: "af531c8592e0918b81a8ed11fae8db3f77147c034040f67b13ecebcdfd086d73"},
				{Level: 1, Position: "RIGHT", SiblingHash: "054280f3fa276b1648d93f6e74745077c7770e49086c36230e6d9bab12297235"},
			}},
			{LeafIndex: 1, LeafHash: "af531c8592e0918b81a8ed11fae8db3f77147c034040f67b13ecebcdfd086d73", MerklePath: []MerklePathNode{
				{Level: 0, Position: "LEFT", SiblingHash: "e9f74b2745ffe15a17b07b77dce175849c9d0b131a883ac93b87472aaed558aa"},
				{Level: 1, Position: "RIGHT", SiblingHash: "054280f3fa276b1648d93f6e74745077c7770e49086c36230e6d9bab12297235"},
			}},
			{LeafIndex: 2, LeafHash: "c4ce25f5c9ce3582f7466634cbe8c0ff7a50428ae1c73c690be05f8d8e6074de", MerklePath: []MerklePathNode{
				{Level: 0, Position: "RIGHT", SiblingHash: "f1dbaaa3b9e64db2008a4f3228fd187d59e5e7d4c67fe44576357945b89a6bf5"},
				{Level: 1, Position: "LEFT", SiblingHash: "596450d57f743733edfc3dcb9fd4153b3daa2a105d14b68040a02751221c8dd2"},
			}},
			{LeafIndex: 3, LeafHash: "f1dbaaa3b9e64db2008a4f3228fd187d59e5e7d4c67fe44576357945b89a6bf5", MerklePath: []MerklePathNode{
				{Level: 0, Position: "LEFT", SiblingHash: "c4ce25f5c9ce3582f7466634cbe8c0ff7a50428ae1c73c690be05f8d8e6074de"},
				{Level: 1, Position: "LEFT", SiblingHash: "596450d57f743733edfc3dcb9fd4153b3daa2a105d14b68040a02751221c8dd2"},
			}},
		},
	}

	// Any entity in the batch must verify from the same full-tree blob.
	for _, tc := range []struct {
		name string
		hash string
	}{
		{"first leaf", "e9f74b2745ffe15a17b07b77dce175849c9d0b131a883ac93b87472aaed558aa"},
		{"middle leaf", "c4ce25f5c9ce3582f7466634cbe8c0ff7a50428ae1c73c690be05f8d8e6074de"},
		{"last leaf", "f1dbaaa3b9e64db2008a4f3228fd187d59e5e7d4c67fe44576357945b89a6bf5"},
	} {
		res, err := VerifyHash(tc.hash, data)
		if err != nil {
			t.Fatalf("%s: VerifyHash: %v", tc.name, err)
		}
		if !res.Success || res.CalculatedRoot != root {
			t.Fatalf("%s: legacy full-tree verification failed: calc=%s want=%s (%s)", tc.name, res.CalculatedRoot, root, res.Error)
		}
	}
}
