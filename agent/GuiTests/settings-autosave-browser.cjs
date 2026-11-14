// Run: node agent/GuiTests/settings-autosave-browser.cjs
// Shipped markup and settings handlers, synthetic acknowledgments, no provider.
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const { chromium } = require('../../edition-2/agent/gui/node_modules/playwright');
const root = path.join(__dirname, '../gui/wwwroot');
const source = fs.readFileSync(path.join(root, 'app.js'), 'utf8');
const start = source.indexOf('function sameWorkingContextModel(');
const end = source.indexOf('document.querySelector("#font-size").onchange', start);
assert(start >= 0 && end > start);
(async () => {
    const browser = await chromium.launch({channel: 'chrome', headless: true});
    try {
        const page = await browser.newPage();
        const errors = []; page.on('pageerror', error => errors.push(error.message));
        await page.setContent(fs.readFileSync(path.join(root, 'index.html'), 'utf8')
            .replace(/<script\b[^>]*>[\s\S]*?<\/script>/gi, ''));
        await page.addScriptTag({content: `
            var sent = [], notices = [], agentID = 'test', workingContext = null, workingContextForm = null;
            var socket = {readyState: WebSocket.OPEN, send: text => sent.push(JSON.parse(text))};
            var browserTheme = {seed() {}}, browserDevice = {seed() {}};
            function renderControls() {} function resizeComposer() {} function updatePauseCause() {}
            function notice(text) { notices.push(text); } function send(type) { sent.push({type}); }
            var settings = {max_tool_rounds: 5, context_target_bytes: 1000,
                working_context: {vendor: 'test', model: 'one', working_tokens: 100000, capacity_tokens: 200000},
                memory: Object.fromEntries(['conversation','soul','memory','archive64','archive8','session']
                    .map(band => [band, {budget: 6000, disabled: false}])),
                recall: {disabled: false, disable_judge: false, judge_model: '', min_score: 0,
                    max_snippets: 3, min_query_length: 5, candidates: 10, context_messages: 2, max_recall_bytes: 6000}};
        ` + source.slice(start, end) + '\napplySettings();'});
        const edit = (id, value, commit = true) => page.evaluate(({id, value, commit}) => {
            const control = document.getElementById(id);
            if (control.type === 'checkbox') control.checked = value; else control.value = String(value);
            control.dispatchEvent(new Event('input', {bubbles: true}));
            if (commit) control.dispatchEvent(new Event('change', {bubbles: true}));
        }, {id, value, commit});
        await edit('memory-session-budget', 7000);
        assert.equal(await page.evaluate(() => sent.at(-1).settings.memory.session.budget), 7000);
        await edit('memory-session-budget', 8000, false);
        await page.evaluate(() => { settings.memory.session.budget = 7000; applySettings(); });
        assert.equal(await page.locator('#memory-session-budget').inputValue(), '8000');
        assert.equal(await page.evaluate(() => settings.memory.session.budget), 7000, 'Draft is not policy');
        await edit('memory-session-budget', 8000);
        await page.evaluate(() => { settings.memory.session.budget = 8000; applySettings(); });
        assert.equal(await page.evaluate(() => settingsDrafts.has('memory-session-budget')), false);
        const beforeInvalid = await page.evaluate(() => sent.length);
        await edit('memory-session-budget', '');
        await edit('recall-max_snippets', 2147483648);
        assert.equal(await page.evaluate(() => sent.length), beforeInvalid, 'HTML required/max checks survive autosave');
        await edit('memory-session-budget', 8000);
        await edit('recall-max_snippets', 3);
        await edit('recall-enabled', false);
        assert.equal(await page.evaluate(() => sent.at(-1).settings.recall.disabled), true);
        await edit('working-context-tokens', 120000);
        assert.equal(await page.evaluate(() => sent.at(-1).settings.working_context.working_tokens), 120000);
        const count = await page.evaluate(() => sent.length);
        await edit('context-capacity-tokens', 100000);
        assert.equal(await page.evaluate(() => sent.length), count, 'Invalid pair is not sent');
        await page.evaluate(() => applySettings()); // Includes the unrelated-error refresh path.
        assert.equal(await page.locator('#context-capacity-tokens').inputValue(), '100000');
        await page.evaluate(() => showWorkingContext({vendor: 'test', model: 'two', working_tokens: 50000, capacity_tokens: 90000}));
        assert.equal(await page.locator('#working-context-tokens').inputValue(), '50000');
        assert.equal(await page.evaluate(() => contextSettingIDs.some(id => settingsDrafts.has(id))), false);
        await edit('memory-session-budget', 9000, false);
        await page.evaluate(() => { document.querySelector('#memory-settings-reload').click(); applySettings(); });
        assert.equal(await page.locator('#memory-session-budget').inputValue(), '8000');
        await page.evaluate(() => socket.readyState = WebSocket.CLOSED);
        await edit('memory-session-budget', 9500);
        assert.equal(await page.locator('#memory-session-budget').inputValue(), '9500');
        assert.equal(await page.evaluate(() => notices.at(-1)), 'Not connected; settings were not saved.');
        await page.evaluate(() => { settingsDrafts.clear(); settings.memory.session.budget = 4000; applySettings(); });
        assert.equal(await page.locator('#memory-session-budget').inputValue(), '4000');
        assert.deepEqual(errors, []);
        console.log('PASS autosave: all groups, delayed acknowledgment, draft/policy separation, invalid pair, model change, reload, disconnect');
    } finally { await browser.close(); }
})().catch(error => { console.error(error); process.exitCode = 1; });
