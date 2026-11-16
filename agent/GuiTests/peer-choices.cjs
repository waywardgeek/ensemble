// DOM/HTTP-boundary test of the shipped creation dialog. Core route behavior is
// separately exercised by PeerControlsTests; this fixture performs no inference.
const { chromium } = require(process.env.PLAYWRIGHT_MODULE || 'playwright');
const fs = require('node:fs');
const path = require('node:path');
const http = require('node:http');
const assert = require('node:assert/strict');
const root = path.resolve(__dirname, '../gui/wwwroot');
const offering = { vendor: 'ollama', endpoint: 'et', model: 'mimo' };
const preference = { vendor: 'ollama', model: 'fixture-mimo' };
const flint = endpoint => ({ vendor: 'openrouter', endpoint, model: 'flint-v1_0-free' });
let flintChoices = false;
let failChoices = false;
const creations = [];

(async () => {
    const server = http.createServer(async (request, response) => {
        const url = new URL(request.url, 'http://localhost');
        if (url.pathname.startsWith('/peers/')) {
            if (request.headers['x-ensemble-peers'] !== '1') { response.writeHead(403).end(); return; }
            response.setHeader('Content-Type', 'application/json');
            if (url.pathname === '/peers/saved') { response.end('[]'); return; }
            if (url.pathname === '/peers/choices') {
                if (failChoices) { response.writeHead(503).end('{}'); return; }
                if (flintChoices) {
                    response.end(JSON.stringify({ configured_models: [
                        { vendor: 'OpenRouter', label: 'Flint legacy Messages', model: 'recursive/flint-v1.0:free', offering: flint('api'), enabled: true },
                        { vendor: 'OpenRouter', label: 'Flint Chat', model: 'recursive/flint-v1.0:free', offering: flint('chat'), enabled: true }
                    ], preferred_models: [
                        { vendor: 'OpenRouter', label: 'Flint default unavailable', preference: { vendor: 'openrouter', model: 'recursive/flint-v1.0:free' }, enabled: false }
                    ], primary_skills: [] })); return;
                }
                response.end(JSON.stringify({ model_choices: ['astra'], configured_models: [
                    { vendor: 'Ollama', label: 'MiMo', model: 'fixture-mimo', offering, funding: 'No authentication', enabled: true },
                    { vendor: 'Ollama', label: 'Unavailable', model: 'gone', offering: { ...offering, model: 'gone' }, funding: 'Unavailable', enabled: false }
                ], preferred_models: [
                    { vendor: 'Ollama', label: 'MiMo', preference, funding: 'No authentication', enabled: true },
                    { vendor: 'Ollama', label: 'Missing default', preference: { ...preference, model: 'gone' }, funding: 'unavailable; choose explicit funding', enabled: false }
                ], primary_skills: [{ name: 'small-coding', description: 'Six tools' }, { name: 'ensemble', description: 'Full skill' }] }));
                return;
            }
            if (url.pathname === '/peers/create') {
                let body = '';
                for await (const chunk of request) body += chunk;
                const value = JSON.parse(body); creations.push(value);
                response.end(JSON.stringify({ name: value.name, created: true })); return;
            }
            response.writeHead(404).end(); return;
        }
        const file = path.resolve(root, '.' + (url.pathname === '/' ? '/index.html' : url.pathname));
        if (!file.startsWith(root + path.sep)) { response.writeHead(403).end(); return; }
        try {
            const data = fs.readFileSync(file);
            response.setHeader('Content-Type', file.endsWith('.js') ? 'text/javascript' :
                file.endsWith('.css') ? 'text/css' : file.endsWith('.html') ? 'text/html' : 'application/octet-stream');
            response.end(data);
        } catch { response.writeHead(404).end(); }
    });
    let browser;
    try {
        await new Promise(resolve => server.listen(0, '127.0.0.1', resolve));
        browser = await chromium.launch({ headless: true,
            executablePath: process.env.CHROME_PATH || '/Applications/Google Chrome.app/Contents/MacOS/Google Chrome' });
        const page = await browser.newPage();
        await page.route('**/*.js', route => ['/peers.js', '/layout.js'].some(file => route.request().url().endsWith(file)) ? route.continue() : route.abort());
        await page.goto(`http://127.0.0.1:${server.address().port}/`);
        for (const width of [1200, 390]) {
            await page.setViewportSize({ width, height: 850 });
            await page.waitForFunction(() => document.body.classList.contains('compact') === (innerWidth < 62 * parseFloat(getComputedStyle(document.documentElement).fontSize)));
            if (await page.locator('body').evaluate(el => el.classList.contains('compact')))
                await page.locator('button[data-pane=agents]').click();
            await page.locator('#peer-open').click();
            await page.waitForFunction(() => !document.querySelector('#peer-create').disabled);
            assert.equal(await page.locator('#peer-model').evaluate(e => e.tagName), 'SELECT');
            assert.equal(await page.locator('#peer-skill').evaluate(e => e.tagName), 'SELECT');
            assert.equal(await page.locator('#peer-model optgroup').getAttribute('label'), 'Ollama');
            assert.equal(await page.locator('#peer-model option[data-offering-key]:enabled').count(), 0);
            assert.equal(await page.locator('#peer-model option').filter({ hasText: /^MiMo$/ }).count(), 1);
            await page.locator('#peer-name').fill(`small-${width}`);
            await page.locator('#peer-model').selectOption({ label: 'MiMo' });
            await page.locator('#peer-skill').selectOption('small-coding');
            await page.locator('#peer-create').click();
            await page.locator('#peer-notice').filter({ hasText: `Created Agent small-${width}` }).waitFor();
            assert.deepEqual(creations.at(-1), { name: `small-${width}`, skill: 'small-coding', preference });
            await page.locator('#peer-open').click();
            await page.waitForFunction(() => !document.querySelector('#peer-create').disabled);
            assert.equal(await page.locator('#peer-model option').filter({ hasText: 'Missing default' }).isDisabled(), true);
            await page.locator('#peer-name').fill(`preferred-${width}`);
            await page.locator('#peer-model').selectOption({ label: 'MiMo' });
            await page.locator('#peer-skill').selectOption('small-coding');
            await page.locator('#peer-create').click();
            await page.locator('#peer-notice').filter({ hasText: `Created Agent preferred-${width}` }).waitFor();
            assert.match(await page.locator('#peer-notice').textContent(), new RegExp(`Created Agent preferred-${width}`));
            assert.deepEqual(creations.at(-1), { name: `preferred-${width}`, skill: 'small-coding', preference });
        }
        await page.locator('#peer-open').click();
        await page.waitForFunction(() => !document.querySelector('#peer-create').disabled);
        await page.locator('#peer-name').fill('inherited');
        await page.locator('#peer-create').click();
        await page.locator('#peer-notice').filter({ hasText: 'Created Agent inherited' }).waitFor();
        assert.deepEqual(creations.at(-1), { name: 'inherited', skill: '', model: '' });
        failChoices = true;
        await page.locator('#peer-open').click();
        await page.locator('#peer-choice-status').filter({ hasText: 'choices unavailable' }).waitFor();
        assert.equal(await page.locator('#peer-create').isDisabled(), true);
        assert.equal(creations.length, 5);
        await page.locator('#peer-close').click();
        failChoices = false;
        await page.locator('#peer-open').click();
        await page.waitForFunction(() => !document.querySelector('#peer-create').disabled);
        await page.locator('#peer-close').click();
        flintChoices = true;
        for (const width of [1200, 390]) {
            await page.setViewportSize({ width, height: 850 });
            await page.waitForFunction(() => document.body.classList.contains('compact') === (innerWidth < 62 * parseFloat(getComputedStyle(document.documentElement).fontSize)));
            if (await page.locator('body').evaluate(el => el.classList.contains('compact')))
                await page.locator('button[data-pane=agents]').click();
            await page.locator('#peer-open').click();
            await page.waitForFunction(() => !document.querySelector('#peer-create').disabled);
            assert.equal(await page.locator('#peer-model option[data-offering-key]:enabled').count(), 2,
                'Ambiguous grouped funding must not hide explicit protocol choices');
            await page.locator('#peer-name').fill(`flint-${width}`);
            await page.locator('#peer-model').selectOption({ label: 'Flint Chat' });
            await page.locator('#peer-create').click();
            await page.locator('#peer-notice').filter({ hasText: `Created Agent flint-${width}` }).waitFor();
            assert.deepEqual(creations.at(-1), { name: `flint-${width}`, skill: '', offering: flint('chat') });
        }
        // Exercise the shipped picker projection independently of usage counters
        // and the conversation socket; creation above uses the full peers script.
        const appSource = fs.readFileSync(path.join(root, 'app.js'), 'utf8');
        const start = appSource.indexOf('function showUsage(frame) {');
        const end = appSource.indexOf('    renderControls();', start);
        assert(start >= 0 && end > start);
        const pickerBody = appSource.slice(start + 'function showUsage(frame) {'.length, end);
        for (const endpoint of ['api', 'chat']) {
            const result = await page.evaluate(({ body, endpoint }) => {
                const offering = endpoint => ({ vendor: 'openrouter', endpoint, model: 'flint-v1_0-free' });
                const frame = {
                    selected_model: 'recursive/flint-v1.0:free', selected_offering: offering(endpoint), selection_available: true,
                    preferred_models: [{ vendor: 'OpenRouter', label: 'Unresolved Flint', model: 'recursive/flint-v1.0:free',
                        preference: { vendor: 'openrouter', model: 'recursive/flint-v1.0:free' }, selected: true, enabled: false }],
                    configured_models: ['api', 'chat'].map(endpoint => ({ vendor: 'OpenRouter', label: endpoint,
                        model: 'recursive/flint-v1.0:free', offering: offering(endpoint), enabled: true }))
                };
                new Function('frame', body)(frame);
                const picker = document.querySelector('#model-choice');
                return { selected: JSON.parse(picker.value), enabled: [...picker.options].filter(option => !option.disabled).length };
            }, { body: pickerBody, endpoint });
            assert.deepEqual(result, { selected: flint(endpoint), enabled: 2 });
        }
        console.log('PASS peer dropdowns: desktop/mobile, grouped preferences, explicit protocol recovery, exact selected endpoint, primary skill, inheritance, refusal/retry');
    } finally {
        try { if (browser) await browser.close(); }
        finally { await new Promise(resolve => server.close(resolve)); }
    }
})().catch(error => { console.error(error); process.exitCode = 1; });
