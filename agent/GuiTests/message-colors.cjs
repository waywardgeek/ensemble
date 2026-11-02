'use strict';
// Real CSS, isolated DOM: no application startup, network provider, or capture.
const assert = require('node:assert/strict');
const path = require('node:path');
const { chromium } = require('../../edition-2/agent/gui/node_modules/playwright');
const root = path.resolve(__dirname, '../Gui/wwwroot');
function luminance(rgb) {
    const channels = rgb.match(/[\d.]+/g).slice(0, 3).map(Number).map(value => {
        value /= 255;
        return value <= .04045 ? value / 12.92 : ((value + .055) / 1.055) ** 2.4;
    });
    return channels[0] * .2126 + channels[1] * .7152 + channels[2] * .0722;
}
function contrast(a, b) {
    const values = [luminance(a), luminance(b)].sort((x, y) => y - x);
    return (values[0] + .05) / (values[1] + .05);
}
(async () => {
    const browser = await chromium.launch({ executablePath: '/Applications/Google Chrome.app/Contents/MacOS/Google Chrome', headless: true });
    try {
        const page = await browser.newPage();
        await page.setContent('<main id="transcript">' + ['human', 'assistant'].map((role, i) =>
            `<article class="part" data-kind="text" data-role="${role}"><header>${i ? 'Cadence' : 'Bill'}</header><div class="message"><p>Readable prose <a href="#">and link</a></p></div></article>`).join('') + '</main>');
        await page.addStyleTag({ path: path.join(root, 'vendor/bootstrap-5.3.8.min.css') });
        await page.addStyleTag({ path: path.join(root, 'style.css') });
        for (const [theme, expected] of Object.entries({
            dark: { human: 'rgb(12, 32, 48)', assistant: 'rgb(31, 37, 46)' },
            light: { human: 'rgb(196, 221, 247)', assistant: 'rgb(255, 255, 255)' }
        })) {
            await page.evaluate(theme => { document.documentElement.dataset.bsTheme = theme; }, theme);
            for (const role of ['human', 'assistant']) {
                const colors = await page.locator(`[data-role="${role}"]`).evaluate(article => {
                    const message = article.querySelector('.message');
                    const label = article.querySelector('header');
                    return { background: getComputedStyle(message).backgroundColor,
                        text: getComputedStyle(message).color,
                        link: getComputedStyle(message.querySelector('a')).color,
                        label: label.textContent, labelHeight: label.getBoundingClientRect().height };
                });
                assert.equal(colors.background, expected[role], `${theme}/${role}: exact background swap`);
                assert.ok(contrast(colors.text, colors.background) >= 4.5, `${theme}/${role}: text contrast`);
                assert.ok(contrast(colors.link, colors.background) >= 4.5, `${theme}/${role}: link contrast`);
                assert.equal(colors.label, role === 'human' ? 'Bill' : 'Cadence');
                assert.ok(colors.labelHeight > 0, 'non-color speaker identity remains visible');
                console.log(`PASS ${theme}/${role} background, contrast, speaker label`);
            }
        }
    } finally { await browser.close(); }
})().catch(error => { console.error(error); process.exitCode = 1; });
