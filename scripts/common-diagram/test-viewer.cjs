// Use an explicitly supplied existing Playwright installation; never install one.
const { chromium } = require(process.argv[2]);
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
(async () => {
  const root = path.resolve(__dirname, '../../docs/common-diagrams');
  const evidence = path.join(__dirname, '.build');
  const browser = await chromium.launch({channel:'chrome',headless:true});
  try {
    const page = await browser.newPage({viewport:{width:1500,height:1000}});
    const errors = [];
    page.on('pageerror', e => errors.push(e.message));
    await page.route(/^https?:/, route => route.abort());
    await page.goto('file://' + path.join(root,'index.html'));
    const relationships = JSON.parse(fs.readFileSync(path.join(__dirname,'relationships.json'),'utf8')).relationships;
    const expected = new Set(relationships.flatMap(e => [e.owner,e.child])).size;
    assert.equal(await page.locator('g.node').count(), expected);
    await page.locator('#search').fill('Model');
    await page.locator('#focus').click();
    assert.match(await page.locator('#details').innerText(), /selectedModel/);
    assert.match(await page.locator('#details').innerText(), /weak/);
    await page.locator('#view').selectOption('all');
    assert.equal(await page.locator('g.node').count(),156);
    await page.locator('#search').fill('MemoryStage');
    await page.locator('#search').press('Enter');
    assert.match(await page.locator('#details').innerText(), /MemoryCapture/);
    const before=await page.locator('#diagram svg').evaluate(e=>e.getBoundingClientRect().width);
    await page.locator('#plus').click();
    const after=await page.locator('#diagram svg').evaluate(e=>e.getBoundingClientRect().width);
    assert.ok(after>before);
    for (const name of ['history','models','memory','tools','ownership','context','contents']) {
      await page.locator('#view').selectOption(name);
      assert.ok(await page.locator('g.node').count()>0);
    }
    await page.locator('#view').selectOption('context');
    await page.locator('#search').fill('Context');
    await page.locator('#focus').click();
    assert.match(await page.locator('#details').innerText(), /Dialogue/);
    assert.match(await page.locator('#details').innerText(), /History/);
    await page.locator('#fit').click();
    await page.screenshot({path:path.join(evidence,'viewer-desktop.png')});
    await page.setViewportSize({width:430,height:900});
    await page.locator('#fit').click();
    await page.screenshot({path:path.join(evidence,'viewer-phone.png')});
    assert.deepEqual(errors,[]);
    console.log('PASS: eight diagrams, type focus and source details, 156 full-graph nodes, zoom, desktop and phone rendering; no network or application connection.');
  } finally {
    await browser.close();
  }
})().catch(e=>{console.error(e);process.exitCode=1;});
