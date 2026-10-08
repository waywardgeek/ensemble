// Student-owned local browser controls. No real provider or audible-speech claim.
import {chromium} from '/Users/bill/projects/ensemble-edition-2-revisions/browser-tools/node_modules/playwright/index.mjs';
import {spawn} from 'node:child_process';
import {createServer} from 'node:http';
import {mkdtemp,mkdir,writeFile,readFile,rm} from 'node:fs/promises';
import {tmpdir} from 'node:os';
import {join,dirname} from 'node:path';
import {fileURLToPath} from 'node:url';
import assert from 'node:assert/strict';
const here=dirname(fileURLToPath(import.meta.url)),dir=await mkdtemp(join(tmpdir(),'ch09-student-ui-')),checks=[];
const manual='<img src=x onerror="globalThis.injected=true">\n'+'Read before editing. '.repeat(90)+'END-MANUAL\n';
let requests=0,gui,browser;
const fixture=createServer(async(req,res)=>{for await(const ignored of req){} requests++;const calls={1:{name:'load_skill',input:{name:'edit'}},3:{name:'write_file',input:{path:'note.txt',content:'actual bytes'}},5:{name:'unload_skill',input:{name:'edit'}}};const call=calls[requests];res.setHeader('content-type','application/json');res.end(JSON.stringify({content:call?[{type:'tool_use',id:'c'+requests,...call}]:[{type:'text',text:'Completed.'}],usage:{input_tokens:1,output_tokens:1}}));});
await new Promise(r=>fixture.listen(0,'127.0.0.1',r));
try{
 for(const [name,type,fields,body] of [['base','primary','loadable-skills: edit\n','Identity.'],['edit','loadable','tools: write_file\n',manual]]){await mkdir(join(dir,'skills',name),{recursive:true});await writeFile(join(dir,'skills',name,'SKILL.md'),`---\nname: ${name}\ndescription: Manual ${name}\ntype: ${type}\n${fields}---\n${body}`)}
 gui=spawn(process.argv[2]||'/tmp/ensemble-ch09-student-gui',['--port','0'],{cwd:dir,env:{...process.env,LLM_VENDOR:'anthropic',LLM_MODEL:'fixture',LLM_API_KEY:'fixture',LLM_BASE_URL:`http://127.0.0.1:${fixture.address().port}`,LLM_SKILLS_DIR:join(dir,'skills'),LLM_PRIMARY_SKILL:'base',LLM_SYSTEM:'',EN_DISABLE_STREAMING:'1',CH02_LOG:join(dir,'events')},stdio:['ignore','pipe','pipe']});
 let diagnostics='';gui.stderr.on('data',d=>diagnostics+=d);const url=await new Promise((resolve,reject)=>{gui.stdout.once('data',d=>resolve(d.toString().trim()));gui.once('exit',code=>reject(new Error(`${code}: ${diagnostics}`)))});
 browser=await chromium.launch({headless:true,channel:'chrome'});const page=await browser.newPage();const errors=[];page.on('pageerror',e=>errors.push(e.message));await page.goto(url);await page.locator('[data-status]').filter({hasText:'Connected'}).waitFor();assert.match(await page.locator('[data-skills]').innerText(),/Primary: base/);assert.equal(requests,0);checks.push('initial read-only sidebar');
 for(const [index,prompt] of ['load edit','write note','unload edit'].entries()){
  await page.locator('[data-input]').fill(prompt);await page.locator('[data-prompt]').click();await page.locator('[data-notice]').filter({hasText:`Request r${index+1}: success`}).waitFor();
 }
 assert.equal(await readFile(join(dir,'note.txt'),'utf8'),'actual bytes');assert.equal(requests,6);checks.push('load write unload through real GUI with local fixture');
 const card=page.locator('[data-artifacts] .artifact').filter({hasText:'Skill: edit'});assert.equal(await card.count(),1);assert.match(await card.innerText(),/Retired skill/);await card.getByRole('button',{name:'Expand full retained text'}).press('Enter');assert.match(await card.innerText(),/END-MANUAL/);assert.equal(await page.locator('img').count(),0);assert.equal(await page.evaluate(()=>globalThis.injected),undefined);assert.equal(await page.locator('[data-actions] .artifact').filter({hasText:'Skill: edit'}).count(),0);checks.push('safe retained chat manual and keyboard expansion');
 await page.reload();await page.locator('[data-status]').filter({hasText:'Connected'}).waitFor();assert.equal(await page.locator('[data-artifacts] .artifact').filter({hasText:'Skill: edit'}).count(),1);assert.match(await page.locator('[data-skills]').innerText(),/Revision: 2/);checks.push('reconnect current state and one retired card');
 const component=await page.evaluate(async()=>{
  const {Connector}=await import('/connector.js'),{ArtifactScroll}=await import('/artifacts.js'),{SpeechQueue}=await import('/speech.js');const spoken=[];const owner={speak:(key,text)=>spoken.push({key,text}),diagnostic(){},application(){return {speech(){return {available(){return true}}}}}};
  const connector=new Connector(owner);const raw='{"type":"snapshot_event","event":{"skills":{"state":{"revision":18446744073709551615,"active":[{"activation":9007199254740992}],"retired":[{"activation":9007199254740993}]},"activated":[{"activation":9007199254740992,"name":"one","body":"FIRST","dependencies":[9007199254740993]},{"activation":9007199254740993,"name":"two","body":"SECOND FULL","dependencies":[]}]}}}';
  const decoded=connector.decode(raw);const element=document.createElement('section');document.body.append(element);const scroll=new ArtifactScroll(owner,element,'chat');const event={...decoded.event,type:'skills_changed',seq:1};scroll.event(event,false,'agent-x');const identities=[...scroll.cards.keys()];[...scroll.cards.values()][1].speaker.click();
  const queue=new SpeechQueue(owner);queue.enabled=true;queue.playbackReady=true;queue.observe({kind:'skills_changed',event});const silent=!queue.busy();scroll.reset({agent_id:'agent-x',events:[],partials:[],omitted:101,state:{skills:decoded.event.skills.state}});const windowKeys=[...scroll.cards.keys()];scroll.close();element.remove();
  const arbitrary=connector.decode('{"type":"snapshot_event","event":{"tool":{"args":{"activation":9007199254740993,"revision":9007199254740993}}}}');
  return {revision:String(decoded.event.skills.state.revision),dependency:String(decoded.event.skills.activated[0].dependencies[0]),identities,spoken,silent,windowKeys,ordinary:typeof arbitrary.event.tool.args.activation};
 });
 assert.equal(component.revision,'18446744073709551615');assert.equal(component.dependency,'9007199254740993');assert.deepEqual(component.identities,['skill/agent-x/9007199254740992','skill/agent-x/9007199254740993']);assert.equal(component.spoken[0].text,'SECOND FULL');assert(component.silent);assert.deepEqual(component.windowKeys,['omitted']);assert.equal(component.ordinary,'number');checks.push('isolated uint64 card identities, explicit full speech, silent automatic material, unchanged arguments');
 assert.deepEqual(errors,[]);console.log(JSON.stringify({passed:true,scope:'local fake HTTP and browser component controls',checks,requests},null,2));
}finally{await browser?.close();if(gui&&!gui.killed){const done=new Promise(r=>gui.once('exit',r));gui.kill('SIGTERM');await done}await new Promise(r=>fixture.close(r));await rm(dir,{recursive:true,force:true})}
