// Real GUI/Agent/WebSocket regression; only native speech is controlled.
// Start GuiTests.dll --crossover-host in an isolated working directory first.
// DISPOSABLE HOST ONLY: setup explicitly resumes its captain/worker fixtures.
// PLAYWRIGHT_MODULE may identify an existing installation; no download is done.
const { chromium } = require(process.env.PLAYWRIGHT_MODULE || 'playwright');
const assert = require('node:assert/strict');
const base = process.env.GUI_FIXTURE_URL || 'http://127.0.0.1:18202/assistant/';
const target = new URL(base);
assert.equal(target.hostname, '127.0.0.1', 'Use a disposable loopback fixture');
assert.ok(!['8088', '8089'].includes(target.port), 'Never run this test against the working hosts');
const check = (ok, label) => { assert.ok(ok, label); console.log('PASS ' + label); };
async function nativeSpeech() {
    class Utterance { constructor(text) { this.text = text; } }
    const engine = {
        speaking: false, pending: false, current: null,
        getVoices() { return []; }, addEventListener() {},
        speak(item) {
            if (!item.text) return; // Existing silent Chrome wakeup.
            this.current = item; this.speaking = true; this.pending = false;
        },
        cancel() { this.speaking = false; this.pending = false; },
        stopWithoutCallback() { this.cancel(); },
        finish() { const item = this.current; this.cancel(); item.onend?.(); }
    };
    Object.defineProperty(window, 'speechSynthesis', { value: engine });
    Object.defineProperty(window, 'SpeechSynthesisUtterance', { value: Utterance });
    window.testSpeech = engine;
}
// Real streamed turn, real actor status, controlled native speech only.
(async () => {
    const browser = await chromium.launch({
        executablePath: process.env.CHROME_BIN || '/Applications/Google Chrome.app/Contents/MacOS/Google Chrome',
        headless: true
    });
    const context = await browser.newContext();
    context.setDefaultTimeout(10000);
    await context.addInitScript(nativeSpeech);
    const mode = process.env.SPEECH_COMPLETION || 'normal';
    assert.ok(['normal', 'missing-callback', 'disconnect'].includes(mode));
    try {
        const page = await context.newPage();
        const errors = [];
        page.on('pageerror', error => { errors.push(error.message); console.error('PAGE', error.message); });
        page.on('websocket', connection => connection.on('framereceived', ({ payload }) => {
            try { const frame = JSON.parse(String(payload));
                if (frame.type === 'error' || (frame.type === 'turn_ended' && frame.error)) console.error('SERVER', JSON.stringify(frame));
            } catch { }
        }));
        await page.goto(base);
        await page.waitForFunction(() => !selectionPending && agentID === 'captain');
        await page.evaluate(() => send('unpause'));
        await page.waitForFunction(() => acknowledgedState === 'idle');
        // Seeded fixture images require an actually supported media model,
        // not the launch placeholder or the first discovery-menu entry.
        await page.evaluate(() => send('set_model', 'claude-3-haiku-20240307'));
        await page.waitForFunction(() => !document.querySelector('#model-choice').disabled
            && document.querySelector('#model-choice').value === 'claude-3-haiku-20240307');
        await page.evaluate(() => {
            settings = { ...settings, tts_enabled: true, pause_while_speaking: true };
            send('prompt', 'Stream the isolated speech completion test.');
        });
        await page.waitForFunction(() => document.querySelector('#transcript').textContent.includes('STREAM-IDENTITY first'))
            .catch(async error => { console.error('fixture setup', await page.evaluate(() => ({ state: acknowledgedState, selected: document.querySelector('#model-choice').value, automatic: pauseIntent, queue: speechQueue.length, speech: !!activeSpeech }))); throw error; });
        await page.evaluate(() => speakText('held-across-turn', 'Hold this speech across model completion.'));
        await page.waitForFunction(() => acknowledgedState === 'paused');
        await page.waitForFunction(() => [...cards.values()].some(card => card.final && card.text.includes('STREAM-IDENTITY first and final.')));
        const status = async () => {
            const response = await context.request.get(new URL('/fixture/activity', base).href);
            assert.equal(response.status(), 200);
            return response.json();
        };
        const deadline = Date.now() + 10000;
        let actual;
        do {
            actual = await status();
            if (actual.activity === 'Idle') break;
            await page.waitForTimeout(50);
        } while (Date.now() < deadline);
        assert.deepEqual(actual, { activity: 'Idle', paused: true });
        check(await page.evaluate(() => activeSpeech !== null), mode + ': real turn finished idle while speech still holds pause');
        if (mode === 'disconnect') {
            await page.evaluate(() => socket.close());
            await page.waitForFunction(() => !selectionPending && socket.readyState === WebSocket.OPEN);
        } else {
            // A final answer may queue a second utterance; settle each one,
            // never force the application counters or pause intent.
            for (let i = 0; i < 8; i++) {
                const active = await page.evaluate(() => activeSpeech !== null);
                if (!active) break;
                await page.evaluate(mode => mode === 'normal' ? testSpeech.finish() : testSpeech.stopWithoutCallback(), mode);
                await page.waitForTimeout(mode === 'normal' ? 50 : 2600);
            }
        }
        await page.waitForFunction(() => activeSpeech === null && speechQueue.length === 0);
        const releaseDeadline = Date.now() + 3500;
        do {
            actual = await status();
            if (!actual.paused) break;
            await page.waitForTimeout(50);
        } while (Date.now() < releaseDeadline);
        console.log('Final actor status: ' + JSON.stringify(actual));
        assert.deepEqual(actual, { activity: 'Idle', paused: false }, mode + ': completed speech must release its automatic pause');
        assert.deepEqual(errors, []);
        check(true, mode + ': automatic pause released after normal turn completion');
    } finally {
        await context.close();
        await browser.close();
    }
})().catch(error => { console.error(error); process.exitCode = 1; });
