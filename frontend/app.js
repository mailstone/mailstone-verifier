// Tab switching
document.querySelectorAll('.tab-button').forEach(button => {
    button.addEventListener('click', () => {
        const tabName = button.dataset.tab;

        // Remove active class from all tabs and buttons
        document.querySelectorAll('.tab-button').forEach(btn => btn.classList.remove('active'));
        document.querySelectorAll('.tab-content').forEach(content => content.classList.remove('active'));

        // Add active class to clicked button and corresponding content
        button.classList.add('active');
        document.getElementById(`${tabName}-tab`).classList.add('active');
    });
});

// Info Modal
const infoBtn = document.getElementById('info-btn');
const infoModal = document.getElementById('info-modal');
const modalClose = document.getElementById('modal-close');
const modalOverlay = document.getElementById('modal-overlay');

// Open modal
infoBtn.addEventListener('click', () => {
    infoModal.style.display = 'block';
});

// Close modal on close button click
modalClose.addEventListener('click', () => {
    infoModal.style.display = 'none';
});

// Close modal on overlay click
modalOverlay.addEventListener('click', () => {
    infoModal.style.display = 'none';
});

// Close modal on ESC key
document.addEventListener('keydown', (e) => {
    if (e.key === 'Escape' && infoModal.style.display === 'block') {
        infoModal.style.display = 'none';
    }
});

// ========================================
// TAB 1: Hash Calculation
// ========================================

const hashDropZone = document.getElementById('hash-drop-zone');
const hashFileInfo = document.getElementById('hash-file-info');
const hashFilename = document.getElementById('hash-filename');
const hashCalculateBtn = document.getElementById('hash-calculate-btn');
const hashResult = document.getElementById('hash-result');
const hashValue = document.getElementById('hash-value');
const hashCopyBtn = document.getElementById('hash-copy-btn');
const hashError = document.getElementById('hash-error');

let selectedFilePath = null;

// Click to browse - use Go backend to open file dialog
hashDropZone.addEventListener('click', async () => {
    try {
        // Call Go method to open file dialog
        const filePath = await window.go.main.App.SelectFile();

        if (filePath) {
            selectedFilePath = filePath;
            // Extract filename from path
            const filename = filePath.split(/[/\\]/).pop();
            displayHashFile(filename);
        }
    } catch (error) {
        hashError.textContent = `Error selecting file: ${error.message || error}`;
        hashError.style.display = 'block';
    }
});

// Drag and drop styling
hashDropZone.addEventListener('dragover', (e) => {
    e.preventDefault();
    hashDropZone.style.borderColor = '#667eea';
    hashDropZone.style.background = '#edf2f7';
});

hashDropZone.addEventListener('dragleave', () => {
    hashDropZone.style.borderColor = '#cbd5e0';
    hashDropZone.style.background = '#f7fafc';
});

hashDropZone.addEventListener('drop', (e) => {
    e.preventDefault();
    hashDropZone.style.borderColor = '#cbd5e0';
    hashDropZone.style.background = '#f7fafc';

    // Show error message for drag & drop
    hashError.textContent = 'Drag & drop is not supported. Please click to browse files.';
    hashError.style.display = 'block';
    setTimeout(() => {
        hashError.style.display = 'none';
    }, 3000);
});

function displayHashFile(filename) {
    hashFilename.textContent = filename;
    hashFileInfo.style.display = 'block';
    hashResult.style.display = 'none';
    hashError.style.display = 'none';
}

// Calculate hash
hashCalculateBtn.addEventListener('click', async () => {
    if (!selectedFilePath) return;

    hashCalculateBtn.disabled = true;
    hashCalculateBtn.textContent = 'Calculating...';
    hashResult.style.display = 'none';
    hashError.style.display = 'none';

    try {
        const hash = await window.go.main.App.CalculateHash(selectedFilePath);

        hashValue.textContent = hash;
        hashResult.style.display = 'block';

    } catch (error) {
        hashError.textContent = `Error: ${error.message || error}`;
        hashError.style.display = 'block';
    } finally {
        hashCalculateBtn.disabled = false;
        hashCalculateBtn.textContent = 'Generate SHA-256 Hash';
    }
});

