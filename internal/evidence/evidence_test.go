package evidence

import (
	"strings"
	"testing"
)

// Every vector below is a real leaf read from a MailStone platform
// (merkle_proofs.leaf_hash) next to the facts the proof document prints for
// that event. Recomputing the leaf from the facts is the whole point.
const ereID = "0fce8041-5a07-4737-9fc0-6f774a64c29f"

func mustCompute(t *testing.T, stage string, fields map[string]string) string {
	t.Helper()
	r, err := Compute(stage, fields)
	if err != nil {
		t.Fatalf("Compute(%s): %v", stage, err)
	}
	return r.LeafHash
}

func TestDepositLeafRecomputedFromTheReceipt(t *testing.T) {
	got := mustCompute(t, StageDeposit, map[string]string{
		FieldEREID:          ereID,
		FieldSenderEmail:    "contact@mailstone.fr",
		FieldRecipientEmail: "p.durand@exemple.fr",
		FieldSubject:        "Révision de nos tarifs au 1er janvier 2027",
		FieldContentHash:    "4e824a1d478d9160b8d2983a724650dc1aedb1b9c8bf1f3e077946547edd89f4",
	})
	if got != "284b32ce4b42b2ffc23c758a47e03dd5eab6aba0feb257dc8eb95f3d22fed474" {
		t.Fatalf("deposit leaf = %s", got)
	}
}

// TestEmissionLeafFormat locks the canonical format of the dispatch leaf.
// No platform vector here, deliberately: the platform hashes the hand-over
// instant at nanosecond precision while its database keeps microseconds,
// so the leaf cannot be recomputed from any stored or printed value today.
// The format below is the platform's; the vector will come once the
// platform truncates the instant before hashing and prints it.
func TestEmissionLeafFormat(t *testing.T) {
	r, err := Compute(StageEmission, map[string]string{
		FieldEREID:       ereID,
		FieldMessageID:   "92c50a67-1531-4b0a-8948-b1b34a6122ec",
		FieldSubmittedAt: "2026-09-29T12:56:03.551849Z",
	})
	if err != nil {
		t.Fatalf("Compute: %v", err)
	}
	want := "ere-emission:v1\n" + ereID + "\n92c50a67-1531-4b0a-8948-b1b34a6122ec\n2026-09-29T12:56:03.551849Z"
	if r.Canonical != want {
		t.Fatalf("canonical = %q, want %q", r.Canonical, want)
	}
	if len(r.LeafHash) != 64 {
		t.Fatalf("leaf hash = %q", r.LeafHash)
	}
}

func TestDeliveryLeafRecomputedFromTheProviderConfirmation(t *testing.T) {
	got := mustCompute(t, StageDelivery, map[string]string{
		FieldEREID:       ereID,
		FieldDeliveredAt: "2026-09-29T12:56:05Z",
		FieldMessageID:   "92c50a67-1531-4b0a-8948-b1b34a6122ec",
	})
	if got != "571067d6b5740d878a4a5f733ae0c02866f7e3624f00cf2c7964c6dbc42191e5" {
		t.Fatalf("delivery leaf = %s", got)
	}
	// Same instant written with a zone: the platform hashes UTC.
	again := mustCompute(t, StageDelivery, map[string]string{
		FieldEREID:       ereID,
		FieldDeliveredAt: "2026-09-29T14:56:05+02:00",
		FieldMessageID:   "92c50a67-1531-4b0a-8948-b1b34a6122ec",
	})
	if again != got {
		t.Fatal("a zoned time must hash like its UTC equivalent")
	}
}

func TestPresentationLeafRecomputed(t *testing.T) {
	got := mustCompute(t, StagePresentation, map[string]string{
		FieldEREID:       "f3354663-57a6-421f-b134-fc485187b4ba",
		FieldOrdinal:     "3",
		FieldPresentedAt: "2026-10-05T06:46:46Z",
		FieldMessageID:   "edb6ad98-bce8-4c1e-ae90-dc45765b1900",
	})
	if got != "0c5a0ee1ce503434de2b40791ad3b6beba6d3876ccf89d362d839a800aaa1d28" {
		t.Fatalf("presentation leaf = %s", got)
	}
}

func TestDecisionLeafBindsSignatureAndReceiptTime(t *testing.T) {
	got := mustCompute(t, StageDecision, map[string]string{
		FieldSignature:  "c2yaX5pLnAmt2QjlScurK2jgMkFZ3GFH2/1Nsv5CslBDmFTJpgg+oRl9Xs/mnruh3xLb8lwIy5OwJAnkJ5rbBQ==",
		FieldReceivedAt: "2026-09-29T13:00:09Z",
	})
	if got != "90054fd6feecec12aa6aa9c2e3651741f8acb325d471890873915fb274fc24fd" {
		t.Fatalf("decision leaf = %s", got)
	}
}

