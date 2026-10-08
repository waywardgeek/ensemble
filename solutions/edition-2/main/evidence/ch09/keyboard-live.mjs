// Separately bound zero-model-request keyboard action on an existing live GUI.
import {chromium} from '/Users/bill/projects/ensemble-edition-2-revisions/browser-tools/node_modules/playwright/index.mjs';
import {readFileSync,writeFileSync,mkdirSync} from 'node:fs';
import {spawnSync} from 'node:child_process';
import {createHash} from 'node:crypto';
import {dirname,resolve,join} from 'node:path';
import {fileURLToPath} from 'node:url';
const self=fileURLToPath(import.meta.url),here=dirname(self);
const [bindingPath,runPath,url,outPath,sourceRevision]=process.argv.slice(2);
const run=resolve(runPath),out=resolve(outPath),binding=JSON.parse(readFileSync(bindingPath));
const check=spawnSync('python3',['-c',`import json,sys;from pathlib import Path;from evidence import preflight,check_launch; b=json.loads(Path(sys.argv[1]).read_text());preflight(b);check_launch(b,json.loads((Path(sys.argv[2])/'launch.json').read_text()))`,resolve(bindingPath),run],{cwd:here,encoding:'utf8'});
if(check.status!==0)throw Error(check.stderr||check.stdout);
if(!/^[0-9a-f]{40}$/.test(sourceRevision))throw Error('immutable helper revision required');
const original=spawnSync('git',['show',sourceRevision+':solutions/edition-2/main/evidence/ch09/keyboard-live.mjs'],{cwd:here});
if(original.status!==0||!original.stdout.equals(readFileSync(self)))throw Error('helper source identity mismatch');
const sha=b=>createHash('sha256').update(b).digest('hex');
const browser=await chromium.launch({executablePath:binding.executables.chrome.path,headless:false});
const events=[];let result;
try{
 const page=await browser.newPage();page.on('websocket',socket=>socket.on('framesent',event=>events.push(event.payload.toString())));
 await page.goto(url);await page.locator('[data-status]').filter({hasText:'Connected'}).waitFor();
 const card=page.locator('.artifact').filter({hasText:'Skill: edit'}).first();
 const button=card.getByRole('button',{name:'Expand full retained text'});
 await button.waitFor();await button.focus();
 const before=await card.innerText();await button.press('Enter');
 const after=await card.innerText();
 if(!before.includes('omitted from preview')||after.includes('omitted from preview')||!after.includes('Review step 20:')||!after.includes('<example>'))throw Error('keyboard expansion did not expose exact retained text');
 if(events.some(x=>{try{return ['prompt','hint'].includes(JSON.parse(x).type)}catch{return false}}))throw Error('unexpected model action');
 mkdirSync(out);await page.screenshot({path:join(out,'keyboard.png'),fullPage:true});
 result={source_revision:binding.source_revision,helper_revision:sourceRevision,helper_sha256:sha(readFileSync(self)),binding_sha256:sha(readFileSync(bindingPath)),launch_sha256:sha(readFileSync(join(run,'launch.json'))),url,at:new Date().toISOString(),actor:'Codex student coder, not Bill',action:'focus actual edit-card button, press Enter',before,after,frames_sent:events,model_prompts:0,browser:browser.version(),node:binding.executables.node,chrome:binding.executables.chrome,screenshot_sha256:sha(readFileSync(join(out,'keyboard.png')))};
}finally{await browser.close()}
writeFileSync(join(out,'keyboard.json'),JSON.stringify({...result,browser_closed:true},null,2)+'\n');
console.log(JSON.stringify({keyboard_expanded:true,model_prompts:0,output:out}));
