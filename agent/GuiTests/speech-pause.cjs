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
(async () => {
    const browser = await chromium.launch({
        executablePath: process.env.CHROME_BIN || '/Applications/Google Chrome.app/Contents/MacOS/Google Chrome',
        headless: true
    });
    const errors = [];
    const context = await browser.newContext();
    context.setDefaultTimeout(8000);
    if (process.env.GUI_BASELINE_APP) {
        const original = require('node:fs').readFileSync(process.env.GUI_BASELINE_APP, 'utf8');
        await context.route('**/app.js', route => route.fulfill({ contentType: 'text/javascript', body: original }));
    }
    await context.addInitScript(nativeSpeech);
    const open = async () => {
        const page = await context.newPage();
        page.on('pageerror', error => errors.push(error.message));
        page.on('websocket', ws => ws.on('framereceived', frame => {
            try {
                const value = JSON.parse(String(frame.payload));
                if (value.type === 'error') errors.push('Server error: ' + JSON.stringify(value));
            } catch { /* Non-JSON frames are outside this assertion. */ }
        }));
        await page.goto(base);
        await page.waitForFunction(() => !selectionPending && agentID === 'captain' && acknowledgedState !== null);
        // Reset only this disposable fixture, including a prior negative-control failure.
        await page.evaluate(() => send('unpause'));
        await page.waitForFunction(() => acknowledgedState === 'idle');
        await page.evaluate(() => { settings = { ...settings, tts_enabled: true, pause_while_speaking: true }; });
        return page;
    };
    const state = (page, wanted) => page.waitForFunction(wanted => document.querySelector('#state').textContent === wanted, wanted);
    let primary, observer;
    try {
        primary = await open(); observer = await open();
        // Attach before this test's controls; the socket is already connected.
        await primary.evaluate(() => {
            window.testSent = [];
            const original = socket.send.bind(socket);
            socket.send = value => { window.testSent.push(JSON.parse(value)); original(value); };
        });
        const speak = async () => {
            // Independent scenarios start only after both connections have
            // observed the preceding release; their delivery order can differ.
            await state(primary, 'idle');
            await state(observer, 'idle');
            await primary.evaluate(() => speakText('pause-test', 'Speech used for the pause regression.'));
            await state(observer, 'paused');
        };
        const recovered = async () => {
            try { await primary.waitForFunction(() => activeSpeech === null && speechQueue.length === 0); }
            catch (error) {
                const actual = await primary.evaluate(() => ({ active: activeSpeech?.id,
                    speaking: testSpeech.speaking, pending: testSpeech.pending,
                    state: document.querySelector('#state').textContent }));
                throw new Error('Speech did not settle: ' + JSON.stringify(actual), { cause: error });
            }
        };
        await speak();
        await primary.evaluate(() => testSpeech.finish());
        await state(observer, 'idle');
        check(true, 'normal native completion releases the actual Agent pause');

        // Render and narrate actual provider-supplied thinking without an expander.
        for (const kind of ['reasoning_summary', 'thinking']) {
            const visible = await primary.evaluate(kind => {
                const card = { id: 'thinking-test', agent: 'captain', role: 'assistant', kind,
                    text: 'I will check the actual inputs.', final: true, node: document.createElement('article') };
                window.testThinkingCard = card;
                render(card);
                narrate(card);
                const firstSpeech = testSpeech.current;
                narrate(card); // Re-rendering must not repeat previously narrated text.
                return { heading: card.node.querySelector('strong').textContent,
                    body: card.node.querySelector('.message')?.textContent,
                    details: card.node.querySelectorAll('details').length,
                    spoken: firstSpeech.text, queued: speechQueue.length };
            }, kind);
            check(visible.heading === 'Thinking' && visible.body.includes('actual inputs') && !visible.details,
                kind + ': Thinking prose is immediately visible without an expander');
            check(visible.spoken === 'Thinking: I will check the actual inputs.' && visible.queued === 0,
                kind + ': automatic read-aloud occurs once without a click');
            await state(observer, 'paused');
            await primary.evaluate(() => testSpeech.finish());
            await state(observer, 'idle');
            const complete = await primary.evaluate(() => {
                testThinkingCard.text = 'Full thinking text. '.repeat(140) + 'VISIBLE_END';
                render(testThinkingCard);
                const body = testThinkingCard.node.querySelector('.message').textContent;
                testThinkingCard.node.remove();
                return body.endsWith('VISIBLE_END') && body.length > 2000;
            });
            check(complete, kind + ': complete text is rendered rather than a bounded preview');
        }
        const quietThinking = await primary.evaluate(() => {
            const card = () => ({ id: 'quiet-thinking', role: 'assistant', kind: 'reasoning_summary',
                text: 'This must remain silent.', final: true });
            settings.tts_enabled = false;
            narrate(card());
            settings.tts_enabled = true;
            replaying = true;
            narrate(card());
            replaying = false;
            return activeSpeech === null && speechQueue.length === 0;
        });
        check(quietThinking, 'thinking respects speech opt-out and silent history replay');


        await speak();
        await primary.evaluate(() => {
            // Open the actual Create Agent dialog without submitting it.
            document.querySelector('#peer-open').click();
            testSpeech.stopWithoutCallback();
        });
        await recovered(); await state(observer, 'idle');
        await primary.evaluate(() => document.querySelector('dialog[open]')?.close());
        check(true, 'lost native callback with Create Agent dialog open releases pause');

        await speak();
        await observer.bringToFront();
        await primary.evaluate(() => testSpeech.stopWithoutCallback());
        await primary.bringToFront();
        await recovered(); await state(observer, 'idle');
        check(true, 'lost callback around browser-tab switching releases pause');

        await speak();
        await primary.waitForTimeout(2700);
        check(await observer.locator('#state').textContent() === 'paused', 'still-speaking native engine is not timed out');
        await primary.evaluate(() => testSpeech.finish()); await state(observer, 'idle');

        await speak();
        await primary.evaluate(() => { testSpeech.speaking = false; testSpeech.pending = true; });
        await primary.waitForTimeout(2700);
        check(await primary.evaluate(() => activeSpeech !== null), 'native queued speech is not mistaken for completion');
        await primary.evaluate(() => testSpeech.finish()); await state(observer, 'idle');

        await speak();
        await primary.evaluate(() => { testSpeech.speaking = false; });
        await primary.waitForTimeout(700);
        check(await primary.evaluate(() => activeSpeech !== null), 'startup grace tolerates temporarily unpublished native state');
        await primary.evaluate(() => { testSpeech.speaking = true; testSpeech.finish(); });
        await state(observer, 'idle');

        await speak();
        await primary.evaluate(() => {
            window.testOldUtterance = testSpeech.current;
            enqueueSpeech('next-item', 'Next queued item.');
            testSpeech.stopWithoutCallback();
        });
        await primary.waitForFunction(() => activeSpeech?.id === 'next-item');
        await primary.evaluate(() => testOldUtterance.onend());
        check(await primary.evaluate(() => activeSpeech?.id === 'next-item'), 'old callback cannot settle the next utterance');
        check(await observer.locator('#state').textContent() === 'paused', 'queue advancement retains the speech pause');
        await primary.evaluate(() => testSpeech.finish()); await state(observer, 'idle');

        await speak();
        await observer.locator('#resume').click(); await state(primary, 'idle');
        await primary.waitForTimeout(2700);
        check(await observer.locator('#state').textContent() === 'idle', 'external Resume is not vetoed by continuing speech');
        await primary.evaluate(() => testSpeech.finish()); await recovered();

        await speak();
        await primary.locator('#prompt').fill('An unfinished draft');
        await primary.evaluate(() => testSpeech.stopWithoutCallback());
        await recovered();
        check(await observer.locator('#state').textContent() === 'paused', 'speech recovery preserves the typing pause');
        await primary.locator('#prompt').fill(''); await state(observer, 'idle');

        await speak(); await primary.locator('#pause').click();
        await primary.evaluate(() => testSpeech.stopWithoutCallback()); await recovered();
        check(await observer.locator('#state').textContent() === 'paused', 'manual Pause pins an automatic pause through completion');
        await primary.locator('#resume').click(); await state(observer, 'idle');

        await observer.locator('#pause').click(); await state(primary, 'paused');
        await primary.evaluate(() => { speakText('foreign', 'Keep the other tab pause.'); testSpeech.stopWithoutCallback(); });
        await recovered();
        check(await observer.locator('#state').textContent() === 'paused', 'speech never releases a pause owned by another tab');
        await observer.locator('#resume').click(); await state(primary, 'idle');

        await observer.evaluate(() => selectAgent('captain/worker'));
        await observer.waitForFunction(() => agentID === 'captain/worker' && !selectionPending);
        // A negative control may have stopped while this disposable worker was paused.
        await observer.evaluate(() => send('unpause'));
        await state(observer, 'idle');
        await observer.locator('#pause').click(); await state(observer, 'paused');
        await primary.evaluate(() => speakText('switch-test', 'Outgoing agent speech.'));
        await state(primary, 'paused');
        await primary.evaluate(() => {
            window.testOldUtterance = testSpeech.current;
            window.testSent.length = 0;
            selectAgent('captain/worker');
            const late = { id: 'late', text: 'Late outgoing response.', final: true, kind: 'message', narrated: 0 };
            narrate(late, true);
            window.testSuppressed = late.speechSuppressed;
        });
        await primary.waitForFunction(() => agentID === 'captain/worker' && !selectionPending);
        await state(primary, 'paused');
        const controls = await primary.evaluate(() => testSent);
        assert.deepEqual(controls.slice(0, 2).map(x => [x.type, x.agent_id]),
            [['unpause', 'captain'], ['select_agent', 'captain/worker']]);
        check(await primary.evaluate(() => testSuppressed && activeSpeech === null),
            'release precedes selection and outgoing late narration stays silent');
        await primary.evaluate(() => testOldUtterance.onend());
        check(await observer.locator('#state').textContent() === 'paused', 'obsolete callback cannot resume the incoming agent');
        await observer.locator('#resume').click(); await state(primary, 'idle');
        await primary.evaluate(() => selectAgent('captain'));
        await primary.waitForFunction(() => agentID === 'captain' && !selectionPending);
        await state(primary, 'idle');
        check(true, 'outgoing Agent was actually resumed, not merely forgotten by the UI');
    } finally {
        await context.close(); await browser.close();
    }
    check(errors.length === 0, 'no browser exceptions or server error frames: ' + errors.join('; '));
})().catch(error => { console.error(error); process.exitCode = 1; });