func TestMissingFieldIsAClearError(t *testing.T) {
	if _, err := Compute(StageDeposit, map[string]string{FieldEREID: ereID}); err == nil {
		t.Fatal("missing fields must be reported")
	}
	if _, err := Compute(StageContent, nil); err == nil {
		t.Fatal("the content stage needs the file hash")
	}
	if r, err := Compute(StageContent, map[string]string{FieldContentHash: "4E824A1D478D9160B8D2983A724650DC1AEDB1B9C8BF1F3E077946547EDD89F4"}); err != nil || r.LeafHash != "4e824a1d478d9160b8d2983a724650dc1aedb1b9c8bf1f3e077946547edd89f4" {
		t.Fatalf("the content leaf is the file hash itself, lower-cased (got %v, %v)", r, err)
	}
	if _, err := Compute("teleportation", nil); err == nil {
		t.Fatal("unknown stage must be an error")
	}
}

// A hash copied from the proof's "Envoi" table arrives with the break the
// page inserted where the 64 characters wrapped; the ere_id can too. Real
// staging delivery 493da761…: the leaf anchored on 2026-09-30.
func TestValuesCopiedFromThePageWithWraps(t *testing.T) {
	want := "5dd81caa0b3e82e25a3a37a3cbbd35e1704e32fa7fc0820d143d5c6b38a0a893"
	got := mustCompute(t, StageDeposit, map[string]string{
		FieldEREID:          "493da761-3473-4572-\na9f4-d8f39d92b451",
		FieldSenderEmail:    " jr.quiriconi@mailstone.fr",
		FieldRecipientEmail: "jr@cyberialabs.io\n",
		FieldSubject:        "ere 1",
		FieldContentHash:    "ae0713e023a8d1c20e9902dab814ceb0e8b30603cb06c473b0 229afd63c71a59\n",
	})
	if got != want {
		t.Fatalf("deposit leaf = %s, want %s", got, want)
	}
}

// A deposit of the staging platform (ERE 493da761…, 2026-09-30), read back
// from ere_envelopes and merkle_proofs. The same facts with the ere_id
// missing its last character — a copy that lost one character — must be
// refused with the reason, not hashed into an unexplainable mismatch.
func TestComputeDepositStagingAndTruncatedID(t *testing.T) {
	fields := map[string]string{
		FieldEREID:          "493da761-3473-4572-a9f4-d8f39d92b451",
		FieldSenderEmail:    "jr.quiriconi@mailstone.fr",
		FieldRecipientEmail: "jr@cyberialabs.io",
		FieldSubject:        "ere 1",
		FieldContentHash:    "ae0713e023a8d1c20e9902dab814ceb0e8b30603cb06c473b0229afd63c71a59",
	}
	res, err := Compute(StageDeposit, fields)
	if err != nil {
		t.Fatal(err)
	}
	if want := "5dd81caa0b3e82e25a3a37a3cbbd35e1704e32fa7fc0820d143d5c6b38a0a893"; res.LeafHash != want {
		t.Fatalf("deposit leaf = %s, want %s", res.LeafHash, want)
	}
	// Uppercase UUID, as a reader might retype it: hashed lowercase, same leaf.
	fields[FieldEREID] = strings.ToUpper(fields[FieldEREID])
	if res2, err := Compute(StageDeposit, fields); err != nil || res2.LeafHash != res.LeafHash {
		t.Fatalf("uppercase ere_id: err=%v leaf=%v", err, res2)
	}
	fields[FieldEREID] = "493da761-3473-4572-a9f4-d8f39d92b45"
	if _, err := Compute(StageDeposit, fields); err == nil || !strings.Contains(err.Error(), "36 characters") || !strings.Contains(err.Error(), "got 35") {
		t.Fatalf("truncated ere_id must be refused with its length, got err=%v", err)
	}
	fields[FieldEREID] = "493da761-3473-4572-a9f4-d8f39d92b451"
	fields[FieldContentHash] = fields[FieldContentHash][:63]
	if _, err := Compute(StageDeposit, fields); err == nil || !strings.Contains(err.Error(), "64 hex") {
		t.Fatalf("truncated content_hash must be refused, got err=%v", err)
	}
}
