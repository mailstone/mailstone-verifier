package signature

import (
	"testing"
	"time"
)

// A real decision from the MailStone staging platform (2026-10-05): the
// recipient accepted delivery 20f8521f…; key, signature and message are the
// ones printed in the proof document.
const (
	stagingPK  = "ChW40iw5g4UvRWYDzLzxJyQSEgGm1+rFKoBAr2/kql0="
	stagingSig = "hfO08m4p2Oz3ishxJv1E9G7vio0PXkSNnEUdTzDIg6oAUj5nXN/2AUvQu4jTBChtLT45paoX0xiEO95Ie5mxCg=="
	stagingMsg = "ere/decision/v1\nere_id=20f8521f-177f-468e-bd04-341f7caf0583\ndecision=accepted\ndecided_at=2026-10-05T09:36:36Z\n"
)

func TestVerifyRealStagingDecision(t *testing.T) {
	r, err := Verify(stagingPK, stagingSig, stagingMsg)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if !r.Valid {
		t.Fatal("the real staging decision must verify")
	}
}

func TestDecisionMessageMatchesTheSignedBytes(t *testing.T) {
	at := time.Date(2026, 10, 5, 9, 36, 36, 0, time.UTC)
	if got := DecisionMessage("20f8521f-177f-468e-bd04-341f7caf0583", "accepted", at); got != stagingMsg {
		t.Fatalf("DecisionMessage built\n%q\nwant\n%q", got, stagingMsg)
	}
	// The display format of the proof's timeline is not the signed format.
	paris := at.In(time.FixedZone("CEST", 2*3600))
	if DecisionMessage("20f8521f-177f-468e-bd04-341f7caf0583", "accepted", paris) != stagingMsg {
		t.Fatal("decided_at must be rendered in UTC whatever zone the caller holds")
	}
}

func TestNormalizeMessageForgivesCopyPaste(t *testing.T) {
	for _, pasted := range []string{
		stagingMsg,                     // exact
		stagingMsg[:len(stagingMsg)-1], // final LF dropped
		"  " + stagingMsg + "\n\n",     // extra blank lines
		"ere/decision/v1\r\nere_id=20f8521f-177f-468e-bd04-341f7caf0583\r\ndecision=accepted\r\ndecided_at=2026-10-05T09:36:36Z\r\n", // Windows line endings
	} {
		r, err := Verify(stagingPK, stagingSig, NormalizeMessage(pasted))
		if err != nil || !r.Valid {
			t.Errorf("pasted %q must verify after normalisation (valid=%v err=%v)", pasted, r != nil && r.Valid, err)
		}
	}
}

func TestOneByteOffIsFalseNotAnError(t *testing.T) {
	for _, msg := range []string{
		"ere/decision/v1\nere_id=20f8521f-177f-468e-bd04-341f7caf0583\ndecision=refused\ndecided_at=2026-10-05T09:36:36Z\n",
		"ere/decision/v1\nere_id=20f8521e-177f-468e-bd04-341f7caf0583\ndecision=accepted\ndecided_at=2026-10-05T09:36:36Z\n",
		"ere/decision/v1\nere_id=20f8521f-177f-468e-bd04-341f7caf0583\ndecision=accepted\ndecided_at=2026-10-05 09:36:36 UTC · 11:36:36 CEST\n",
	} {
		r, err := Verify(stagingPK, stagingSig, msg)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if r.Valid {
			t.Errorf("altered message must not verify: %q", msg)
		}
	}
}

func TestMalformedInputsAreErrors(t *testing.T) {
	if _, err := Verify("not base64!", stagingSig, stagingMsg); err == nil {
		t.Error("bad base64 key must be an error")
	}
	if _, err := Verify("AAAA", stagingSig, stagingMsg); err == nil {
		t.Error("short key must be an error")
	}
	if _, err := Verify(stagingPK, "AAAA", stagingMsg); err == nil {
		t.Error("short signature must be an error")
	}
}

// A signature copied from the PDF carries the break the page inserted where
// the 88-character string wrapped — a space or a newline in the middle.
func TestBase64CopiedFromThePageWithALineBreak(t *testing.T) {
	broken := stagingSig[:69] + " " + stagingSig[69:]
	r, err := Verify(stagingPK, broken, stagingMsg)
	if err != nil || !r.Valid {
		t.Fatalf("a space at the wrap point must be ignored (valid=%v err=%v)", r != nil && r.Valid, err)
	}
	r, err = Verify(stagingPK[:20]+"\n"+stagingPK[20:], stagingSig[:40]+"\r\n"+stagingSig[40:], stagingMsg)
	if err != nil || !r.Valid {
		t.Fatalf("line breaks inside key and signature must be ignored (valid=%v err=%v)", r != nil && r.Valid, err)
	}
}
