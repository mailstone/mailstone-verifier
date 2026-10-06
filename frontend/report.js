// Session journal and report export.
//
// Every verification the reader runs is appended to an in-memory journal;
// "Export" in the title bar writes the whole journal as a text report (in
// the interface language) or as JSON (stable English keys, exact values).
// The report documents what was checked, when, with which values — it is
// produced by the reader's own copy of the tool, not issued by MailStone.

const journal = [];

function journalAdd(entry) {
    entry.seq = journal.length + 1;
    entry.at = new Date().toISOString();
    journal.push(entry);
    refreshJournalUI();
    return entry.seq;
}

function journalClear() {
    journal.length = 0;
    document.querySelectorAll('.journal-note').forEach((n) => n.remove());
    refreshJournalUI();
}

// markAdded: a discreet line in a result box, so the reader sees that the
// verifications accumulate without any action on their part.
function markAdded(box, seq) {
    let note = box.querySelector('.journal-note');
    if (!note) {
        note = document.createElement('p');
        note.className = 'journal-note';
        box.appendChild(note);
    }
    note.dataset.seq = String(seq);
    note.textContent = t('report.added', { n: seq });
}

function refreshJournalUI() {
    const count = document.getElementById('export-count');
    const exportBtn = document.getElementById('export-btn');
    const clearBtn = document.getElementById('clear-btn');
    if (!count) return;
    count.textContent = String(journal.length);
    exportBtn.disabled = journal.length === 0;
    clearBtn.disabled = journal.length === 0;
    document.querySelectorAll('.journal-note').forEach((n) => { n.textContent = t('report.added', { n: n.dataset.seq }); });
}

// ---- rendering ----------------------------------------------------------

