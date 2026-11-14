// Run: node agent/GuiTests/error-cards.cjs
// Exercise the actual frame handler; browser rendering/speech are separate boundaries.
const assert = require('node:assert/strict');
const fs = require('node:fs');
const vm = require('node:vm');
const path = require('node:path');
const source = fs.readFileSync(path.join(__dirname, '../gui/wwwroot/app.js'), 'utf8');
const start = source.indexOf('function handle(frame) {');
// Parse one function with the JS engine, without loading page initialization.
const end = source.indexOf('\nfunction ', start + 1);
assert(start >= 0 && end > start);
const cards = new Map(), spoken = [], rendered = [];
const state = vm.createContext({ cards, agentID: 'test', selectionPending: false,
    goalSelectionPending: false, replaying: false, browserSpeech: { value: { enabled: true } },
    localErrorSequence: 0, pendingCardRenders: new Set(), jobSessions: null,
    applySettings() {}, flushNarration() {}, notice() {}, renderPairLink() {},
    speakText(id, text) { spoken.push(text); },
    cardFor(id) {
        if (!cards.has(id)) cards.set(id, { id, agent: 'test', node: { remove() {} } });
        return cards.get(id);
    },
    render(card) { rendered.push(card.text); }
});
vm.runInContext(source.slice(start, end), state);
const error = { type: 'lifecycle', kind: 'error_occurred', agent_id: 'test', seq: 7,
    part_id: 'test:e7:error', message: 'Safe failure <img src=x onerror=bad()>', recorded_at: '2026-11-14T00:00:00Z' };
state.handle(error);
assert.equal(cards.get(error.part_id)?.kind, 'error', 'Durable error needs a visible card');
assert.equal(cards.get(error.part_id)?.text, error.message);
assert.equal(cards.get(error.part_id)?.recordedAt, error.recorded_at);
state.handle(error);
assert.equal(cards.size, 1);
assert.deepEqual(spoken, [error.message], 'Repeated event must not repeat speech');
state.handle({ type: 'lifecycle', kind: 'turn_interrupted', agent_id: 'test', message: 'Interrupted' });
assert.equal(cards.size, 1, 'Nonerrors must not create error cards');
cards.clear(); state.replaying = true; state.handle(error); state.replaying = false;
assert.equal(cards.size, 1, 'Replay must restore the error');
assert.equal(spoken.length, 1, 'Replay must be silent');
state.handle({ ...error, agent_id: 'other', part_id: 'other:e7:error' });
assert.equal(cards.size, 1, 'Other Agent cannot insert errors');
state.handle({ type: 'error', message: 'Control refused' });
assert.equal(cards.size, 2, 'Control refusal must be visible');
state.handle({ type: 'display_window', agent_id: 'test', part_ids: [error.part_id] });
assert.equal(cards.size, 1, 'Authoritative refresh expires local control cards');
state.handle({ type: 'display_window', agent_id: 'test', part_ids: [] });
assert.equal(cards.size, 0, 'Error cards must obey window eviction');
console.log('PASS error cards: live display, deduplication, silent replay, Agent isolation, controls and eviction');
