'use strict';
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const { chromium } = require('../../edition-2/agent/gui/node_modules/playwright');
const root = path.resolve(__dirname, '../gui/wwwroot');
const source = fs.readFileSync(path.join(root, 'app.js'), 'utf8');
const render = source.slice(source.indexOf('function render(card) {'), source.indexOf('// State is acknowledged,', source.indexOf('function render(card) {')));
(async () => {
 const browser = await chromium.launch({ executablePath: '/Applications/Google Chrome.app/Contents/MacOS/Google Chrome', headless: true });
 try {
  const page = await browser.newPage();
  await page.setContent('<main id="transcript"></main><aside id="tools"></aside>');
  await page.addStyleTag({path:path.join(root,'vendor/bootstrap-5.3.8.min.css')});
  await page.addStyleTag({path:path.join(root,'style.css')});
  await page.addScriptTag({content:`const transcript=document.querySelector('#transcript'), toolTranscript=document.querySelector('#tools');
   const settings={human_name:'Bill'}, agentRows=[{key:'agent',id:'Cadence'}], speech=true, transcriptSpeech=false;
   function actionButton(label,handler){const b=document.createElement('button');b.className='btn icon-button';b.onclick=handler;return b;}
   function setButtonIcon(b){b.textContent='▶';} function speak(card){window.spoken=card.text;}
   function markdown(body,text){const p=document.createElement('p');p.textContent=text;body.append(p);}
   function fullOutputLinks(){} ${render}
   window.draw=(role)=>{transcript.replaceChildren();window.card={node:document.createElement('article'),role,kind:'text',agent:'agent',text:'Short message',status:''};render(card);};`});
  for(const width of [320,1200]) for(const size of [18,28]) for(const role of ['human','assistant']) {
   await page.setViewportSize({width,height:800});
   await page.evaluate(({size,role})=>{document.body.style.fontSize=size+'px';window.draw(role);},{size,role});
   const result=await page.locator('article').evaluate(a=>{const b=a.querySelector('[data-action="speak"]'),h=a.querySelector('.card-heading'),m=a.querySelector('.message');const br=b.getBoundingClientRect(),hr=h.getBoundingClientRect(),mr=m.getBoundingClientRect();return {parent:b.parentElement.className,w:br.width,h:br.height,headingHeight:hr.height,cardHeight:a.getBoundingClientRect().height,headingBottom:hr.bottom,messageRight:mr.right,buttonLeft:br.left,messageTop:mr.top,bodyHeight:mr.height,label:h.querySelector('strong').textContent,overflow:a.scrollWidth>a.clientWidth};});
   assert.ok(result.headingHeight<44, 'name row is not button-height');assert.ok(result.cardHeight<85, 'whole short card stays compact');assert.ok(result.messageRight<=result.buttonLeft, 'button target cannot cover prose');assert.equal(result.parent,'card-heading');assert.ok(result.w>=44&&result.h>=44);assert.ok(result.messageTop>=result.headingBottom);assert.ok(result.bodyHeight<44);assert.equal(result.label,role==='human'?'Bill':'Cadence');assert.equal(result.overflow,false);
   await page.locator('[data-action="speak"]').click();assert.equal(await page.evaluate(()=>window.spoken),'Short message');
   const separated = await page.evaluate(() => {
    transcript.replaceChildren();
    for (let i=0;i<3;i++) render({node:document.createElement('article'),role:'assistant',kind:'thinking',agent:'agent',text:'',status:''});
    const boxes=[...transcript.querySelectorAll('[data-action="speak"]')].map(b=>b.getBoundingClientRect());
    return boxes.every((b,i)=>i===0 || b.top>=boxes[i-1].bottom);
   });
   assert.ok(separated, 'adjacent empty cards have disjoint playback targets');
   console.log(`PASS ${width}px ${size}px ${role}: compact card, safe targets, playback and empty neighbours`);
  }
 } finally {await browser.close();}
})().catch(e=>{console.error(e);process.exitCode=1;});
