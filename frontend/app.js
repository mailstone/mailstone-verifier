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
        const result = await window.go.main.App.DecodeTimestamp(base64Token);

        if (result.error) {
            throw new Error(result.error);
        }

        document.getElementById('ts-provider').textContent = result.provider;
        document.getElementById('ts-datetime').textContent = result.dateTime;
        document.getElementById('ts-serial').textContent = result.serialNumber;
        document.getElementById('ts-hashalgo').textContent = result.hashAlgo;
        document.getElementById('ts-hash').textContent = result.timestampedHash;
        document.getElementById('ts-status').textContent = result.status;

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

    if (!hash) {
        merkleError.textContent = 'Please enter the hash to verify';
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
