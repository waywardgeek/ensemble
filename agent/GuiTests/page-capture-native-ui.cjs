// Native this-tab capture through the real GUI, on an owned synthetic host only.
const { chromium } = require('../../edition-2/agent/gui/node_modules/playwright');
const assert = require('node:assert/strict');
const base=process.env.GUI_URL || 'http://127.0.0.1:18202/assistant/';
const url=new URL(base);
assert.equal(url.hostname,'127.0.0.1');
assert.ok(!['8088','8089'].includes(url.port),'Never capture a live host');
(async()=>{
  const browser=await chromium.launch({headless:true,channel:'chrome',args:['--auto-accept-this-tab-capture','--autoplay-policy=no-user-gesture-required','--mute-audio']});
  const page=await browser.newPage({viewport:{width:960,height:640}});
  page.setDefaultTimeout(15000);
  const errors=[];page.on('pageerror',error=>errors.push(error.message));
  try {
    await page.addInitScript(()=>{
      window.imageSends=[];window.imageAcks=[];window.imageErrors=[];
      const NativeSocket=WebSocket;
      window.WebSocket=class extends NativeSocket {
        constructor(...args){super(...args);this.addEventListener('message',event=>{
          const frame=JSON.parse(event.data);
          if(frame.type==='image_accepted') imageAcks.push(frame);
          if(frame.type==='error') imageErrors.push(frame);
        });}
        send(data){const frame=JSON.parse(data);if(frame.type==='image_prompt')imageSends.push(frame);return super.send(data);}
      };
    });
    await page.goto(base);
    await page.waitForFunction(()=>!document.querySelector('#capture-page').disabled);
    await page.evaluate(()=>{if(document.querySelector('#model-choice').value!=='claude-3-haiku-20240307') send('set_model','claude-3-haiku-20240307');});
    await page.waitForFunction(()=>document.querySelector('#model-choice').value==='claude-3-haiku-20240307' && !document.querySelector('#model-choice').disabled);
    const original=await page.evaluate(()=>agentID);
    await page.locator('#prompt').fill('Synthetic native tab screenshot');
    await page.evaluate(()=>{
      for(const [side,color] of [['left','red'],['right','blue']]) {
        const marker=document.createElement('div');
        marker.style.cssText=`position:fixed;${side}:0;top:70px;width:20px;height:20px;background:${color};z-index:2147483647;pointer-events:none`;
        document.body.append(marker);
      }
      return new Promise(resolve=>requestAnimationFrame(()=>requestAnimationFrame(resolve)));
    });
    await page.locator('#capture-page').click();
    await page.waitForFunction(()=>imageComposer.hasImages && !imageComposer.blocked && !pageCapture.busy);
    const pixels=await page.evaluate(async()=>{
      const file=imageComposer.entries[0].file;
      const bitmap=await createImageBitmap(file);
      const canvas=document.createElement('canvas');canvas.width=bitmap.width;canvas.height=bitmap.height;
      const context=canvas.getContext('2d');context.drawImage(bitmap,0,0);bitmap.close();
      return {width:canvas.width,height:canvas.height,bytes:file.size,left:[...context.getImageData(10,80,1,1).data],right:[...context.getImageData(canvas.width-10,80,1,1).data]};
    });
    assert.equal(pixels.width,960);assert.equal(pixels.height,640);assert.ok(pixels.bytes>0);
    assert.ok(pixels.left[0]>=250 && pixels.left[1]<=4 && pixels.left[2]<=4,'actual red marker');
    assert.ok(pixels.right[2]>=250 && pixels.right[0]<=4 && pixels.right[1]<=4,'actual blue marker');
    await page.locator('#send').click();
    await page.waitForFunction(()=>imageAcks.length===1 || imageErrors.length>0);
    assert.deepEqual(await page.evaluate(()=>imageErrors),[]);
    assert.equal(await page.evaluate(()=>imageSends.length),1);
    assert.equal(await page.evaluate(()=>imageSends[0].agent_id),original);
    assert.equal(await page.evaluate(()=>imageSends[0].text),'Synthetic native tab screenshot');
    await page.waitForFunction(()=>!imageComposer.hasImages);
    assert.deepEqual(errors,[]);
    console.log('PASS native camera click -> actual tab pixels -> preview -> caption/image admission',JSON.stringify(pixels));
  } finally {await browser.close();}
})().catch(error=>{console.error(error);process.exitCode=1;});
