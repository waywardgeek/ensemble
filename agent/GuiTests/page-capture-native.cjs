const fs = require('fs');
const path = require('path');
const http = require('http');
const { chromium } = require('../../edition-2/agent/gui/node_modules/playwright');
const title = 'Owned Ensemble Native Capture Fixture';
const html = `<!doctype html><title>${title}</title><style>html,body{margin:0;width:100%;height:100%;background:linear-gradient(to right,#f00 50%,#00f 50%)}button{position:absolute;top:8px;left:8px}</style><button id="capture">Capture this fixture</button><script type="module">
import {PageCapture} from './page-capture.js';
const capture = new PageCapture();
document.querySelector('#capture').onclick = async () => {
  try {
    const result = await capture.capture();
    const bitmap = await createImageBitmap(result.file);
    const canvas = document.createElement('canvas'); canvas.width = result.width; canvas.height = result.height;
    const ctx = canvas.getContext('2d'); ctx.drawImage(bitmap,0,0); bitmap.close();
    const left = Array.from(ctx.getImageData(Math.floor(result.width/4),Math.floor(result.height/2),1,1).data);
    const right = Array.from(ctx.getImageData(Math.floor(3*result.width/4),Math.floor(result.height/2),1,1).data);
    window.result = {ok:true,width:result.width,height:result.height,bytes:result.file.size,left,right,busy:capture.busy};
    canvas.width=0;
  } catch(error) { window.result = {ok:false,name:error.name,message:error.message,busy:capture.busy}; }
};</script>`;
const server = http.createServer((req,res) => {
  if (req.url === '/') { res.setHeader('Content-Type','text/html'); res.end(html); }
  else if (req.url === '/page-capture.js') {res.setHeader('Content-Type','text/javascript');res.end(fs.readFileSync(path.resolve(__dirname,'../gui/wwwroot/page-capture.js')));}
  else {res.statusCode=404;res.end();}
});
(async()=>{
 let browser;
 try {
  await new Promise(resolve=>server.listen(0,'127.0.0.1',resolve));
  browser=await chromium.launch({headless:true,channel:'chrome',args:['--auto-accept-this-tab-capture','--autoplay-policy=no-user-gesture-required','--mute-audio']});
  const page=await browser.newPage({viewport:{width:640,height:360}});
  await page.goto(`http://127.0.0.1:${server.address().port}/`);
  await page.click('#capture');
  await page.waitForFunction(()=>!!window.result,{},{timeout:15000});
  const result=await page.evaluate(()=>window.result);
  console.log(JSON.stringify(result,null,2));
  if(!result.ok || result.busy || result.left[0]<240 || result.left[2]>15 || result.right[2]<240 || result.right[0]>15)
    throw new Error('Native fixture capture did not establish correct pixels and cleanup.');
  console.log('PASS native this-tab capture, verified handle/origin, actual red/blue pixels and closed capture owner');
 } finally {await browser?.close();await new Promise(resolve=>server.close(resolve));}
})().catch(error=>{console.error(error);process.exitCode=1;});
