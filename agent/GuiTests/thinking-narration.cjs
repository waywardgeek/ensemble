// Run: node agent/GuiTests/thinking-narration.cjs
// Exercise actual browser narration functions; native audio/DOM are external stubs.
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const vm = require('node:vm');
const assets = path.join(__dirname, '../gui/wwwroot');
const app = fs.readFileSync(path.join(assets, 'app.js'), 'utf8');
function functionSource(name, next) {
    const start = app.indexOf(`function ${name}(`);
    const end = app.indexOf(`function ${next}(`, start);
    assert.ok(start >= 0 && end > start, `Locate ${name}`);
    return app.slice(start, end);
}
const context = vm.createContext({
    transcriptSpeech: true, speechQueue: [], updatePauseCause() {}, startSpeech() {},
    selectionPending: false, replaying: false, browserSpeech: { value: { enabled: true } },
    cards: new Map()
});
vm.runInContext(fs.readFileSync(path.join(assets, 'speech.js'), 'utf8') + '\n'
    + functionSource('enqueueSpeech', 'cancelSpeech') + '\n'
    + functionSource('flushNarration', 'actionButton'), context);
function drain() { return Array.from(context.speechQueue.splice(0), item => item.text); }
for (const kind of ['thinking', 'reasoning_summary']) {
    const card = { id: kind, role: 'assistant', kind, text: '', final: false };
    context.narrate(card);
    card.text = 'First thought';
    context.narrate(card);
    assert.deepEqual(drain(), [], 'Incomplete fragment stays buffered');
    card.text += '. Second thought. ';
    context.narrate(card);
    assert.deepEqual(drain(), ['Thinking: First thought.', 'Second thought.']);
    card.text += 'Third thought. Final tail';
    context.narrate(card);
    assert.deepEqual(drain(), ['Third thought.']);
    context.narrate(card, true);
    assert.deepEqual(drain(), ['Final tail']);
    card.final = true;
    context.narrate(card);
    assert.deepEqual(drain(), [], 'Terminal update cannot repeat speech');
    context.narrate({ id: kind + '-next', role: 'assistant', kind, text: 'New passage.', final: true });
    assert.deepEqual(drain(), ['Thinking: New passage.']);
    for (const gate of ['replay', 'muted']) {
        const quiet = { id: gate, role: 'assistant', kind, text: 'Old sentence. ', final: false };
        context.replaying = gate === 'replay';
        context.browserSpeech.value.enabled = gate !== 'muted';
        context.narrate(quiet);
        context.replaying = false;
        context.browserSpeech.value.enabled = true;
        quiet.text += 'Restored tail.'; quiet.final = true;
        context.narrate(quiet);
        assert.deepEqual(drain(), [], 'Replay/muted card remains suppressed');
    }
    console.log('PASS ' + kind + ': one announcement across streaming, flush, final and new passage; replay/mute silent');
}
context.narrate({ id: 'answer', role: 'assistant', kind: 'text', text: 'First answer. Second answer.', final: true });
assert.deepEqual(drain(), ['First answer.', 'Second answer.']);
context.speakText('manual', 'Manual first. Manual second.');
assert.deepEqual(drain(), ['Manual first.', 'Manual second.']);
context.speakText('error', 'First error. Second error.', 'error');
assert.deepEqual(drain(), ['Error: First error.', 'Error: Second error.']);
console.log('PASS ordinary text, explicit reading and error labels unchanged');
