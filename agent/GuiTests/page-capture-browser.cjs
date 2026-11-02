// Full-page capture-to-draft checks. Requires an isolated --crossover-host.
const { chromium } = require('../../edition-2/agent/gui/node_modules/playwright');
const assert = require('node:assert/strict');
const base = process.env.GUI_URL || 'http://127.0.0.1:18202/assistant/';
const url = new URL(base);
assert.equal(url.hostname, '127.0.0.1');
assert.ok(!['8088', '8089'].includes(url.port), 'Never exercise live hosts');
(async () => {
  const browser = await chromium.launch({ headless:true, channel:'chrome', args:['--autoplay-policy=no-user-gesture-required'] });
  const page = await browser.newPage();
  page.setDefaultTimeout(15000);
  const errors=[];
  page.on('pageerror', e => errors.push(e.message));
  try {
    await page.addInitScript(() => {
      window.captureMode='normal'; window.captureTracks=[]; window.captureRequests=0;
      window.imageSends=[]; window.imageAcks=[]; window.imageRefusals=[]; window.textSends=[];
      const NativeSocket=WebSocket;
      window.WebSocket=class extends NativeSocket {
        constructor(...args) {
          super(...args); window.fixtureSocket=this;
          this.addEventListener('message', event => {
            const frame=JSON.parse(event.data);
            if(frame.type==='image_accepted') imageAcks.push(frame);
            if(frame.type==='image_refused' || frame.type==='error') imageRefusals.push(frame);
          });
        }
        send(data) {
          const frame=JSON.parse(data);
          if(frame.type==='image_prompt') imageSends.push(frame);
          if(frame.type==='prompt') textSends.push(frame);
          return super.send(data);
        }
      };
      let identity;
      Object.defineProperty(navigator.mediaDevices,'setCaptureHandleConfig',{configurable:true,value:config=>{identity=config;}});
      window.fakeDisplay=()=>{
        captureRequests++;
        const canvas=document.createElement('canvas'); canvas.width=64; canvas.height=32;
        const context=canvas.getContext('2d'); context.fillStyle='red'; context.fillRect(0,0,32,32);
        context.fillStyle='blue'; context.fillRect(32,0,32,32);
        const stream=canvas.captureStream(); const track=stream.getVideoTracks()[0];
        captureTracks.push(track);
        const selected=identity.handle;
        Object.defineProperty(track,'getSettings',{value:()=>({displaySurface:'browser'})});
        Object.defineProperty(track,'getCaptureHandle',{value:()=>({handle:selected,origin:location.origin})});
        if(captureMode==='hold') return new Promise(resolve=>{window.releaseCapture=()=>resolve(stream);});
        return Promise.resolve(stream);
      };
      Object.defineProperty(navigator.mediaDevices,'getDisplayMedia',{configurable:true,value:fakeDisplay});
    });
    if(process.env.CAPTURE_UI_MUTANT==='selection') {
      await page.route('**/app.js',async route=>{
        const response=await route.fetch();let source=await response.text();
        const hook='cancelPageCapture("Page capture cancelled on Agent change.");';
        assert.equal(source.split(hook).length-1,2,'mutation must hit both selection hooks');
        source=source.split(hook).join('/* test-only missing selection invalidation */');
        await route.fulfill({response,body:source});
      });
    }
    await page.goto(base);
    await page.waitForFunction(()=>!document.querySelector('#capture-page').disabled);
    await page.evaluate(()=>{if(document.querySelector('#model-choice').value!=='claude-3-haiku-20240307') send('set_model','claude-3-haiku-20240307');});
    await page.waitForFunction(()=>document.querySelector('#model-choice').value==='claude-3-haiku-20240307' && !document.querySelector('#model-choice').disabled);
    const original=await page.evaluate(()=>agentID);
    await page.locator('#prompt').fill('Synthetic screenshot caption');
    await page.locator('#capture-page').click();
    await page.waitForFunction(()=>imageComposer.hasImages && !imageComposer.blocked && !pageCapture.busy);
    assert.equal(await page.locator('#prompt').inputValue(),'Synthetic screenshot caption');
    assert.equal(await page.evaluate(()=>captureTracks.every(t=>t.readyState==='ended')),true);
    const pixels=await page.evaluate(async()=>{
      const image=await createImageBitmap(imageComposer.entries[0].file);
      const canvas=document.createElement('canvas');canvas.width=image.width;canvas.height=image.height;
      const context=canvas.getContext('2d');context.drawImage(image,0,0);image.close();
      return {width:canvas.width,height:canvas.height,left:[...context.getImageData(8,16,1,1).data],right:[...context.getImageData(56,16,1,1).data]};
    });
    assert.equal(pixels.width,64); assert.equal(pixels.height,32);
    // Video color conversion may shift low channels by one or two levels.
    assert.ok(pixels.left[0]>=250 && pixels.left[1]<=4 && pixels.left[2]<=4 && pixels.left[3]===255);
    assert.ok(pixels.right[2]>=250 && pixels.right[0]<=4 && pixels.right[1]<=4 && pixels.right[3]===255);
    await page.locator('#send').click();
    await page.waitForFunction(()=>imageAcks.length===1 || imageRefusals.length>0);
    assert.deepEqual(await page.evaluate(()=>imageRefusals),[]);
    assert.equal(await page.evaluate(()=>imageSends.length),1);
    assert.equal(await page.evaluate(()=>imageSends[0].text),'Synthetic screenshot caption');
    assert.equal(await page.evaluate(()=>imageSends[0].agent_id),original);
    await page.waitForFunction(()=>!imageComposer.hasImages && !document.querySelector('#capture-page').disabled);
    console.log('PASS actual screenshot pixels + caption + addressed server admission');

    async function heldCapture() {
      await page.evaluate(()=>{captureMode='hold';releaseCapture=null;});
      await page.locator('#capture-page').click();
      await page.waitForFunction(()=>typeof releaseCapture==='function');
      assert.equal(await page.locator('#send').isDisabled(),true);
      assert.equal(await page.locator('#capture-page').getAttribute('aria-label'),'Cancel page capture');
    }
    async function releaseWithoutAdmission(message='Cancelled capture must not enter draft') {
      await page.evaluate(()=>releaseCapture());
      await page.waitForFunction(()=>!pageCapture.busy);
      assert.equal(await page.evaluate(()=>imageComposer.hasImages),false,message);
      assert.equal(await page.evaluate(()=>captureTracks.every(t=>t.readyState==='ended')),true);
      assert.equal(await page.evaluate(()=>imageSends.length),1);
    }
    await heldCapture();
    await page.locator('#prompt').fill('Wait for capture before sending this text');
    await page.evaluate(()=>document.querySelector('#compose').requestSubmit());
    assert.equal(await page.evaluate(()=>textSends.length),0);
    assert.equal(await page.locator('#prompt').inputValue(),'Wait for capture before sending this text');
    await page.locator('#capture-page').click();
    await releaseWithoutAdmission();
    await page.locator('#prompt').fill('');
    console.log('PASS visible cancel + late picker result inert + pending Send blocked');

    const other=await page.evaluate(()=>Array.from(document.querySelector('#agent-choice').options).find(o=>o.value && o.value!==agentID && !o.disabled)?.value);
    assert.ok(other,'fixture must offer another Agent');
    await heldCapture();
    await page.evaluate(key=>selectAgent(key),other);
    await page.waitForFunction(key=>agentID===key && !selectionPending,other);
    await page.evaluate(key=>selectAgent(key),original);
    await page.waitForFunction(key=>agentID===key && !selectionPending,original);
    await releaseWithoutAdmission('Agent change must invalidate old capture even after selecting back');
    console.log('PASS switch away and back invalidates old capture');

    await heldCapture();
    await page.evaluate(()=>fixtureSocket.close());
    await page.waitForFunction(()=>captureDraft===null);
    await releaseWithoutAdmission();
    await page.waitForFunction(()=>!document.querySelector('#capture-page').disabled);
    console.log('PASS disconnect/reconnect does not admit or resend capture');

    await heldCapture();
    const previousState=await page.evaluate(()=>acknowledgedState);
    await page.evaluate(()=>handle({type:'agent_state',state:'shutdown'}));
    assert.equal(await page.evaluate(()=>captureDraft===null),true);
    await releaseWithoutAdmission();
    await page.evaluate(state=>handle({type:'agent_state',state}),previousState);
    await page.waitForFunction(()=>!document.querySelector('#capture-page').disabled);
    console.log('PASS controlled shutdown frame cancels capture immediately');

    const requests=await page.evaluate(()=>captureRequests);
    await page.evaluate(()=>Object.defineProperty(navigator.mediaDevices,'getDisplayMedia',{configurable:true,value:undefined}));
    await page.locator('#capture-page').click();
    await page.waitForFunction(()=>document.querySelector('#notice').textContent.includes('attach the image'));
    assert.equal(await page.evaluate(()=>captureRequests),requests);
    assert.equal(await page.evaluate(()=>imageComposer.hasImages),false);
    await page.evaluate(()=>Object.defineProperty(navigator.mediaDevices,'getDisplayMedia',{configurable:true,value:fakeDisplay}));
    console.log('PASS unsupported browser explains manual attachment without capture');

    for(const width of [393,1280]) {
      await page.setViewportSize({width,height:851});
      const sizes=await page.evaluate(()=>{
        const camera=document.querySelector('#capture-page');const input=document.querySelector('#prompt');
        const shown=input.getBoundingClientRect().width;
        camera.hidden=true;const hidden=input.getBoundingClientRect().width;camera.hidden=false;
        return {shown,hidden,header:!!camera.closest('header'),inComposer:!!camera.closest('#compose')};
      });
      assert.equal(sizes.header,true);assert.equal(sizes.inComposer,false);
      assert.ok(Math.abs(sizes.shown-sizes.hidden)<1,'camera must not consume composer width');
    }
    assert.deepEqual(errors,[]);
    console.log('PASS phone/desktop camera placement preserves input width; no page errors');
  } finally { await browser.close(); }
})().catch(error=>{console.error(error);process.exitCode=1;});
