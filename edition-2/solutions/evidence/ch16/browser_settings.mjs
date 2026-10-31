// One bounded real-browser check of the shipped memory controls and saved values.
import { createRequire } from 'node:module';
import { mkdtemp, mkdir, readFile, writeFile, rm } from 'node:fs/promises';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { spawn } from 'node:child_process';
import { createServer } from 'node:net';
import { once } from 'node:events';
const root=process.cwd();
const {chromium}=createRequire(join(root,'solutions/edition-2/main/gui/package.json'))('playwright');
const dir=await mkdtemp(join(tmpdir(),'ch16-browser-'));
await mkdir(join(dir,'cr'));
await writeFile(join(dir,'cr/settings.json'),JSON.stringify({tts_enabled:false}));
const listener=createServer();listener.listen(0,'127.0.0.1');await once(listener,'listening');const port=listener.address().port;await new Promise(resolve=>listener.close(resolve));
const child=spawn('/tmp/ensemble-ch16-gui',['--port',String(port),'--gui-dir',process.env.CH16_GUI_DIR || join(root,'solutions/edition-2/main/gui/web/gui')],{cwd:dir,env:{PATH:process.env.PATH,HOME:process.env.HOME,LLM_VENDOR:'anthropic',LLM_MODEL:'claude-ch16-course',LLM_API_KEY:'boundary-unused'},stdio:['pipe','pipe','pipe']});
let diagnostic='';child.stderr.on('data',b=>diagnostic+=b);child.stdout.resume();
let browser;
try {
 browser=await chromium.launch({headless:true,executablePath:'/Applications/Google Chrome.app/Contents/MacOS/Google Chrome'});const page=await browser.newPage({viewport:{width:1280,height:800}});
 for(let n=0;;n++){try{await page.goto(`http://127.0.0.1:${port}`);break}catch(e){if(n===10)throw e;await new Promise(r=>setTimeout(r,100))}}
 await page.getByRole('tab',{name:'Context Management'}).click();
 const budget=page.getByRole('spinbutton',{name:'session budget bytes (0 = default)',exact:true});
 await budget.fill('77');await budget.press('Tab');
 await page.getByLabel('session enabled',{exact:true}).uncheck();
 await page.waitForFunction(()=>!document.querySelector('#state').textContent.includes('connecting'));
 let saved;
 for(let n=0;n<30;n++){
  saved=JSON.parse(await readFile(join(dir,'cr/settings.json'),'utf8'));
  if(saved.memory?.session?.budget===77&&saved.memory?.session?.disabled===true)break;
  await new Promise(r=>setTimeout(r,100));
 }
 if(saved.memory.session.budget!==77||!saved.memory.session.disabled)throw Error('browser changes not persisted');
 await page.reload();await page.getByRole('tab',{name:'Context Management'}).click();
 await page.waitForFunction(()=>document.querySelector('[aria-label="session budget bytes (0 = default)"]').value==='77');
 if(await page.getByLabel('session enabled',{exact:true}).isChecked())throw Error('disabled switch not restored');
 await writeFile(join(root,'solutions/edition-2/evidence/ch16/browser-settings.json'),JSON.stringify({checks:['real shipped GUI loaded','session budget77 persisted','disabled true persisted','reload restored77 and disabled'],saved_session:saved.memory.session},null,2));
 console.log('PASS actual browser memory settings and reload');
} finally {
 if(browser)await browser.close();child.stdin.end();
 const timer=setTimeout(()=>child.kill('SIGKILL'),5000);await once(child,'exit');clearTimeout(timer);
 await rm(dir,{recursive:true,force:true});
}
