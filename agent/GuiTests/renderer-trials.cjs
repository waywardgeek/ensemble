'use strict';
// Real controls and imports against an isolated, no-inference host.
const assert = require('node:assert/strict');
const fs = require('node:fs/promises');
const os = require('node:os');
const path = require('node:path');
const { spawn } = require('node:child_process');
const { chromium } = require('../../edition-2/agent/gui/node_modules/playwright');
(async () => {
    const directory = await fs.mkdtemp(path.join(os.tmpdir(), 'renderer-trial-'));
    const child = spawn(process.argv[2], [process.argv[3], '--artifact-renderer-fixture', directory]);
    let log = '';
    child.stdout.on('data', data => { log += data; });
    child.stderr.on('data', data => { log += data; });
    const exited = new Promise(resolve => child.on('exit', resolve));
    let browser;
    try {
        let base;
        for (let i = 0; i < 300; i++) {
            try { base = await fs.readFile(path.join(directory, 'url'), 'utf8'); break; } catch {}
            if (child.exitCode !== null) throw new Error(log);
            await new Promise(resolve => setTimeout(resolve, 100));
        }
        assert(base, 'fixture readiness');
        browser = await chromium.launch({ executablePath: '/Applications/Google Chrome.app/Contents/MacOS/Google Chrome', headless: true });
        for (const width of [1280, 390]) {
            const page = await browser.newPage({ viewport: { width, height: 844 } });
            const errors = [];
            page.on('pageerror', error => errors.push(error.message));
            const state = page.locator('#renderer-status');
            async function ready(text) { await page.waitForFunction(text => document.querySelector('#renderer-status').textContent.includes(text), text); }
            async function panel() {
                await page.locator('#settings-open').click();
                await page.locator('#settings-display-tab').click();
            }
            async function candidate(marker, broken = false) {
                const pkg = { entry: 'main.js', files: {
                    'main.js': broken ? "throw new Error('deliberate candidate failure');" :
                        "import { label } from './helper.js'; window.trialMarker = label; export default (document,text) => { const node=document.createElement('pre'); node.textContent=label+': '+text; return node; };",
                    'helper.js': 'export const label = ' + JSON.stringify(marker) + ';'
                } };
                await page.locator('#renderer-package').setInputFiles({ name: 'candidate.json', mimeType: 'application/json', buffer: Buffer.from(JSON.stringify(pkg)) });
                await ready('Staging has not executed');
                assert((await page.locator('#renderer-source').textContent()).includes('helper.js'));
            }
            await page.goto(base);
            await ready('Using the shipped');
            await page.waitForFunction(() => settings !== null && !selectionPending);
            await page.evaluate(() => {
                handle({ type: 'tool_dispatched', part_id: 'trial-call', name: 'list_directory', input: { path: '.' } });
                handle({ type: 'tool_finished', part_id: 'trial-result', tool_part_id: 'trial-call', result: 'recorded-file.txt' });
            });
            await panel();
            await candidate('trial-' + width);
            assert.equal(await page.evaluate(() => window.trialMarker), undefined, 'stage executed module');
            await page.locator('#renderer-try').click();
            await ready('Trial active');
            assert.equal(await page.evaluate(() => window.trialMarker), 'trial-' + width);
            await page.waitForFunction(marker => cards.get('trial-result').node.querySelector('.preview').textContent === marker + ': recorded-file.txt', 'trial-' + width);
            assert.equal(await page.evaluate(() => cards.get('trial-result').node.querySelector('details > pre').textContent), 'recorded-file.txt');
            await page.reload();
            await ready('Using the shipped');
            assert.equal(await page.evaluate(() => window.trialMarker), undefined, 'reload resumed trial');
            await panel();
            await candidate('kept-' + width);
            await page.locator('#renderer-try').click();
            await ready('Trial active');
            await page.locator('#renderer-keep').click();
            await ready('Accepted ');
            await page.reload();
            await ready('Using accepted');
            assert.equal(await page.evaluate(() => window.trialMarker), 'kept-' + width);
            await panel();
            await candidate('broken', true);
            await page.locator('#renderer-try').click();
            await ready('Candidate could not load');
            assert(await page.locator('#renderer-keep').isDisabled(), 'failed import can be kept');
            await page.reload();
            await ready('Using accepted');
            assert.equal(await page.evaluate(() => window.trialMarker), 'kept-' + width);
            await page.goto(base + '?renderers=shipped');
            await ready('Recovery mode');
            assert.equal(await page.evaluate(() => window.trialMarker), undefined, 'recovery mode imported custom code');
            await panel();
            await page.locator('#renderer-rollback').click();
            await ready('Recovery mode');
            await page.goto(base);
            await ready('Using the shipped');
            assert.equal(await page.evaluate(() => window.trialMarker), undefined);
            assert.deepEqual(errors, []);
            console.log(`PASS renderer trial ${width}: inert stage, helpers, try, reload discard, keep, failed import, recovery, rollback`);
            await page.close();
        }
    } finally {
        if (browser) await browser.close();
        await fs.writeFile(path.join(directory, 'stop'), '');
        let timer;
        const code = await Promise.race([exited, new Promise(resolve => { timer = setTimeout(() => resolve('timeout'), 20000); })]);
        clearTimeout(timer);
        if (code === 'timeout') { child.kill('SIGKILL'); await exited; throw new Error('Fixture required forced cleanup'); }
        assert.equal(code, 0, log);
        await fs.rm(directory, { recursive: true, force: true });
        console.log('PASS fixture host joined');
    }
})().catch(error => { console.error(error); process.exitCode = 1; });
