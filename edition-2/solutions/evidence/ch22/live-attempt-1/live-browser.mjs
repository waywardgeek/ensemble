// A bounded human interaction with the actual browser application. No injected
// responses or internal Agent calls: use the visible prompt and model controls.
import { chromium } from '../../../../solutions/edition-2/main/gui/node_modules/playwright/index.mjs';
const browser = await chromium.launch({ headless: true });
try {
  const page = await browser.newPage();
  page.setDefaultTimeout(90000);
  await page.goto(process.argv[2] + '?speech=transcript');
  await page.getByRole('tab', { name: 'Settings' }).click();
  await page.waitForFunction(() => document.querySelector('#model').value === 'gpt-6-astra');
  const prompts = [
    'Call agent_status exactly once and report its result verbatim. Do not use any other tools. Explain briefly whether the dollar cost is known. This is a status demonstration, not a coding task.',
    'The human changed the selected model between turns. Call agent_status exactly once and report its result verbatim. State whether session usage includes earlier model work and whether dollars are known. Do not use other tools.'
  ];
  for (let index = 0; index < prompts.length; index++) {
    if (index) {
      await page.getByRole('tab', { name: 'Settings' }).click();
      await page.locator('#model').fill('gpt-6.1-sol');
      await page.locator('#model').press('Tab');
      await page.waitForFunction(() => document.querySelector('#settings-status').textContent === '');
    }
    await page.getByRole('tab', { name: 'Agents', exact: true }).click();
    await page.locator('#input').fill(prompts[index]);
    await page.getByRole('button', { name: 'Send', exact: true }).click();
    await page.waitForFunction(n => window.ensembleTurn >= n, index + 1);
  }
  console.log(JSON.stringify({ turns: await page.evaluate(() => window.ensembleTurn), model: await page.locator('#model').inputValue(), transcript: await page.locator('body').innerText() }, null, 2));
} finally { await browser.close(); }