function pad(n) { return String(n).padStart(2, '0'); }
function fmtUTC(d) { return `${d.getUTCFullYear()}-${pad(d.getUTCMonth() + 1)}-${pad(d.getUTCDate())} ${pad(d.getUTCHours())}:${pad(d.getUTCMinutes())}:${pad(d.getUTCSeconds())}`; }
function fmtLocal(d) { return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`; }
function yesNo(b) { return b ? t('report.yes') : t('report.no'); }

function reportJSON(info, now) {
    return {
        tool: info.name,
        version: info.version,
        platform: `${info.os}/${info.arch}`,
        exported_at: now.toISOString(),
        language: currentLang,
        note: t('report.note'),
        entries: journal,
    };
}

function reportText(info, now) {
    const L = [];
    L.push(`${info.name} ${info.version} — ${t('report.title')}`);
    L.push(t('report.exportedAt', { utc: fmtUTC(now), local: fmtLocal(now), os: `${info.os}/${info.arch}` }));
    L.push(t('report.count', { n: journal.length }));
    L.push(t('report.note'));
    L.push('');
    journal.forEach((e) => {
        const at = new Date(e.at);
        L.push(`${e.seq}. ${t('report.op.' + e.operation)} — ${fmtLocal(at)} (${fmtUTC(at)} UTC)`);
        switch (e.operation) {
            case 'file_hash':
                L.push(`   ${t('report.file')} : ${e.file.name}${e.file.size ? ` (${e.file.size} ${t('report.bytes')})` : ''}`);
                L.push(`   SHA-256 : ${e.sha256}`);
                break;
            case 'timestamp': {
                const r = e.result;
                L.push(`   ${t('ts.provider')} ${r.provider}`);
                L.push(`   ${t('ts.datetime')} ${r.date_time}`);
                L.push(`   ${t('ts.serial')} ${r.serial_number}`);
                L.push(`   ${t('ts.algo')} ${r.hash_algo}`);
                L.push(`   ${t('ts.hash')} ${r.timestamped_hash}`);
                L.push(`   ${t('ts.status')} ${r.status}`);
                L.push(`   ${t('ts.signature')} ${r.signature_verified ? t('ts.sigOk') : t('ts.sigKo')}`);
                if (r.signer_subject) L.push(`   ${t('ts.signer')} ${r.signer_subject}`);
                if (r.signer_issuer) L.push(`   ${t('ts.issuer')} ${r.signer_issuer}`);
                if (r.signer_valid_from) L.push(`   ${t('ts.validity')} ${r.signer_valid_from} → ${r.signer_valid_to}`);
                if (e.expected_hash) {
                    L.push(`   ${t('report.expectedHash')} : ${e.expected_hash}`);
                    L.push(`   ${t('ts.covers')} ${r.covers_hash === 'yes' ? t('ts.coversYes') : t('ts.coversNo')}`);
                }
                L.push(`   ${t('report.token')} : ${e.token.slice(0, 32)}… (${t('report.tokenFull', { n: e.token.length })})`);
                break;
            }
            case 'merkle': {
                const r = e.result;
                L.push(`   ${t('report.blockLeaf')} : ${e.hash || t('report.blockOwnLeaf')}`);
                L.push(`   ${t('merkle.calc')} ${r.calculated_root}`);
                L.push(`   ${t('merkle.expected')} ${r.expected_root}`);
                if (r.success) {
                    L.push(`   ${t('merkle.position')} ${t('merkle.leafOf', { index: r.leaf_index, total: r.total_leaves })} · ${t('merkle.type')} ${(r.leaf_type || '').toUpperCase()}`);
                }
                L.push(`   ${r.success ? '✓ ' + t('merkle.ok') : '✗ ' + t('merkle.ko')}`);
                break;
            }
            case 'ere_decision':
                L.push(`   ${t('report.publicKey')} : ${e.public_key}`);
                L.push(`   ${t('report.signature')} : ${e.signature}`);
                L.push(`   ${t('report.message')} :`);
                e.message.split('\n').forEach((line) => { if (line !== '') L.push(`      ${line}`); });
                L.push(`   ${e.result.valid ? t('ere.sig.validTitle') : t('ere.sig.invalidTitle')}`);
                break;
            case 'ere_evidence': {
                L.push(`   ${t('report.event')} : ${t('ere.ev.stage.' + e.event)}`);
                Object.keys(e.inputs).forEach((k, i) => { L.push(`   (${i + 1}) ${k} = ${e.inputs[k]}`); });
                if (e.file) L.push(`   ${t('report.file')} : ${e.file.name} — SHA-256 ${e.file.hash}`);
                if (e.canonical) L.push(`   ${t('report.canonical')} : ${e.canonical.replace(/\n/g, '\\n')}`);
                L.push(`   ${t('report.leaf')} : ${e.leaf_hash}`);
                if (e.printed_hash) {
                    L.push(`   ${t('report.printedHash')} : ${e.printed_hash}`);
                    if (e.printed_matches !== null) L.push(`   ${e.printed_matches ? t('ere.ev.printedOkTitle') : t('ere.ev.printedKoTitle')}`);
                }
                if (e.proof_block) {
                    const b = e.proof_block;
                    if (b.error) {
                        L.push(`   ✗ ${b.error}`);
                    } else if (b.verdict) {
                        L.push(`   ${t('ere.ev.okTitle')} — ${t('merkle.leafOf', { index: b.leaf_index, total: b.total_leaves })}, ${t('merkle.calc')} ${b.root_hash}`);
                    } else {
                        L.push(`   ${t('ere.ev.koTitle')}${b.block_leaf ? ` — leaf_hash ${b.block_leaf}` : ''}`);
                    }
                }
                break;
            }
            default:
                L.push(`   ${JSON.stringify(e)}`);
        }
        L.push('');
    });
    return L.join('\n') + '\n';
}

// ---- export UI ----------------------------------------------------------

function showToast(text) {
    const el = document.getElementById('toast');
    el.textContent = text;
    el.style.display = 'block';
    clearTimeout(showToast.timer);
    showToast.timer = setTimeout(() => { el.style.display = 'none'; }, 5000);
}

async function exportReport(format) {
    document.getElementById('export-menu').style.display = 'none';
    if (journal.length === 0) return;
    try {
        const info = await window.go.main.App.GetAppInfo();
        const now = new Date();
        const content = format === 'json' ? JSON.stringify(reportJSON(info, now), null, 2) : reportText(info, now);
        const stamp = `${now.getFullYear()}${pad(now.getMonth() + 1)}${pad(now.getDate())}-${pad(now.getHours())}${pad(now.getMinutes())}`;
        const path = await window.go.main.App.ExportReport({ format, content, defaultName: `mailstone-verifier-${t('report.fileStem')}-${stamp}.${format}` });
        if (path) showToast(t('report.exported', { path }));
    } catch (error) {
        showToast(t('error.prefix') + (error.message || error));
    }
}

document.getElementById('export-btn').addEventListener('click', (ev) => {
    ev.stopPropagation();
    const menu = document.getElementById('export-menu');
    menu.style.display = menu.style.display === 'block' ? 'none' : 'block';
});
document.addEventListener('click', () => { document.getElementById('export-menu').style.display = 'none'; });
document.getElementById('export-txt').addEventListener('click', (ev) => { ev.stopPropagation(); exportReport('txt'); });
document.getElementById('export-json').addEventListener('click', (ev) => { ev.stopPropagation(); exportReport('json'); });
document.getElementById('clear-btn').addEventListener('click', async () => {
    if (journal.length === 0) return;
    const ok = await window.go.main.App.ConfirmDialog(t('report.clearConfirmTitle'), t('report.clearConfirm', { n: journal.length }));
    if (ok) journalClear();
});
document.addEventListener('languagechange', refreshJournalUI);
refreshJournalUI();
