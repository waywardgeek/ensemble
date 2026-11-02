// Execute the production control renderer without a socket or model request.
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const vm = require('node:vm');
const source = fs.readFileSync(path.join(__dirname, '../gui/wwwroot/app.js'), 'utf8');
const start = source.indexOf('function renderControls()');
const end = source.indexOf('// Format acknowledged Engine values', start);
assert(start >= 0 && end > start);
const cases = [
    ['idle', 'inputpending', true, false, false, false],
    ['paused', 'paused', true, false, false, false],
    ['request in flight', 'inflight', true, false, false, true],
    ['tools pending', 'toolspending', true, false, false, true],
    ['disconnected', 'inputpending', false, false, false, true],
    ['selection pending', 'inputpending', true, true, false, true],
    ['closed', 'inputpending', true, false, true, true],
    ['unacknowledged', null, true, false, false, true],
];
for (const [name, state, connected, selecting, closed, disabled] of cases) {
    const elements = new Map();
    const context = {
        document: {querySelector(selector) {
            if (!elements.has(selector)) elements.set(selector, {value: '', setAttribute() {}});
            return elements.get(selector);
        }},
        socket: {readyState: connected ? 1 : 3}, WebSocket: {OPEN: 1},
        selectionPending: selecting, acknowledgedState: state, agentClosed: closed,
        voiceControl: null, imageComposer: {}, pageCapture: null, captureDraft: null,
        captureConsent: null, pauseIntent: null,
    };
    vm.runInNewContext(source.slice(start, end) + '\nrenderControls();', context);
    assert.equal(elements.get('#model-choice').disabled, disabled, name);
}
console.log('PASS model picker controls: eight acknowledged-state cases; no model requests.');
