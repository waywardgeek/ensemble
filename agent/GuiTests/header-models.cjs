// Static DOM checks only: no Agent, socket, credential, or model request.
const {chromium}=require(process.env.PLAYWRIGHT_MODULE || 'playwright');
const fs=require('node:fs'), http=require('node:http'), path=require('node:path'), assert=require('node:assert/strict'), os=require('node:os');
const root=path.resolve(__dirname,'../gui/wwwroot');
const artifacts=process.env.GUI_ARTIFACT_DIR || fs.mkdtempSync(path.join(require('node:os').tmpdir(),'ensemble-header-'));
fs.mkdirSync(artifacts,{recursive:true});
const old=fs.readFileSync(path.join(root,'style.css'),'utf8');
const current=fs.readFileSync(path.join(root,'style.css'),'utf8');
(async()=>{
 const server=http.createServer((req,res)=>{const name=decodeURIComponent(new URL(req.url,'http://localhost').pathname);const p=path.join(root,name==='/'?'index.html':name);if(!p.startsWith(root+'/')){res.writeHead(403).end();return;}try{const data=fs.readFileSync(p);res.setHeader('Content-Type',p.endsWith('.css')?'text/css':p.endsWith('.svg')?'image/svg+xml':p.endsWith('.html')?'text/html':p.endsWith('.png')?'image/png':'application/octet-stream');res.end(p.endsWith('/style.css')?old:data);}catch{res.writeHead(404).end();}});
 await new Promise(r=>server.listen(0,'127.0.0.1',r));let browser;
 try{
 browser=await chromium.launch({executablePath:process.env.CHROME_PATH || '/Applications/Google Chrome.app/Contents/MacOS/Google Chrome',headless:true});const page=await browser.newPage();
 await page.route('**/*.js',r=>r.abort()); // Static layout only: no app connection or model requests.

 const renderSource=fs.readFileSync(path.join(root,'app.js'),'utf8');
 const renderer=renderSource.slice(renderSource.indexOf('function showUsage(frame)'),renderSource.indexOf('function showAgents()'));
 const goalRenderer=renderSource.slice(renderSource.indexOf('function showGoal('),renderSource.indexOf('function updateGoal('));
 const rows=[['OpenAI','gpt-6-astra','GPT-6 Astra'],['OpenAI','gpt-5.6-sol','GPT-5.6 Sol'],['Anthropic','claude-opus-5','Claude Opus 5'],['Anthropic','claude-sonnet-5','Claude Sonnet 5'],['Anthropic','claude-opus-4-6','Claude Opus 4.6'],['Gemini','gemini-3.8-flash','Gemini 3.8 Flash'],['Gemini','gemini-4-argon','Gemini 4 Argon']];
 const results=[];
 for(const theme of ['dark','light'])for(const width of [1600,1024,640])for(const font of [18,28]){
 await page.setViewportSize({width,height:900});await page.goto(`http://127.0.0.1:${server.address().port}/`);
 await page.evaluate(({theme,font})=>{document.documentElement.dataset.bsTheme=theme;document.documentElement.style.fontSize=font+'px';document.querySelector('#peer-open').disabled=false;},{theme,font});
 await page.addScriptTag({content:'function renderControls() {}\n'+renderer});
 await page.addScriptTag({content:'let socket={readyState:1}; let goalSelectionPending=false; let goalDirty=false;\n'+goalRenderer});
 const goals=[
  [null,'Goal …'],
  [{file:null,available:true,enabled:false,phase:'unconfigured'},'Don’t Be Evil'],
  [{file:'GOAL.md',available:true,enabled:true,idle_seconds:480,phase:'busy'},'Goal on · reminders 8 min'],
  [{file:'GOAL.md',available:true,enabled:false,idle_seconds:480,phase:'paused'},'Goal on · reminders off'],
  [{file:'GOAL.md',available:true,enabled:true,idle_seconds:90,phase:'scheduled'},'Goal on · reminders 90 s'],
  [{file:'GOAL.md',available:true,enabled:false,last_error:'fixture error',phase:'file_error'},'Goal file error · reminders off']
 ];
 for(const [goal,label] of goals){
  await page.evaluate(goal=>showGoal(goal),goal);
  assert.equal(await page.locator('#header-goal').innerText(),label);
  const bounds=await page.evaluate(()=>{const g=document.querySelector('#header-goal').getBoundingClientRect(),p=document.querySelector('.header-start').getBoundingClientRect();return {right:g.right,parent:p.right};});
  assert(bounds.right<=bounds.parent+1,'Goal display exceeds its header slot');
 }
 await page.evaluate(()=>{socket.readyState=3;showGoal({file:'GOAL.md',enabled:true});});
 assert.equal(await page.locator('#header-goal').innerText(),'Goal …');
 await page.evaluate(()=>{socket.readyState=1;goalSelectionPending=true;showGoal({file:null});});
 assert.equal(await page.locator('#header-goal').innerText(),'Goal …');
 await page.evaluate(()=>{goalSelectionPending=false;showGoal({file:null,available:true,phase:'unconfigured'});});

 for(const vendor of ['OpenAI','Anthropic','Gemini']){
 const frame={selected_model:rows.find(row=>row[0]===vendor)[1],model_vendor:vendor,model_choices:rows.filter(row=>row[0]===vendor).map(row=>row[1]),recommended_models:rows.map(([v,id,label])=>({vendor:v,id,label,enabled:v===vendor})),usage:{totals:{}},total_cost_usd:null,hit_rate:null};
 await page.evaluate(frame=>showUsage(frame),frame);
 const menu=await page.locator('#model-choice').evaluate(e=>({selected:e.value,groups:[...e.children].map(g=>g.label),options:[...e.options].map(o=>({id:o.value,label:o.textContent,disabled:o.disabled}))}));
 assert.deepEqual(menu.groups,['OpenAI','Anthropic','Google']);assert.equal(menu.selected,frame.selected_model);assert.equal(menu.options.length,7);
 for(const o of menu.options)assert.equal(o.disabled,!frame.model_choices.includes(o.id));
 assert(!menu.options.some(o=>/5\.5|carbon/i.test(o.id)));
 }
 const legacy={selected_model:'offline',model_vendor:'Anthropic',model_choices:['offline'],recommended_models:rows.map(([vendor,id,label])=>({vendor,id,label,enabled:vendor==='Anthropic'})),usage:{totals:{}}};
 await page.evaluate(frame=>showUsage(frame),legacy);
 assert.equal(await page.locator('#model-choice').inputValue(),'offline');
 assert.equal(await page.locator('#model-choice option').count(),8);
 assert.equal(await page.locator('#model-choice option[value="offline"]').textContent(),'offline (current)');
 const geometry=await page.evaluate(()=>{const rect=s=>{const r=document.querySelector(s).getBoundingClientRect();return {width:r.width,height:r.height}};const p=document.querySelector('#peer-open');return {header:rect('#status-bar'),logo:rect('.brand-mark'),model:rect('#model-choice'),interrupt:rect('#interrupt'),usage:parseFloat(getComputedStyle(document.querySelector('#usage-meter')).fontSize),createBackground:getComputedStyle(p).backgroundColor,createText:getComputedStyle(p).color};});
 assert.equal(geometry.logo.width,36);assert.equal(geometry.interrupt.height,36);assert.equal(geometry.createBackground,'rgba(0, 0, 0, 0)');assert(Math.abs(geometry.usage-font*.85)<.1);
 if(font===18){assert(geometry.header.height<=42);assert(geometry.model.height<=36.1);}
 assert(geometry.model.height>=font*1.15);
 await page.locator('#peer-open').hover();assert.notEqual(await page.locator('#peer-open').evaluate(e=>getComputedStyle(e).backgroundColor),'rgba(0, 0, 0, 0)');
 await page.locator('#peer-open').focus();assert(await page.locator('#peer-open').evaluate(e=>e===document.activeElement));
 // Decode the SVG at actual favicon sizes; this checks rendering, not human recognition.
 const icon=await page.evaluate(async()=>{const link=document.querySelector('link[rel="icon"]');const img=new Image();img.src=link.href;await img.decode();return [16,32].map(size=>{const c=document.createElement('canvas');c.width=c.height=size;const x=c.getContext('2d');x.drawImage(img,0,0,size,size);const d=x.getImageData(0,0,size,size).data;let count=0;for(let i=3;i<d.length;i+=4)if(d[i]>100)count++;return {size,count,href:link.href};});});
 for(const i of icon){assert(i.href.endsWith('music-note.svg'));assert(i.count>i.size*i.size*.15);assert(i.count<i.size*i.size*.85);}
 results.push({theme,width,font,...geometry,icon});
 }
 assert.notEqual(results.find(x=>x.theme==='dark').createText,results.find(x=>x.theme==='light').createText);
 fs.writeFileSync(path.join(artifacts,'geometry.json'),JSON.stringify(results,null,2));console.log('PASS 12 layouts: compact header, unchanged logo/usage font, themed Create Agent, vendor groups, provider restrictions, SVG decoding at 16/32px.');console.log(JSON.stringify(results[0]));
 }finally{if(browser)await browser.close();await new Promise(r=>server.close(r));}
})().catch(e=>{console.error(e);process.exitCode=1});