// Copy hash to clipboard
hashCopyBtn.addEventListener('click', async () => {
    const hash = hashValue.textContent;
    try {
        await navigator.clipboard.writeText(hash);
        const successMsg = hashResult.querySelector('.success-message');
        successMsg.style.display = 'block';
        setTimeout(() => {
            successMsg.style.display = 'none';
        }, 2000);
    } catch (error) {
        alert('Failed to copy to clipboard');
    }
});

// ========================================
// TAB 2: TimeStamp Decoder
// ========================================

const timestampInput = document.getElementById('timestamp-input');
const timestampDecodeBtn = document.getElementById('timestamp-decode-btn');
const timestampResult = document.getElementById('timestamp-result');
const timestampError = document.getElementById('timestamp-error');

timestampDecodeBtn.addEventListener('click', async () => {
    const base64Token = timestampInput.value.trim();

    if (!base64Token) {
        timestampError.textContent = 'Please paste a Base64 timestamp token';
        timestampError.style.display = 'block';
        return;
    }

    timestampDecodeBtn.disabled = true;
    timestampDecodeBtn.textContent = 'Decoding...';
    timestampResult.style.display = 'none';
    timestampError.style.display = 'none';

    try {
        const expectedHash = document.getElementById('timestamp-expected-hash').value.trim();
        const result = await window.go.main.App.DecodeTimestamp(base64Token, expectedHash);

        if (result.error) {
            throw new Error(result.error);
        }

        document.getElementById('ts-provider').textContent = result.provider;
        document.getElementById('ts-datetime').textContent = result.dateTime;
        document.getElementById('ts-serial').textContent = result.serialNumber;
        document.getElementById('ts-hashalgo').textContent = result.hashAlgo;
        document.getElementById('ts-hash').textContent = result.timestampedHash;
        document.getElementById('ts-status').textContent = result.status;

        const sigCell = document.getElementById('ts-signature');
        sigCell.textContent = result.signatureVerified ? '✓ Verified against the embedded certificate' : '✗ Not verified';
        sigCell.className = result.signatureVerified ? 'verdict-ok' : 'verdict-ko';
        document.getElementById('ts-signer').textContent = result.signerSubject || '—';
        document.getElementById('ts-issuer').textContent = result.signerIssuer || '—';
        document.getElementById('ts-validity').textContent = result.signerValidFrom ? `${result.signerValidFrom} → ${result.signerValidTo}` : '—';
        const coversRow = document.getElementById('ts-covers-row');
        const coversCell = document.getElementById('ts-covers');
        if (result.coversHash) {
            coversRow.style.display = '';
            coversCell.textContent = result.coversHash === 'yes' ? '✓ Yes — this timestamp dates that hash' : '✗ No — the token covers a different hash';
            coversCell.className = result.coversHash === 'yes' ? 'verdict-ok' : 'verdict-ko';
        } else {
            coversRow.style.display = 'none';
        }
        document.getElementById('ts-chain-note').textContent = result.chainNote || '';

        timestampResult.style.display = 'block';

    } catch (error) {
        timestampError.textContent = `Error: ${error.message || error}`;
        timestampError.style.display = 'block';
    } finally {
        timestampDecodeBtn.disabled = false;
        timestampDecodeBtn.textContent = 'Decode TimeStamp';
    }
});

// ========================================
// TAB 3: Merkle Verification
// ========================================

const merkleJsonInput = document.getElementById('merkle-json-input');
const merkleHashInput = document.getElementById('merkle-hash-input');
const merkleVerifyBtn = document.getElementById('merkle-verify-btn');
const merkleResult = document.getElementById('merkle-result');
const merkleSuccess = document.getElementById('merkle-success');
const merkleFailure = document.getElementById('merkle-failure');
const merkleError = document.getElementById('merkle-error');

