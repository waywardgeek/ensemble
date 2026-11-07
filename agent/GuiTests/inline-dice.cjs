'use strict';
// Run against a copied distribution, not repository renderer source.
const assert = require('node:assert/strict');
const fs = require('node:fs/promises');
const path = require('node:path');
const { chromium } = require('../../edition-2/agent/gui/node_modules/playwright');
const directory = process.argv[2];
assert(directory, 'Pass the distribution skills/shared directory');
(async () => {
 const browser = await chromium.launch({executablePath:'/Applications/Google Chrome.app/Contents/MacOS/Google Chrome',headless:true});
 try {
  for (const mobile of [false,true]) {
   const page = await browser.newPage({viewport:{width:mobile?390:1280,height:844},hasTouch:mobile,isMobile:mobile});
   const errors=[]; page.on('pageerror',e=>errors.push(e.message));
   await page.route('**/*',async route=>{
    const url=new URL(route.request().url());
    if(url.origin!=='http://127.0.0.1:18765' || !['/inline-dice.html','/inline-dice.js'].includes(url.pathname)) return route.abort();
    await route.fulfill({body:await fs.readFile(path.join(directory,url.pathname.slice(1))),contentType:url.pathname.endsWith('.js')?'text/javascript':'text/html'});
   });
   await page.goto('http://127.0.0.1:18765/inline-dice.html');
   const summary=page.locator('summary'), details=page.locator('details');
   await summary.waitFor();
   assert.equal(await summary.textContent(),'Attack: 23');
   assert.equal(await details.getAttribute('open'),null);
   assert.equal(await page.locator('#result').getAttribute('data-artifact-id'),'example-attack-1');
   assert.equal(await page.locator('.card, .tool-card').count(),0);
   await page.evaluate(()=>{window.savedDisclosure=document.querySelector('details'); Math.random=()=>{throw Error('Reroll forbidden');};});
   if(mobile) await summary.tap(); else await summary.click();
   assert.notEqual(await details.getAttribute('open'),null);
   assert.match(await details.textContent(),/d20: 17 \+ proficiency: 3 \+ ability: 3 = 23/);
   if(mobile) await summary.tap(); else await summary.click();
   assert.equal(await details.getAttribute('open'),null);
   await summary.focus(); await page.keyboard.press('Enter');
   assert.notEqual(await details.getAttribute('open'),null);
   await page.keyboard.press('Space');
   assert.equal(await details.getAttribute('open'),null);
   assert(await page.evaluate(()=>savedDisclosure===document.querySelector('details')));
   assert(await summary.evaluate(el=>el.getBoundingClientRect().height>=44));
   assert(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth));
   const result=await page.evaluate(async()=>{
    const {default:render}=await import('./inline-dice.js');
    const record=Object.freeze({label:'<img src=x onerror=alert(1)>',total:15,terms:Object.freeze([
     Object.freeze({label:'<b>d20</b>',value:17}),Object.freeze({label:'penalty',value:-2})])});
    const before=JSON.stringify(record), node=render(document,record);
    document.body.append(node);
    let refused=0;
    for(const bad of [null,{...record,total:16},{...record,terms:[]},{...record,total:NaN},
     {...record,terms:[{label:'overflow',value:Number.MAX_SAFE_INTEGER},{label:'one',value:1}]},
     {...record,terms:[{label:'fraction',value:1.5}]}, {...record,terms:Array(33).fill({label:'x',value:0})}]) {
     try {render(document,bad);} catch {refused++;}
    }
    return {unchanged:before===JSON.stringify(record),htmlNodes:node.querySelectorAll('img,b').length,text:node.textContent,refused};
   });
   assert(result.unchanged); assert.equal(result.htmlNodes,0); assert.equal(result.refused,7);
   assert.match(result.text,/<b>d20<\/b>: 17 \+ penalty: -2 = 15/);
   assert.deepEqual(errors,[]);
   await page.close();
   console.log(`inline dice ${mobile?'touch 390':'desktop 1280'}: recorded facts, keyboard/click/touch, identity, literal labels, bounds/refusal passed`);
  }
 } finally {await browser.close();}
})().catch(error=>{console.error(error);process.exitCode=1;});
