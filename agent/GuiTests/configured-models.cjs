const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const { chromium } = require(process.env.PLAYWRIGHT_MODULE);
const directory = process.argv[2];
const { origin } = JSON.parse(fs.readFileSync(path.join(directory, 'ready.json'), 'utf8'));
const identity = endpoint => JSON.stringify(['local', endpoint, 'small']);
(async () => {
    const browser = await chromium.launch({ channel: 'chrome', headless: true });
    try {
        const page = await browser.newPage();
        const commands = [];
        page.on('websocket', socket => socket.on('framesent', frame => {
            try {
                const value = JSON.parse(frame.payload);
                if (value.type === 'set_model') commands.push(value);
            } catch {}
        }));
        await page.goto(origin);
        const picker = page.locator('#model-choice');
        await picker.waitFor();
        await page.waitForFunction(() => !document.querySelector('#model-choice').disabled);
        async function select(endpoint) {
            await picker.selectOption(identity(endpoint));
            await page.waitForFunction(value => {
                const picker = document.querySelector('#model-choice');
                return !picker.disabled && picker.value === value;
            }, identity(endpoint));
            assert.deepEqual(commands.at(-1).offering, { vendor: 'local', endpoint, model: 'small' });
            assert.equal(commands.at(-1).text, '');
        }
        assert.equal(await picker.locator('option[data-preference]').count(), 1);
        assert.equal(await picker.locator('option[data-preference]').isDisabled(), true,
            'Ambiguous funding must not invent a default');
        assert.match(await picker.locator('option[data-preference]').getAttribute('title'), /Remembered funding/);
        await select('a');
        await select('b');
        const preference = { vendor: 'local', model: 'same-wire' };
        await picker.selectOption('preference:' + JSON.stringify(preference));
        await page.waitForFunction(value => {
            const picker = document.querySelector('#model-choice');
            return !picker.disabled && picker.value === value;
        }, identity('b'));
        assert.deepEqual(commands.at(-1).preference, preference);
        assert.equal(commands.at(-1).offering, undefined);
        assert.equal(commands.at(-1).text, '');
        assert.equal(await picker.locator('option[data-offering]').count(), 2);
        const labels = await picker.locator('option[data-offering]').allTextContents();
        assert(labels.every(label => label.includes('serving context unknown')), 'Serving context must not be invented');
        await select('a');
        fs.writeFileSync(path.join(directory, 'remove-a'), 'remove');
        const deadline = Date.now() + 5000;
        while (!fs.existsSync(path.join(directory, 'removed-a'))) {
            assert(Date.now() < deadline, 'Fixture did not remove selected Model');
            await new Promise(resolve => setTimeout(resolve, 20));
        }
        await page.reload();
        await page.waitForFunction(value => {
            const picker = document.querySelector('#model-choice');
            return picker && !picker.disabled && picker.value === value;
        }, 'unavailable:' + identity('a'));
        assert.equal(await picker.locator('option:checked').isDisabled(), true);
        assert.match(await picker.locator('option:checked').textContent(), /unavailable/);
        await select('b');
        assert.equal(commands.length, 5, 'Reload or status must not submit an extra model selection');
        console.log('PASS configured model page: same-wire choices, qualified commands, unavailable reload and recovery');
    } finally {
        await browser.close();
        fs.writeFileSync(path.join(directory, 'stop'), 'stop');
    }
})().catch(error => { console.error(error); process.exitCode = 1; });