merkleVerifyBtn.addEventListener('click', async () => {
    const merkleJson = merkleJsonInput.value.trim();
    const hash = merkleHashInput.value.trim();

    if (!merkleJson) {
        merkleError.textContent = 'Please paste the Merkle JSON';
        merkleError.style.display = 'block';
        return;
    }

    merkleVerifyBtn.disabled = true;
    merkleVerifyBtn.textContent = 'Verifying...';
    merkleResult.style.display = 'none';
    merkleError.style.display = 'none';

    try {
        const result = await window.go.main.App.VerifyMerkle({
            merkleJson: merkleJson,
            hash: hash
        });

        if (result.error) {
            throw new Error(result.error);
        }

        if (result.success) {
            // Success case
            document.getElementById('merkle-calc-root').textContent = result.calculatedRoot;
            document.getElementById('merkle-exp-root').textContent = result.expectedRoot;
            document.getElementById('merkle-position').textContent = `Leaf #${result.leafIndex} / ${result.totalLeaves} total leaves`;
            document.getElementById('merkle-type').textContent = result.leafType.toUpperCase();

            // Display merkle path steps
            const stepsDiv = document.getElementById('merkle-steps');
            stepsDiv.textContent = result.merklePathSteps.join('\n');

            merkleSuccess.style.display = 'block';
            merkleFailure.style.display = 'none';
        } else {
            // Failure case
            document.getElementById('merkle-calc-root-fail').textContent = result.calculatedRoot;
            document.getElementById('merkle-exp-root-fail').textContent = result.expectedRoot;

            merkleSuccess.style.display = 'none';
            merkleFailure.style.display = 'block';
        }

        merkleResult.style.display = 'block';

    } catch (error) {
        merkleError.textContent = `Error: ${error.message || error}`;
        merkleError.style.display = 'block';
    } finally {
        merkleVerifyBtn.disabled = false;
        merkleVerifyBtn.textContent = 'Verify Merkle Proof';
    }
});

// ========================================
// TAB 4: ERE Verification
// ========================================

// --- Recipient decision signature -------------------------------------
const ereVerifyBtn = document.getElementById('ere-verify-btn');
const ereSigResult = document.getElementById('ere-sig-result');
const ereSigError = document.getElementById('ere-sig-error');

ereVerifyBtn.addEventListener('click', async () => {
    ereSigResult.style.display = 'none';
    ereSigError.style.display = 'none';
    ereVerifyBtn.disabled = true;
    ereVerifyBtn.textContent = 'Verifying...';
    try {
        const result = await window.go.main.App.VerifyEreDecision({
            publicKey: document.getElementById('ere-pk-input').value.trim(),
            signature: document.getElementById('ere-sig-input').value.trim(),
            message: document.getElementById('ere-msg-input').value,
            ereId: document.getElementById('ere-id-input').value.trim(),
            decision: document.getElementById('ere-decision-input').value,
            decidedAt: document.getElementById('ere-decided-input').value.trim(),
        });
        if (result.error) {
            throw new Error(result.error);
        }
        const title = document.getElementById('ere-sig-title');
        const text = document.getElementById('ere-sig-text');
        if (result.valid) {
            title.textContent = '✓ Signature valid';
            title.className = 'verdict-ok';
            text.textContent = 'The holder of this public key signed exactly this decision, for exactly this delivery, at exactly this time. The decision cannot be repudiated nor replayed on another delivery.';
        } else {
            title.textContent = '✗ Signature does not match';
            title.className = 'verdict-ko';
            text.textContent = 'Either the message differs from what was signed (one character is enough — check the ere_id, the decision, the RFC 3339 date and that every line ends with a line feed), or the key or signature is not the one printed in the proof.';
        }
        document.getElementById('ere-sig-bytes').textContent = result.message.replace(/\n/g, '\\n\n');
        ereSigResult.style.display = 'block';
    } catch (error) {
        ereSigError.textContent = `Error: ${error.message || error}`;
        ereSigError.style.display = 'block';
    } finally {
        ereVerifyBtn.disabled = false;
        ereVerifyBtn.textContent = 'Verify Decision Signature';
    }
});

