// Optional print backend. Uses explicitly supplied existing tools; installs nothing.
'use strict';
const { pathToFileURL } = require('node:url');
const [modulePath, executablePath, html, pdf] = process.argv.slice(2);
if (!modulePath || !executablePath || !html || !pdf) {
  throw new Error('Expected Playwright module, Chrome, HTML, and PDF paths');
}
const { chromium } = require(modulePath);
(async () => {
  let browser;
  try {
    browser = await chromium.launch({ executablePath, headless: true });
    const page = await browser.newPage();
    await page.goto(pathToFileURL(html).href, { waitUntil: 'load' });
    await page.evaluate(() => document.fonts.ready);
    await page.pdf({ path: pdf, preferCSSPageSize: true,
      printBackground: true, displayHeaderFooter: false });
  } finally {
    if (browser) await browser.close();
  }
})().catch(error => { console.error(error); process.exitCode = 1; });
