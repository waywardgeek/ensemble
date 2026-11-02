// Native capture -> real Agent tool result -> loopback model -> History replay.
const {chromium}=require('../../edition-2/agent/gui/node_modules/playwright');
const fs=require('node:fs');const path=require('node:path');const assert=require('node:assert/strict');
const home=process.env.CAPTURE_FIXTURE_DIR;assert.ok(home,'CAPTURE_FIXTURE_DIR required');
const base='http://127.0.0.1:18202/assistant/';
(async()=>{
  const browser=await chromium.launch({headless:true,channel:'chrome',args:['--auto-accept-this-tab-capture','--autoplay-policy=no-user-gesture-required','--mute-audio']});
  const page=await browser.newPage({viewport:{width:960,height:640}});page.setDefaultTimeout(15000);
  const errors=[];page.on('pageerror',error=>errors.push(error.message));
  const status=()=>JSON.parse(fs.readFileSync(path.join(home,'status.json'),'utf8'));
  async function until(check,label){
    const end=Date.now()+40000;
    while(Date.now()<end){const current=status();if(current.failure)throw new Error(current.failure);if(await check())return;await page.waitForTimeout(50);}
    throw new Error(`${label}: ${JSON.stringify(status())}`);
  }
  try{
    await page.addInitScript(()=>{
      window.nativeCaptureCount=0;window.nativeTracks=[];
      const native=navigator.mediaDevices.getDisplayMedia.bind(navigator.mediaDevices);
      navigator.mediaDevices.getDisplayMedia=async options=>{
        nativeCaptureCount++;
        const stream=await native(options);nativeTracks.push(...stream.getTracks());return stream;
      };
    });
    await page.goto(base);
    await until(()=>page.evaluate(()=>typeof captureConsent!=='undefined'&&captureConsent&&!selectionPending&&!document.querySelector('#send').disabled),'page readiness');
    const declaration=await page.evaluate(()=>browserTools.find(tool=>tool.name==='gui_capture'));
    assert.ok(declaration);assert.deepEqual(declaration.inputSchema.properties,{});
    await page.evaluate(()=>{
      for(const [side,color] of [['left','red'],['right','blue']]){
        const marker=document.createElement('div');
        marker.style.cssText=`position:fixed;${side}:0;top:72px;width:32px;height:32px;background:${color};z-index:2147483647;pointer-events:none`;
        document.body.append(marker);
      }
      return new Promise(resolve=>requestAnimationFrame(()=>requestAnimationFrame(resolve)));
    });
    await page.locator('#prompt').fill('Request a screenshot of this synthetic Ensemble page.');
    await page.locator('#send').click();
    await until(()=>page.locator('#capture-consent').evaluate(dialog=>dialog.open),'model-requested consent');
    assert.equal(await page.evaluate(()=>nativeCaptureCount),0,'Agent request must not start native capture');
    await page.locator('#capture-share').click();
    await until(()=>status().stage>=3,'first model image request');
    await until(()=>page.locator('#transcript').innerText().then(text=>text.includes('Capture verified.')),'first final answer');
    assert.equal(await page.evaluate(()=>nativeCaptureCount),1);
    assert.equal(await page.evaluate(()=>nativeTracks.length===1&&nativeTracks.every(track=>track.readyState==='ended')),true);
    assert.equal(await page.evaluate(()=>imageComposer.hasImages),false,'tool capture must not become a human attachment');
    await page.locator('#prompt').fill('Verify the retained screenshot again without taking another capture.');
    await page.locator('#send').click();
    await until(()=>status().stage===4,'capture replay');
    await until(()=>page.locator('#transcript').innerText().then(text=>text.includes('Capture replay verified.')),'replay final answer');
    assert.equal(await page.evaluate(()=>nativeCaptureCount),1,'replay must not recapture');
    assert.deepEqual(errors,[]);assert.equal(status().requests,4);assert.ok(status().imageBytes>0);
    console.log('PASS native trusted consent, published skill/tool, real Agent/HTTP pixels, canonical History and replay; '+JSON.stringify(status()));
  }finally{await browser.close();}
})().catch(error=>{console.error(error);process.exitCode=1;});