// --- Event evidence hash ------------------------------------------------
// The fields each event commits to, in the order the proof prints them.
const EVIDENCE_FIELDS = {
    deposit: [
        ['ere_id', 'ere_id (UUID)'],
        ['sender_email', 'Sender email'],
        ['recipient_email', 'Recipient email'],
        ['subject', 'Subject (exactly as printed, may be empty)'],
        ['content_hash', 'Content hash (SHA-256 of the Email PDF, hex)'],
    ],
    content: [
        ['content_hash', 'Content hash (SHA-256 of the Email PDF, hex) — or hash the file below'],
    ],
    emission: [
        ['ere_id', 'ere_id (UUID)'],
        ['provider_message_id', 'Provider message ID'],
        ['submitted_at', 'Hand-over time (RFC 3339, with fractional seconds)'],
    ],
    delivery: [
        ['ere_id', 'ere_id (UUID)'],
        ['delivered_at', 'Delivery time confirmed by the provider (RFC 3339)'],
        ['provider_message_id', 'Provider message ID'],
    ],
    presentation: [
        ['ere_id', 'ere_id (UUID)'],
        ['ordinal', 'Presentation number (2, 3, …)'],
        ['presented_at', 'Delivery time confirmed by the provider (RFC 3339)'],
        ['provider_message_id', 'Provider message ID'],
    ],
    decision: [
        ['signature', 'Recipient signature (Ed25519, base64)'],
        ['received_at', 'Time the platform received the decision (RFC 3339)'],
    ],
    abort: [
        ['ere_id', 'ere_id (UUID)'],
        ['sender_user_id', 'Sender user ID (UUID)'],
        ['aborted_at', 'Cancellation time (RFC 3339)'],
    ],
    expiry: [
        ['ere_id', 'ere_id (UUID)'],
        ['expires_at', 'Expiry time (RFC 3339)'],
    ],
};
const EVIDENCE_NOTES = {
    content: 'The content leaf is the SHA-256 of the Email PDF itself (the "Preuve de contenu" you downloaded): there is no canonical text, the file digest is what was anchored.',
    emission: 'The hand-over instant is hashed with its fractional seconds: use the value from the proof\'s technical annex, not the rounded time of the lifecycle table.',
};

const ereStageSelect = document.getElementById('ere-stage-select');
const ereEvidenceFields = document.getElementById('ere-evidence-fields');
const ereEvidenceBtn = document.getElementById('ere-evidence-btn');
const ereEvidenceResult = document.getElementById('ere-evidence-result');
const ereEvidenceError = document.getElementById('ere-evidence-error');
const ereEvidenceJson = document.getElementById('ere-evidence-json');

// Values typed so far, by field name. ere_id, provider_message_id, the
// content hash… are shared between events: switching the event keeps them.
const evidenceValues = {};

function renderEvidenceFields() {
    const stage = ereStageSelect.value;
    ereEvidenceFields.innerHTML = '';
    ereEvidenceResult.style.display = 'none';
    ereEvidenceError.style.display = 'none';
    (EVIDENCE_FIELDS[stage] || []).forEach(([name, label]) => {
        const group = document.createElement('div');
        group.className = ['content_hash', 'signature', 'ere_id', 'sender_user_id', 'provider_message_id'].includes(name)
            ? 'input-group wide' : 'input-group';
        const lab = document.createElement('label');
        lab.setAttribute('for', `ere-field-${name}`);
        lab.textContent = label;
        const input = document.createElement('input');
        input.type = 'text';
        input.id = `ere-field-${name}`;
        input.className = 'input-text';
        input.dataset.field = name;
        input.value = evidenceValues[name] || '';
        input.addEventListener('input', () => { evidenceValues[name] = input.value; });
        group.appendChild(lab);
        group.appendChild(input);
        ereEvidenceFields.appendChild(group);
    });
    if (stage === 'content') {
        // The leaf is the file's digest: offer to hash the Email PDF here.
        const group = document.createElement('div');
        group.className = 'input-group';
        const btn = document.createElement('button');
        btn.className = 'btn btn-primary';
        btn.textContent = 'Hash the Email PDF file…';
        btn.addEventListener('click', async () => {
            try {
                const filePath = await window.go.main.App.SelectFile();
                if (!filePath) return;
                const hash = await window.go.main.App.CalculateHash(filePath);
                evidenceValues.content_hash = hash;
                const field = document.getElementById('ere-field-content_hash');
                if (field) field.value = hash;
            } catch (error) {
                ereEvidenceError.textContent = `Error: ${error.message || error}`;
                ereEvidenceError.style.display = 'block';
            }
        });
        group.appendChild(btn);
        ereEvidenceFields.appendChild(group);
    }
    if (EVIDENCE_NOTES[stage]) {
        const note = document.createElement('p');
        note.className = 'description';
        note.textContent = EVIDENCE_NOTES[stage];
        ereEvidenceFields.appendChild(note);
    }
}
ereStageSelect.addEventListener('change', renderEvidenceFields);
renderEvidenceFields();

