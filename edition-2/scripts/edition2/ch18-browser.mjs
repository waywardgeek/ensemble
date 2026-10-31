import { chromium } from '../../solutions/edition-2/main/gui/node_modules/playwright/index.mjs';
import { writeFileSync } from 'node:fs';
const [url, destination] = process.argv.slice(2);
const browser = await chromium.launch({headless:true,executablePath:'/Applications/Google Chrome.app/Contents/MacOS/Google Chrome'});
const page = await browser.newPage();
const receipts=[];
try {
 await page.goto(url+'?speech=transcript');
 await page.waitForFunction(()=>document.querySelector('#state').textContent!=='connecting');
 receipts.push({initial:await page.locator('#usage').innerText()});
 for (const prompt of ['What is the archive project code? Reply in one short sentence.','What is the rule for archive record 37? Reply in one short sentence.','What color identifies archive record 402? Reply in one short sentence.','Remind me of the archive project code, in five words or fewer.']) {
  const turn=await page.evaluate(()=>window.ensembleTurn);
  await page.locator('#input').fill(prompt);
  await page.locator('#input').press('Enter');
  await page.waitForFunction(n=>window.ensembleTurn>n,turn,{timeout:85000});
  receipts.push({prompt,meter:await page.locator('#usage').innerText(),tooltip:await page.locator('#usage').getAttribute('title'),conversation:await page.locator('#artifacts').innerText()});
  writeFileSync(destination,JSON.stringify(receipts,null,2));
 }
 await page.screenshot({path:destination.replace('.json','.png'),fullPage:true});
} catch(error) {
 receipts.push({error:String(error)});
 writeFileSync(destination,JSON.stringify(receipts,null,2));
 process.exitCode=1;
} finally { await browser.close(); }
