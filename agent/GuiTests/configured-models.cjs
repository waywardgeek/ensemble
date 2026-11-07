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
        const preference = { vendor: 'local', model: 'same-wire' };
        const preferenceKey = 'preference:' + JSON.stringify(preference);
        async function select(endpoint) {
            await page.locator('#settings-open').click();
            const funding = page.getByLabel('Local inference default funding');
            await funding.waitFor();
            await funding.selectOption(endpoint === 'a' ? '0' : '1');
            await page.locator('#funding-choices [role="status"]').filter({ hasText: 'Saved' }).waitFor();
            await page.locator('#settings-close').click();
            await page.waitForFunction(() => !document.querySelector('#model-choice option[data-preference]').disabled);
            await picker.selectOption(preferenceKey);
            await page.waitForFunction(value => {
                const picker = document.querySelector('#model-choice');
                return !picker.disabled && picker.value === value;
            }, preferenceKey);
            assert.deepEqual(commands.at(-1).preference, preference);
            assert.equal(commands.at(-1).offering, undefined);
        }
        assert.equal(await picker.locator('option[data-preference]').count(), 1);
        assert.equal(await picker.locator('option[data-preference]').isDisabled(), true,
            'Ambiguous funding must not invent a default');
        assert.match(await picker.locator('option[data-preference]').getAttribute('title'), /Default funding/);
        assert.equal(await picker.locator('option[data-offering]').count(), 0, 'Model duplicated by funding');
        // A refused automatic save must restore the acknowledged choice and
        // leave the model unavailable; selecting again retries the real route.
        await page.route('**/models/funding', route => route.fulfill({
            status: 409, contentType: 'application/json', body: '{"error":"Test refusal"}'
        }));
        await page.locator('#settings-open').click();
        const funding = page.getByLabel('Local inference default funding');
        await funding.selectOption('0');
        await page.locator('#funding-choices [role="status"]').filter({ hasText: 'Save was not confirmed' }).waitFor();
        assert.equal(await funding.inputValue(), '', 'Refused save left an unconfirmed choice displayed');
        assert.equal(await picker.locator('option[data-preference]').isDisabled(), true);
        assert.equal(await page.getByRole('button', { name: 'Save funding', exact: true }).count(), 0);
        await page.locator('#settings-close').click();
        await page.unroute('**/models/funding');
        await select('a');
        await select('b');
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
        }, preferenceKey);
        assert.match(await picker.locator('option:checked').textContent(), /unavailable/);
        await select('b');
        assert.equal(commands.length, 4, 'Reload or status must not submit an extra model selection');
        await page.getByRole('button', { name: 'Restore stopped-peer', exact: true }).click();
        await page.waitForFunction(() => !document.querySelector('#peer-create').disabled);
        assert.equal(await page.locator('#peer-name').inputValue(), 'stopped-peer');
        assert.equal(await page.locator('#peer-name').getAttribute('readonly'), '');
        assert.equal(await page.locator('#peer-model option[data-preference-key]').count(), 1);
        await page.locator('#peer-model').selectOption(await page.locator('#peer-model option[data-preference-key]').getAttribute('value'));
        await page.getByRole('button', { name: 'Save and retry', exact: true }).click();
        await page.locator('#peer-notice').filter({ hasText: 'Restored Agent stopped-peer' }).waitFor();
        await page.getByRole('button', { name: 'Restore stopped-peer', exact: true }).waitFor({ state: 'detached' });
        console.log('PASS configured model page: same-wire choices, qualified commands, unavailable reload and recovery');
    } finally {
        await browser.close();
        fs.writeFileSync(path.join(directory, 'stop'), 'stop');
    }
})().catch(error => { console.error(error); process.exitCode = 1; });