// checkAgainstBlock: with the event's proof block pasted, confirm the
// recomputed leaf is the block's leaf and that the root reconstructs.
async function checkAgainstBlock(leafHash) {
    const box = document.getElementById('ere-evidence-block');
    const title = document.getElementById('ere-evidence-block-title');
    const text = document.getElementById('ere-evidence-block-text');
    const json = ereEvidenceJson.value.trim();
    if (!json) {
        box.style.display = 'none';
        return;
    }
    let blockLeaf = '';
    try {
        const parsed = JSON.parse(json);
        if (parsed.leaves && parsed.leaves.length === 1) blockLeaf = (parsed.leaves[0].leaf_hash || '').toLowerCase();
    } catch (e) {
        title.textContent = '✗ Proof block is not valid JSON';
        title.className = 'verdict-ko';
        text.textContent = e.message || String(e);
        box.style.display = 'block';
        return;
    }
    const result = await window.go.main.App.VerifyMerkle({ merkleJson: json, hash: leafHash });
    if (result.success) {
        title.textContent = '✓ Facts → leaf → root: all three match';
        title.className = 'verdict-ok';
        text.textContent = `The leaf rebuilt from the facts is exactly the leaf anchored in this block (leaf #${result.leafIndex} of ${result.totalLeaves}), and the block's sibling hashes lead from it to the Racine Merkle ${result.calculatedRoot} — the value printed on the card. These facts were therefore part of the anchored batch. Last link: in the TimeStamp Decoder, paste the card's token with this root as the hash to cover; the token's date is the date of this event.`;
    } else {
        title.textContent = '✗ The leaf rebuilt from these facts is not the one anchored in this block';
        title.className = 'verdict-ko';
        text.textContent = blockLeaf
            ? `Anchored leaf (block's leaf_hash): ${blockLeaf}. Rebuilt leaf: ${leafHash}. At least one fact differs from what the platform hashed — compare each value with the proof character by character (a trailing space in the subject, a wrong date or a date without its fractional seconds, another message id).`
            : (result.error || 'The leaf is not in this block.');
    }
    box.style.display = 'block';
}

ereEvidenceBtn.addEventListener('click', async () => {
    ereEvidenceResult.style.display = 'none';
    ereEvidenceError.style.display = 'none';
    const fields = {};
    ereEvidenceFields.querySelectorAll('input[data-field]').forEach((input) => {
        fields[input.dataset.field] = input.value;
    });
    ereEvidenceBtn.disabled = true;
    try {
        const result = await window.go.main.App.ComputeEreEvidenceHash({ stage: ereStageSelect.value, fields });
        if (result.error) {
            throw new Error(result.error);
        }
        document.getElementById('ere-evidence-hash').textContent = result.leafHash;
        document.getElementById('ere-evidence-canonical').textContent = result.canonical.replace(/\n/g, '\\n\n');
        ereEvidenceResult.style.display = 'block';
        await checkAgainstBlock(result.leafHash);
    } catch (error) {
        ereEvidenceError.textContent = `Error: ${error.message || error}`;
        ereEvidenceError.style.display = 'block';
    } finally {
        ereEvidenceBtn.disabled = false;
    }
});

document.getElementById('ere-evidence-copy-btn').addEventListener('click', async () => {
    try {
        await navigator.clipboard.writeText(document.getElementById('ere-evidence-hash').textContent);
        const msg = ereEvidenceResult.querySelector('.success-message');
        msg.style.display = 'block';
        setTimeout(() => { msg.style.display = 'none'; }, 2000);
    } catch (error) {
        alert('Failed to copy to clipboard');
    }
});

// Hand the leaf to the Merkle tab: the reader pastes that event's proof
// block there and verifies.
document.getElementById('ere-evidence-to-merkle-btn').addEventListener('click', () => {
    merkleHashInput.value = document.getElementById('ere-evidence-hash').textContent;
    if (ereEvidenceJson.value.trim()) merkleJsonInput.value = ereEvidenceJson.value.trim();
    document.querySelector('.tab-button[data-tab="merkle"]').click();
    merkleJsonInput.focus();
});
