// Real isolated GUI, actor and sockets; only native speech is simulated.
const { chromium } = require(process.env.PLAYWRIGHT_MODULE || 'playwright');
const assert = require('node:assert/strict');
const base = 'http://127.0.0.1:18202/assistant/';
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
    const browser = await chromium.launch({ executablePath: '/Applications/Google Chrome.app/Contents/MacOS/Google Chrome', headless: true });
    const context = await browser.newContext();
    context.setDefaultTimeout(8000);
    await context.addInitScript(nativeSpeech);
    const errors = [];
    const open = async () => {
        const page = await context.newPage();
        page.on('pageerror', e => errors.push(e.message));
        await page.goto(base);
        await page.waitForFunction(() => !selectionPending && acknowledgedState !== null);
        await page.evaluate(() => { settings = {...settings, pause_while_speaking: true}; });
        return page;
    };
    const state = (page, value) => page.waitForFunction(value => acknowledgedState === value, value);
    const speak = page => page.evaluate(() => speakText('fixture', 'First sentence. Second sentence.'));
    try {
        const first = await open(), second = await open();
        await first.evaluate(() => send('unpause')); // Only reset explicit fixture startup state.
        await state(first, 'idle');
        assert.equal(await first.locator('#pause, #resume').count(), 0);
        assert.equal(await first.locator('#speech-controls #stop-speech').count(), 1);
        await speak(first); await state(second, 'paused');
        await first.locator('#prompt').fill('draft');
        await first.locator('#stop-speech').click();
        assert.equal(await first.locator('#prompt').inputValue(), 'draft');
        assert.deepEqual(await first.evaluate(() => [activeSpeech, speechQueue.length]), [null, 0]);
        // Settings acknowledgement is ordered after preceding controls.
        await first.evaluate(() => { window.barrier = false; const original = handle; handle = frame => { original(frame); if (frame.type === 'current_settings') window.barrier = true; }; send('get_settings'); });
        await first.waitForFunction(() => window.barrier);
        await state(second, 'paused');
        await first.locator('#prompt').fill(''); await state(second, 'idle');
        console.log('PASS Stop speech cancels queue but preserves typed draft hold');
        await speak(first); await first.locator('#prompt').fill('still typing');
        await first.keyboard.press('Escape');
        assert.equal(await first.locator('#prompt').inputValue(), 'still typing');
        assert.deepEqual(await first.evaluate(() => [activeSpeech, speechQueue.length]), [null, 0]);
        await first.locator('#prompt').fill(''); await state(second, 'idle');
        console.log('PASS Escape matches Stop speech');
        await speak(first); await state(second, 'paused');
        await first.locator('#prompt').fill('draft'); await first.locator('#prompt').fill('');
        assert(await first.evaluate(() => activeSpeech !== null));
        await state(second, 'paused');
        await first.locator('#stop-speech').click(); await state(second, 'idle');
        console.log('PASS clearing typing does not release speech hold');
        await first.locator('#prompt').fill('one'); await second.locator('#prompt').fill('two');
        await state(second, 'paused');
        await first.close();
        await second.locator('#prompt').fill('two still');
        await state(second, 'paused');
        await second.locator('#prompt').fill(''); await state(second, 'idle');
        console.log('PASS disconnect releases only departing connection hold');
        const third = await open(); await speak(third); await state(second, 'paused');
        await third.close(); await state(second, 'idle');
        console.log('PASS last speech owner disconnect resumes actual Agent');
        assert.deepEqual(errors, []);
    } finally { await context.close(); await browser.close(); }
})().catch(e => { console.error(e); process.exitCode = 1; });
