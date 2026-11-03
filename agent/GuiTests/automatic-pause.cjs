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
    const errors = [], sent = [];
    const open = async () => {
        const page = await context.newPage();
        page.on('pageerror', e => errors.push(e.message));
        page.on('websocket', ws => ws.on('framesent', frame => { try { sent.push(JSON.parse(String(frame.payload))); } catch {} }));
        await page.goto(base);
        await page.waitForFunction(() => !selectionPending && acknowledgedState !== null);

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
        await state(second, 'idle');
        await first.evaluate(() => updatePauseCause());
        await state(second, 'idle');
        await first.locator('#prompt').fill('');
        console.log('PASS Stop speech cancels queue, resumes and preserves draft without reassertion');
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
        await first.locator('#prompt').fill('one'); await state(second, 'paused');
        await second.locator('#stop-speech').click(); await state(first, 'idle');
        await first.locator('#prompt').fill('one still');
        await first.evaluate(() => updatePauseCause());
        await state(first, 'idle');
        await first.locator('#prompt').fill('');
        await speak(first); await state(second, 'paused');
        await second.keyboard.press('Escape'); await state(first, 'idle');
        assert(await first.evaluate(() => activeSpeech !== null));
        await first.evaluate(() => updatePauseCause()); await state(first, 'idle');
        await first.locator('#stop-speech').click();
        console.log('PASS other device Stop/Escape overrides typing and speech without stale reassertion');
        await first.locator('#prompt').fill('disconnected draft'); await state(second, 'paused');
        await first.close(); await state(second, 'paused');
        await second.locator('#stop-speech').click(); await state(second, 'idle');
        console.log('PASS disconnect does not invent a resume; another device can resume');
        const third = await open();
        await third.evaluate(() => selectAgent('captain/worker'));
        await third.waitForFunction(() => agentID === 'captain/worker' && !selectionPending);
        await third.locator('#stop-speech').click(); await state(third, 'idle');
        await second.locator('#prompt').fill('root only'); await state(second, 'paused');
        await state(third, 'idle');
        await third.locator('#stop-speech').click(); await state(second, 'paused');
        await third.evaluate(() => {
            window.refusedTarget = false;
            const original = handle;
            handle = frame => { if (frame.type === 'error') window.refusedTarget = true; original(frame); };
            socket.send(JSON.stringify({ type: 'unpause', agent_id: 'captain' }));
        });
        await third.waitForFunction(() => window.refusedTarget);
        await state(second, 'paused');
        await second.locator('#stop-speech').click(); await state(second, 'idle');
        await second.locator('#prompt').fill('');
        await third.evaluate(() => selectAgent('captain'));
        await third.waitForFunction(() => agentID === 'captain' && !selectionPending);
        console.log('PASS commands affect only the selected Agent; a forged other-Agent target is refused');
        assert.equal(await third.locator('#narrate').isChecked(), false);
        await third.locator('#narrate').check();
        await second.waitForFunction(() => browserSpeech.value.enabled);
        await third.evaluate(() => browserSpeech.select({ speed: 3, pause: false, voice: 'saved-voice' }));
        await second.waitForFunction(() => browserSpeech.value.speed === 3 && !browserSpeech.value.pause);
        await third.reload(); await third.waitForFunction(() => !selectionPending && acknowledgedState !== null);
        assert.deepEqual(await third.evaluate(() => browserSpeech.value),
            { enabled: true, speed: 3, pause: false, voice: 'saved-voice' });
        const other = await browser.newContext();
        await other.addInitScript(nativeSpeech);
        const phone = await other.newPage(); await phone.goto(base);
        await phone.waitForFunction(() => !selectionPending && acknowledgedState !== null);
        assert.equal(await phone.locator('#narrate').isChecked(), false);
        await phone.evaluate(() => { settings = { ...settings, tts_enabled: true, tts_speed: 4 }; applySettings(); });
        assert.equal(await phone.locator('#narrate').isChecked(), false);
        assert.equal(await phone.evaluate(() => browserSpeech.value.speed), 1);
        await other.close();
        await speak(third);
        await third.locator('#narrate').uncheck();
        assert.deepEqual(await third.evaluate(() => [activeSpeech, speechQueue.length]), [null, 0]);
        await second.waitForFunction(() => !browserSpeech.value.enabled);
        console.log('PASS speech preferences survive reload, share same-profile tabs, ignore server values and isolate profiles');
        await context.setOffline(true);
        await third.evaluate(() => socket.close());
        await third.locator('#narrate').check();
        assert.equal(await third.evaluate(() => browserSpeech.value.enabled), true);
        await context.setOffline(false);
        await third.waitForFunction(() => !selectionPending && socket.readyState === WebSocket.OPEN);
        assert.equal(await third.locator('#narrate').isChecked(), true);
        await third.evaluate(() => localStorage.removeItem('ensemble.speech.v1'));
        await second.waitForFunction(() => !browserSpeech.value.enabled);
        const blocked = await browser.newContext();
        await blocked.addInitScript(() => Object.defineProperty(window, 'localStorage', { get() { throw new Error('blocked'); } }));
        const blockedPage = await blocked.newPage(); await blockedPage.goto(base);
        await blockedPage.waitForFunction(() => !selectionPending);
        await blockedPage.locator('#narrate').check();
        assert.equal(await blockedPage.evaluate(() => browserSpeech.value.enabled), true);
        await blockedPage.reload(); await blockedPage.waitForFunction(() => !selectionPending);
        assert.equal(await blockedPage.locator('#narrate').isChecked(), false);
        await blocked.close();
        assert.equal(sent.filter(frame => frame.type === 'update_settings'
            && /tts_enabled|tts_speed|pause_while_speaking/.test(JSON.stringify(frame))).length, 0);
        console.log('PASS offline preference, storage removal, blocked-storage fallback and zero server speech writes');
        assert.deepEqual(errors, []);
    } finally { await context.close(); await browser.close(); }
})().catch(e => { console.error(e); process.exitCode = 1; });
