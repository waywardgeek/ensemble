// Real embedded GUI and sockets, isolated browser profiles; no microphone use.
const { chromium } = require(process.env.PLAYWRIGHT_MODULE || 'playwright');
const assert = require('node:assert/strict');
(async () => {
 const browser = await chromium.launch({executablePath:'/Applications/Google Chrome.app/Contents/MacOS/Google Chrome',headless:true});
 try {
  const context = await browser.newContext();
  // Negative control supplies the old app only inside this disposable browser.
  if (process.env.GUI_APP_OVERRIDE) await context.route('**/app.js', route =>
    route.fulfill({contentType:'text/javascript',body:require('node:fs').readFileSync(process.env.GUI_APP_OVERRIDE,'utf8')}));
  await context.addInitScript(() => {
   const original = WebSocket.prototype.send;
   window.settingWrites = [];
   WebSocket.prototype.send = function(text) { try { const v=JSON.parse(text); if(v.type==='update_settings') window.settingWrites.push(v); } catch {} return original.call(this,text); };
  });
  const open = async c => {const p=await c.newPage(); await p.goto('http://127.0.0.1:18202/assistant/'); await p.waitForFunction(()=>typeof browserDevice!=='undefined' && settings && voiceControl);return p;};
  const page=await open(context), tab=await open(context);
  assert.equal(await page.locator('#voice-shortcut').inputValue(),'off');
  assert.deepEqual(await page.evaluate(() => browserDevice.read('{"font":999,"shortcut":"bad"}')), {font:32,shortcut:'off'});
  assert.deepEqual(await page.evaluate(() => browserDevice.read('{"font":1,"shortcut":"F8"}')), {font:12,shortcut:'F8'});
  assert.equal(await page.evaluate(() => browserDevice.read('{"font":18.5}').font), null);
  await page.evaluate(()=>{window.shortcutApplied=[];const original=voiceControl.setShortcut.bind(voiceControl);voiceControl.setShortcut=v=>{window.shortcutApplied.push(v);original(v);};});
  await page.locator('#settings-open').click(); await page.locator('#settings-display-tab').click();
  await page.locator('#font-size').fill('26'); await page.locator('#font-size').dispatchEvent('change');
  await page.locator('#voice-shortcut').selectOption('F8');
  await tab.waitForFunction(()=>browserDevice.value.font===26 && browserDevice.value.shortcut==='F8');
  assert.equal(await page.evaluate(()=>getComputedStyle(document.documentElement).fontSize),'26px');
  assert.equal(await page.evaluate(()=>shortcutApplied.at(-1)),'F8');
  assert.equal(await page.locator('#voice-hold').getAttribute('aria-keyshortcuts'),'F8');
  assert.deepEqual(await page.evaluate(()=>settingWrites),[]);
  await page.reload(); await page.waitForFunction(()=>settings && voiceControl);
  assert.equal(await page.locator('#font-size').inputValue(),'26');
  await page.evaluate(()=>{settings={...settings,font_size:12,voice_shortcut:'AltRight'};applySettings();});
  assert.equal(await page.locator('#voice-shortcut').inputValue(),'F8');
  assert.equal(await page.locator('#font-size').inputValue(),'26');
  const other=await browser.newContext(); const independent=await open(other);
  assert.equal(await independent.locator('#voice-shortcut').inputValue(),'off');
  assert.notEqual(await independent.locator('#font-size').inputValue(),'26');
  await context.setOffline(true); await page.evaluate(()=>socket.close());
  await page.evaluate(()=>browserDevice.select({font:28,shortcut:'off'}));
  assert.equal(await page.evaluate(()=>getComputedStyle(document.documentElement).fontSize),'28px');
  await context.setOffline(false);
  await page.evaluate(()=>localStorage.removeItem('ensemble.device.v1'));
  await tab.waitForFunction(()=>browserDevice.value.font===18 && browserDevice.value.shortcut==='off');
  await tab.evaluate(()=>{settings={...settings,font_size:31};applySettings();});
  assert.equal(await tab.locator('#font-size').inputValue(),'18');
  const blocked=await browser.newContext();await blocked.addInitScript(()=>{Object.defineProperty(window,'localStorage',{get(){throw new DOMException('blocked','SecurityError');}});});
  const fallback=await open(blocked);await fallback.evaluate(()=>browserDevice.select({font:25,shortcut:'F8'}));
  await fallback.evaluate(()=>{settings={...settings,font_size:12,voice_shortcut:'off'};applySettings();});
  assert.equal(await fallback.locator('#font-size').inputValue(),'25');assert.equal(await fallback.locator('#voice-shortcut').inputValue(),'F8');
  console.log('PASS profile isolation, tab sync, real font/controller, reload, server-frame precedence, offline, clear, blocked storage, zero backend writes');
  await blocked.close();await other.close();await context.close();
 } finally {await browser.close();}
})().catch(error=>{console.error(error);process.exitCode=1;});
