// Disposable CrossoverFixture only: this test edits its saved GUI/Agent settings.
// Run GuiTests --crossover-host separately; never point this at a live user host.
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const { chromium } = require(process.env.PLAYWRIGHT_MODULE || 'playwright');
const base = process.env.GUI_FIXTURE_URL || 'http://127.0.0.1:18202/assistant/';
const address = new URL(base);
assert.equal(address.hostname, '127.0.0.1');
assert(!['8088', '8089'].includes(address.port));
const output = process.env.GUI_ARTIFACT_DIR || '/private/tmp/cadence-settings-dialog/browser';
fs.mkdirSync(output, { recursive: true });

(async () => {
    const browser = await chromium.launch({
        executablePath: process.env.CHROME_BIN || '/Applications/Google Chrome.app/Contents/MacOS/Google Chrome',
        headless: true
    });
    const context = await browser.newContext({ viewport: { width: 1600, height: 1000 } });
    context.setDefaultTimeout(10000);
    await context.addInitScript(() => {
        Object.defineProperty(window, 'speechSynthesis', { value: {
            speaking: false, pending: false, getVoices: () => [], addEventListener() {}, cancel() {},
            speak(utterance) { queueMicrotask(() => utterance.onend?.()); }
        } });
        window.SpeechSynthesisUtterance = class { constructor(text) { this.text = text; } };
    });
    const page = await context.newPage();
    const errors = [], replies = [], sent = [], checks = [];
    page.on('pageerror', error => errors.push(error.message));
    page.on('websocket', socket => {
        socket.on('framesent', frame => {
            try { sent.push(JSON.parse(String(frame.payload))); } catch {}
        });
        socket.on('framereceived', frame => {
            try {
                const value = JSON.parse(String(frame.payload));
                replies.push(value);
                if (value.type === 'error') errors.push('Server refusal: ' + JSON.stringify(value));
            } catch {}
        });
    });
    function check(name) { checks.push(name); console.log('PASS ' + name); }
    async function ready() {
        await page.waitForFunction(() => !selectionPending && agentID === 'captain' && settings !== null);
        await page.waitForFunction(() => document.querySelector('#send').disabled === false);
    }
    async function open() {
        await page.locator('#settings-open').click();
        await page.waitForFunction(() => document.querySelector('#settings-dialog').open);
    }
    async function tab(name) {
        await page.locator(`#settings-${name}-tab`).click();
        await page.waitForFunction(name => document.querySelector(`#settings-${name}-tab`).getAttribute('aria-selected') === 'true', name);
    }
    async function save(action) {
        const start = replies.length;
        await action();
        await page.waitForFunction(() => document.querySelector('#notice').textContent === '');
        const deadline = Date.now() + 10000;
        while (!replies.slice(start).some(frame => frame.type === 'settings_updated')) {
            if (Date.now() > deadline) throw new Error('No server settings acknowledgment');
            await new Promise(resolve => setTimeout(resolve, 20));
        }
    }
    try {
        await page.goto(base);
        await ready();
        const original = await page.evaluate(() => ({
            human_name: settings.human_name,
            max_tool_rounds: settings.max_tool_rounds,
            soul: settings.memory.soul.disabled,
            recall: settings.recall.context_messages,
            font_size: settings.font_size
        }));
        assert.equal(await page.locator('#settings-tab').count(), 0);
        assert.equal(await page.locator('#settings-dialog').isVisible(), false);
        assert.equal(await page.locator('#agents-panel').isVisible(), true);
        await open();
        assert.equal(await page.evaluate(() => document.activeElement.id), 'settings-ai-tab');
        await page.keyboard.press('ArrowRight');
        await page.waitForFunction(() => document.querySelector('#settings-accounts-tab').getAttribute('aria-selected') === 'true');
        await page.keyboard.press('End');
        assert.equal(await page.evaluate(() => document.activeElement.id), 'settings-recall-tab');
        await page.keyboard.press('Home');
        assert.equal(await page.evaluate(() => document.activeElement.id), 'settings-ai-tab');
        check('gear opens an accessible dialog; Bootstrap arrow/Home/End tab navigation works');

        await tab('goals');
        const goalsEnabled = await page.locator('#goal-file').isEnabled();
        if (goalsEnabled) await page.locator('#goal-file').fill('unsaved-fixture-goal.md');
        else assert.equal(await page.locator('#goal-file').isDisabled(), true);
        const beforeDraft = sent.length;
        await tab('display');
        await page.locator('#human-name').fill('Unsaved profile');
        await tab('memory');
        await page.keyboard.press('Escape');
        assert.equal(await page.evaluate(() => document.activeElement.id), 'settings-open');
        await open();
        assert.equal(await page.locator('#settings-memory-tab').getAttribute('aria-selected'), 'true');
        await tab('display');
        assert.equal(await page.locator('#human-name').inputValue(), 'Unsaved profile');
        await tab('goals');
        if (goalsEnabled) assert.equal(await page.locator('#goal-file').inputValue(), 'unsaved-fixture-goal.md');
        assert(!sent.slice(beforeDraft).some(frame => frame.type === 'update_settings' || frame.type === 'configure_goal'));
        check('Escape restores gear focus; drafts and selected category survive close/reopen without saving');

        await tab('display');
        await page.locator('#human-name').fill('Fixture settings tester');
        await save(() => page.locator('#user-settings button[type="submit"]').click());
        const font = 22;
        await save(async () => {
            await page.locator('#font-size').fill(String(font));
            await page.locator('#font-size').dispatchEvent('change');
        });
        await tab('ai');
        await save(() => page.locator('#max-tool-rounds').fill(String(original.max_tool_rounds + 1)).then(() => page.locator('#max-tool-rounds').blur()));
        await tab('memory');
        await page.locator('#memory-soul-enabled').setChecked(original.soul);
        await save(() => page.locator('#memory-settings-form button[type="submit"]').click());
        await tab('recall');
        await page.locator('#recall-context_messages').fill(String(original.recall + 1));
        await save(() => page.locator('#recall-settings-form button[type="submit"]').click());
        await page.reload();
        await ready();
        const saved = await page.evaluate(() => ({
            human_name: settings.human_name, max_tool_rounds: settings.max_tool_rounds,
            soul: settings.memory.soul.disabled, recall: settings.recall.context_messages,
            font_size: settings.font_size
        }));
        assert.deepEqual(saved, {
            human_name: 'Fixture settings tester', max_tool_rounds: original.max_tool_rounds + 1,
            soul: !original.soul, recall: original.recall + 1, font_size: Number(font)
        });
        check('profile, appearance, AI, Memory, and Recall saves are acknowledged and persist after reload');

        await open();
        await tab('accounts');
        await page.locator('#account-refresh').click();
        await page.waitForFunction(() => document.querySelector('#account-status').textContent.length > 0);
        assert.equal(await page.locator('#accounts-panel summary').count(), 0);
        assert.equal(await page.locator('#goal-controls summary').count(), 0);
        check('account status remains readable without login actions or redundant expanders');

        const layouts = [];
        for (const theme of ['dark', 'light']) {
            for (const [width, height, text] of [[1600,1000,18],[1024,768,28],[390,844,18],[390,844,28],[320,568,32]]) {
                await page.setViewportSize({ width, height });
                await page.evaluate(({theme,text}) => {
                    document.documentElement.dataset.bsTheme = theme;
                    document.documentElement.style.fontSize = text + 'px';
                }, {theme,text});
                for (const category of ['ai','accounts','display','goals','memory','recall']) {
                    await tab(category);
                    const geometry = await page.evaluate(() => {
                        const dialog = document.querySelector('#settings-dialog');
                        const body = dialog.querySelector('.settings-body');
                        const tabs = dialog.querySelector('.settings-tabs');
                        const rect = dialog.getBoundingClientRect();
                        const active = dialog.querySelector('[role="tabpanel"].active');
                        return {
                            x: rect.x, y: rect.y, right: rect.right, bottom: rect.bottom,
                            width: innerWidth, height: innerHeight, bodyHeight: body.clientHeight,
                            bodyWidth: body.clientWidth, bodyScrollWidth: body.scrollWidth,
                            contentHeight: body.scrollHeight, tabHeight: tabs.clientHeight,
                            selectedTabs: dialog.querySelectorAll('[role="tab"][aria-selected="true"]').length,
                            panels: [...dialog.querySelectorAll('[role="tabpanel"]')].filter(panel => getComputedStyle(panel).display !== 'none').length,
                            active: active.id
                        };
                    });
                    assert(geometry.x >= 0 && geometry.y >= 0 && geometry.right <= width + 1 && geometry.bottom <= height + 1, JSON.stringify(geometry));
                    assert(geometry.bodyHeight >= 120, 'Insufficient scrolling content height: ' + JSON.stringify(geometry));
                    assert(geometry.bodyScrollWidth <= geometry.bodyWidth + 2, 'Horizontal content overflow: ' + JSON.stringify(geometry));
                    assert.equal(geometry.selectedTabs, 1);
                    assert.equal(geometry.panels, 1);
                    if (category === 'memory' && width === 320) {
                        await page.locator('.settings-body').evaluate(element => { element.scrollTop = element.scrollHeight; });
                        assert(await page.locator('.settings-body').evaluate(element => element.scrollTop > 0));
                        const top = await page.locator('.settings-tabs').evaluate(element => element.getBoundingClientRect().top);
                        assert(top >= 0, 'Tabs scrolled out of the dialog');
                    }
                }
                layouts.push({theme,width,height,text});
            }
        }
        check('all six tabs fit both themes, desktop/phone widths, and enlarged text; content scrolls independently');
        await page.locator('#settings-close').click();
        assert.equal(await page.evaluate(() => document.activeElement.id), 'settings-open');
        const header = await page.locator('.status-actions').evaluate(element => [...element.children].map(child => ({top:child.getBoundingClientRect().top, bottom:child.getBoundingClientRect().bottom})));
        assert(header.every(rect => Math.abs(rect.top - header[0].top) < 1), 'Gear created another controls row');
        assert.deepEqual(errors, []);
        check('close button restores focus; header controls stay on one row; no browser/server errors');
        fs.writeFileSync(path.join(output, 'result.json'), JSON.stringify({checks, layouts, errors}, null, 2));
    } finally {
        await context.close();
        await browser.close();
    }
})().catch(error => { console.error(error); process.exitCode = 1; });
