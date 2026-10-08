#!/usr/bin/env node
const path=require('path');const {chromium}=require(path.join(process.env.CH07_BROWSER_TOOLS||'/Users/bill/projects/ensemble-edition-2-revisions/browser-tools','node_modules/playwright'));
(async()=>{const [url,key]=process.argv.slice(2),browser=await chromium.launch({channel:'chrome',headless:true}),page=await browser.newPage();let result={};try{
 await page.goto(url);const meta=page.locator(`article[data-key="${key}"] .meta`);await meta.waitFor();result.before=await meta.textContent();if(!result.before.includes('running'))throw Error('positive fixture did not show running report');
 await page.request.post(url+'/review-advance');await page.waitForFunction(key=>document.querySelector(`article[data-key="${key}"] .meta`)?.textContent.includes('Job killed'),key);result.live=await meta.textContent();
 await page.reload();await meta.waitFor();result.reconnected=await meta.textContent();if(!result.reconnected.includes('Job killed'))throw Error('reconnect lost actual killed lifecycle and restored running report');
 await page.reload();await meta.waitFor();result.reconnectedTwice=await meta.textContent();if(!result.reconnectedTwice.includes('Job killed')||await meta.count()!==1)throw Error('second reconnect lost or duplicated killed card');result.passed=true;
 }catch(e){result.passed=false;result.error=e.message;}finally{await browser.close();}console.log(JSON.stringify(result,null,2));process.exitCode=result.passed?0:1;})().catch(e=>{console.error(e);process.exitCode=1;});
