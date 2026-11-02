// Full app wiring with synthetic MCP requests/capture. Not model-turn acceptance.
const { chromium } = require('../../edition-2/agent/gui/node_modules/playwright');
const assert = require('node:assert/strict');
const base = process.env.GUI_URL || 'http://127.0.0.1:18202/assistant/';
const url = new URL(base);
assert.equal(url.hostname, '127.0.0.1');
assert.ok(!['8088','8089'].includes(url.port), 'Never target live hosts');
(async () => {
  const browser = await chromium.launch({headless:true,channel:'chrome',args:['--autoplay-policy=no-user-gesture-required']});
  const page = await browser.newPage(); page.setDefaultTimeout(15000);
  const errors=[]; page.on('pageerror',e=>errors.push(e.message));
  try {
    await page.addInitScript(() => {
      window.replies=[]; window.captureRequests=0; window.captureTracks=[]; window.captureMode='normal';
      const NativeSocket=WebSocket;
      window.WebSocket=class extends NativeSocket {
        constructor(...args) { super(...args); window.fixtureSocket=this; }
        send(text) {
          const frame=JSON.parse(text);
          if(frame.type==='mcp' && frame.data.id>=9000 && frame.data.id<10000) { replies.push(frame); return; }
          return super.send(text);
        }
      };
      let identity;
      Object.defineProperty(navigator.mediaDevices,'setCaptureHandleConfig',{configurable:true,value:config=>{identity=config;}});
      Object.defineProperty(navigator.mediaDevices,'getDisplayMedia',{configurable:true,value:()=>{
        captureRequests++;
        const canvas=document.createElement('canvas');canvas.width=64;canvas.height=32;
        const context=canvas.getContext('2d');context.fillStyle='red';context.fillRect(0,0,32,32);
        context.fillStyle='blue';context.fillRect(32,0,32,32);
        const stream=canvas.captureStream(), track=stream.getVideoTracks()[0], handle=identity.handle;
        captureTracks.push(track);
        Object.defineProperty(track,'getSettings',{value:()=>({displaySurface:'browser'})});
        Object.defineProperty(track,'getCaptureHandle',{value:()=>({handle,origin:location.origin})});
        return captureMode==='hold' ? new Promise(resolve=>{window.releaseCapture=()=>resolve(stream);}) : Promise.resolve(stream);
      }});
    });
    await page.goto(base);
    await page.waitForFunction(()=>typeof captureConsent!=='undefined' && captureConsent && !selectionPending && fixtureSocket.relayAgentID===agentID);
    const original=await page.evaluate(()=>agentID);
    await page.evaluate(()=>{
      fixtureSocket.mcpInitialized=true;
      window.rpc=(id,name='gui_capture',args={})=>handleMcp(fixtureSocket,{agent_id:fixtureSocket.relayAgentID,data:{jsonrpc:'2.0',id,method:'tools/call',params:{name,arguments:args}}});
      window.notify=(method,params)=>handleMcp(fixtureSocket,{agent_id:fixtureSocket.relayAgentID,data:{jsonrpc:'2.0',method,params}});
    });
    async function reply(id) {
      await page.waitForFunction(id=>replies.some(r=>r.data.id===id),id);
      return page.evaluate(id=>replies.find(r=>r.data.id===id),id);
    }
    await page.locator('#prompt').fill('Preserve this unsent draft');
    await page.evaluate(()=>rpc(9000));
    await page.waitForFunction(()=>document.querySelector('#capture-consent').open);
    assert.equal(await page.locator('[data-capture-requester]').textContent(),await page.locator('#agent-choice option:checked').textContent());
    await page.evaluate(()=>{rpc(9001,'gui_click',{selector:'#capture-share'});rpc(9002,'gui_snapshot');});
    assert.equal((await reply(9001)).data.result.isError,undefined);
    assert.equal((await reply(9002)).data.result.isError,undefined);
    assert.equal(await page.evaluate(()=>captureRequests),0,'Agent click must not approve capture');
    assert.equal(await page.evaluate(()=>document.querySelector('#capture-consent').open),true);
    assert.equal(await page.locator('#send').isDisabled(),true);
    await page.locator('#capture-share').click();
    const accepted=await reply(9000);
    assert.equal(accepted.agent_id,original);
    assert.equal(accepted.data.result.content[0].type,'image');
    assert.equal(accepted.data.result.content[0].mimeType,'image/png');
    const pixels=await page.evaluate(async data=>{
      const blob=await (await fetch('data:image/png;base64,'+data)).blob();
      const image=await createImageBitmap(blob), canvas=document.createElement('canvas');
      canvas.width=image.width;canvas.height=image.height;
      const c=canvas.getContext('2d');c.drawImage(image,0,0);image.close();
      return {w:canvas.width,h:canvas.height,left:[...c.getImageData(8,16,1,1).data],right:[...c.getImageData(56,16,1,1).data]};
    },accepted.data.result.content[0].data);
    assert.equal(pixels.w,64);assert.equal(pixels.h,32);
    assert.ok(pixels.left[0]>=250 && pixels.left[2]<=4 && pixels.right[2]>=250 && pixels.right[0]<=4);
    assert.equal(await page.locator('#prompt').inputValue(),'Preserve this unsent draft');
    assert.equal(await page.evaluate(()=>imageComposer.hasImages),false);
    assert.equal(await page.evaluate(()=>captureTracks.every(t=>t.readyState==='ended')),true);
    await page.locator('#prompt').fill('');
    console.log('PASS full-page trusted consent, ordinary RPC responsiveness, actual pixels and draft preservation');

    await page.evaluate(()=>{rpc(9003);notify('$/cancelRequest',{id:9999});});
    assert.equal(await page.evaluate(()=>document.querySelector('#capture-consent').open),true);
    await page.evaluate(()=>notify('$/cancelRequest',{id:9003}));
    assert.equal((await reply(9003)).data.result.isError,true);
    await page.evaluate(()=>rpc(9004));
    await page.waitForFunction(()=>document.querySelector('#capture-consent').open);
    const other=await page.locator('#agent-choice').evaluate((s,original)=>[...s.options].find(o=>o.value && o.value!==original)?.value,original);
    assert.ok(other,'fixture needs another selectable Agent');
    await page.evaluate(key=>selectAgent(key),other);
    assert.equal((await reply(9004)).data.result.isError,true);
    await page.waitForFunction(key=>agentID===key && !selectionPending,other);
    await page.evaluate(()=>rpc(9005));
    assert.equal((await reply(9005)).data.result.isError,true);
    assert.equal(await page.evaluate(()=>document.querySelector('#capture-consent').open),false);
    await page.evaluate(key=>selectAgent(key),original);
    await page.waitForFunction(key=>agentID===key && !selectionPending,original);
    console.log('PASS exact RPC cancellation and selection invalidation; other displayed Agent refuses');

    await page.evaluate(()=>{captureMode='hold';rpc(9006);});
    await page.waitForFunction(()=>document.querySelector('#capture-consent').open);
    await page.locator('#capture-share').click();
    await page.waitForFunction(()=>typeof releaseCapture==='function');
    await page.evaluate(()=>notify('notifications/ensemble/browser_closed'));
    assert.equal(await page.evaluate(()=>replies.some(r=>r.data.id===9006)),false);
    await page.evaluate(()=>releaseCapture());
    assert.equal((await reply(9006)).data.result.isError,true);
    assert.equal(await page.evaluate(()=>captureTracks.every(t=>t.readyState==='ended')),true);
    await page.evaluate(()=>rpc(9007));
    assert.equal((await reply(9007)).data.error.code,-32000);
    console.log('PASS session close joins late picker and revokes initialized dispatch');

    await page.evaluate(()=>{fixtureSocket.mcpInitialized=true;captureMode='normal';rpc(9008);});
    await page.waitForFunction(()=>document.querySelector('#capture-consent').open);
    await page.keyboard.press('Escape');
    assert.equal((await reply(9008)).data.result.isError,true);
    await page.evaluate(()=>{document.querySelector('#settings-dialog').showModal();rpc(9009);});
    assert.equal((await reply(9009)).data.result.isError,true);
    assert.equal(await page.evaluate(()=>document.querySelector('#settings-dialog').open),true);
    await page.evaluate(()=>document.querySelector('#settings-dialog').close());
    console.log('PASS Escape declines and existing Settings dialog is preserved');

    await page.evaluate(()=>{captureMode='hold';releaseCapture=null;rpc(9010);});
    await page.waitForFunction(()=>document.querySelector('#capture-consent').open);
    await page.locator('#capture-share').click();
    await page.waitForFunction(()=>typeof releaseCapture==='function');
    await page.evaluate(()=>fixtureSocket.close());
    await page.waitForFunction(()=>!document.querySelector('#capture-consent').open);
    await page.evaluate(()=>releaseCapture());
    await page.waitForFunction(()=>!captureConsent.busy && !selectionPending && fixtureSocket.readyState===1 && fixtureSocket.relayAgentID===agentID);
    assert.equal(await page.evaluate(()=>replies.some(r=>r.data.id===9010)),false);
    console.log('PASS disconnect suppresses late pixels; replacement socket receives no replay');

    await page.setViewportSize({width:393,height:852});
    await page.evaluate(()=>{document.documentElement.style.fontSize='32px';fixtureSocket.mcpInitialized=true;rpc(9011);});
    await page.waitForFunction(()=>document.querySelector('#capture-consent').open);
    const geometry=await page.locator('#capture-consent').evaluate(d=>{
      const r=d.getBoundingClientRect();return {left:r.left,right:r.right,width:innerWidth,overflow:d.scrollWidth-d.clientWidth,
        buttons:[...d.querySelectorAll('button')].map(b=>{const q=b.getBoundingClientRect();return {left:q.left,right:q.right,height:q.height};})};
    });
    assert.ok(geometry.left>=0 && geometry.right<=geometry.width+1 && geometry.overflow<=1,JSON.stringify(geometry));
    assert.ok(geometry.buttons.every(b=>b.left>=0 && b.right<=geometry.width+1 && b.height>=44),JSON.stringify(geometry));
    await page.locator('#capture-decline').click();
    assert.equal((await reply(9011)).data.result.isError,true);
    const before=await page.evaluate(()=>captureRequests);
    await page.evaluate(()=>browserCall('gui_click',{selector:'#capture-page'}));
    assert.equal(await page.evaluate(()=>captureRequests),before,'Manual camera also requires human activation');
    assert.deepEqual(errors,[]);
    console.log('PASS phone/enlarged-text dialog geometry and manual-camera trusted activation');
  } finally { await browser.close(); }
})().catch(error=>{console.error(error);process.exitCode=1;});
