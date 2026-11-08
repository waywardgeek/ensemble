'use strict';
// Human-authorized book captures: real page/renderers, explicitly synthetic facts.
// GUI_FIXTURE_URL must name the disposable ArtifactRendererFixture host.
// Args: image output directory, copied distribution skills/shared directory.
const assert = require('node:assert/strict');
const fs = require('node:fs/promises');
const path = require('node:path');
const { chromium } = require('../../edition-2/agent/gui/node_modules/playwright');
const base = process.env.GUI_FIXTURE_URL;
assert.equal(new URL(base).hostname, '127.0.0.1');
assert(!['8088', '8089', '8090'].includes(new URL(base).port));
const output = process.argv[2], shared = process.argv[3];
assert(output && shared);
(async () => {
 await fs.mkdir(output, {recursive:true});
 const browser = await chromium.launch({executablePath:'/Applications/Google Chrome.app/Contents/MacOS/Google Chrome',headless:true});
 const errors=[];
 try {
  const page = await browser.newPage({viewport:{width:690,height:1400},deviceScaleFactor:2,colorScheme:'light'});
  page.on('pageerror',e=>errors.push(e.message));
  await page.route('**/*',route => new URL(route.request().url()).origin===new URL(base).origin ? route.continue() : route.abort());
  await page.route('**/jobs/window?*',route=>{
   const summary=Buffer.from('\x1b[1mBuild summary\x1b[0m\n  Compile: completed\n  Tests: 24 passed, 0 failed\n\nEarlier progress is retained in the Job output file.\n');
   const prefix=Buffer.from("Compiling sample module...\n".repeat(3000)).subarray(0,65536-summary.length);
   prefix[prefix.length-1]=10;
   const bytes=Buffer.concat([prefix,summary]);
   assert.equal(bytes.length,65536);
   return route.fulfill({json:{handle:77,status:'done',start:81920,end:81920+bytes.length,rolled_out:true,encoding:'base64',data:bytes.toString('base64'),origin_call_id:'figure-command-origin',origin_tool:'run_command'}});
  });
  await page.goto(base);
  await page.waitForFunction(()=>settings!==null&&!selectionPending&&jobSessions!==null&&['list_directory','search_files','edit_file','command'].every(k=>typeof window.EnsembleArtifactRenderers?.[k]==='function'));
  await page.locator('.pane-tab[data-pane="tools"]').click();
  await page.evaluate(()=>{
   settings={...settings,tts_enabled:false};
   const add=(id,name,result,extra={})=>{
    handle({type:'tool_dispatched',part_id:id+'-call',name,input:{}});
    handle({type:'tool_finished',part_id:id,tool_part_id:id+'-call',result,...extra});
   };
   add('figure-directory','list_directory','README.md\nsrc/\ntests/\nnotes <draft>.txt',{directory_listing:{total_entries:4,entries:[
    {name:'README.md',kind:'unknown'},{name:'src',kind:'directory'},{name:'tests',kind:'directory'},{name:'notes <draft>.txt',kind:'unknown'}]}});
   add('figure-search','search_files','Synthetic search observation',{search_evidence:{total_lines:6,lines:[
    {path:'src/Report.cs',number:18,text:'public string Summary()',matches:[]},
    {path:'src/Report.cs',number:19,text:'    return Format(total);',matches:[{start:11,length:6}]},
    {path:'tests/ReportTests.cs',number:42,text:'Check(Format(total));',matches:[{start:6,length:6}]}]}});
   const segment=text=>({text,total_bytes:new TextEncoder().encode(text).length});
   add('figure-edit','edit_file','Applied the exact replacement at line 19.',{file_evidence:{operation:'edit',path:'src/Report.cs',start_line:19,start_column:1,removed:segment('    return total.ToString();\n'),content:segment('    return Format(total);\n')}});
   const job={handle:77,status:'done',origin_call_id:'figure-command-origin',origin_tool:'run_command'};
   add('figure-command','run_command','Synthetic command observation: build completed.',{job});
   add('figure-wait','wait_for_job','Synthetic final observation: 24 passed, 0 failed.',{job});
  });
  async function capture(locator,name,expected){
   await locator.scrollIntoViewIfNeeded();
   assert.match(await locator.textContent(),expected);
   await page.evaluate(()=>document.fonts.ready);
   assert(await locator.evaluate(node=>{const p=node.closest('.preview');return !p||p.scrollHeight<=p.clientHeight+1;}),'Figure exceeds preview height: '+name);
   await locator.screenshot({path:path.join(output,name+'.png'),animations:'disabled'});
   console.log('CAPTURE '+name);
  }
  await capture(page.locator('[data-part-id="figure-directory"] .directory-preview'),'directory',/notes <draft>/);
  await capture(page.locator('[data-part-id="figure-search"] .search-preview'),'search',/3 omitted/);
  await capture(page.locator('[data-part-id="figure-edit"] .file-preview'),'replacement',/replacement diff/);
  await page.waitForFunction(()=>document.querySelector('.terminal-output')?.textContent.includes('24 passed'));
  assert.equal(await page.locator('.job-session').count(),1);
  await capture(page.locator('.job-session'),'command',/Earlier output rolled out/);
  await page.setViewportSize({width:1100,height:1900});
  await page.locator('#settings-open').click();
  await page.locator('#settings-display-tab').click();
  const pkg={entry:'main.js',files:{'main.js':'export default function render(document, text) {\n  const output = document.createElement("pre");\n  output.textContent = text;\n  return output;\n}\n'}};
  await page.locator('#renderer-package').setInputFiles({name:'example-renderer.json',mimeType:'application/json',buffer:Buffer.from(JSON.stringify(pkg))});
  await page.waitForFunction(()=>document.querySelector('#renderer-status').textContent.includes('Staging has not executed'));
  await page.locator('#renderer-source').evaluate(node=>node.parentElement.open=false);
  await capture(page.locator('section[aria-labelledby="renderer-heading"]'),'approval',/Staging has not executed/);
  await page.close();
  const dice=await browser.newPage({viewport:{width:880,height:420},deviceScaleFactor:2,colorScheme:'light'});
  dice.on('pageerror',e=>errors.push(e.message));
  await dice.route('**/*',async route=>{
   const url=new URL(route.request().url());
   if(url.origin!=='http://127.0.0.1:18765'||!['/inline-dice.html','/inline-dice.js'].includes(url.pathname))return route.abort();
   await route.fulfill({body:await fs.readFile(path.join(shared,url.pathname.slice(1))),contentType:url.pathname.endsWith('.js')?'text/javascript':'text/html'});
  });
  await dice.goto('http://127.0.0.1:18765/inline-dice.html');
  await dice.locator('summary').click();
  assert.match(await dice.locator('details').textContent(),/17.*3.*3.*23/);
  await dice.locator('body').screenshot({path:path.join(output,'inline-dice.png'),animations:'disabled'});
  console.log('CAPTURE inline-dice');
  await dice.close();
  assert.deepEqual(errors,[]);
 } finally {await browser.close();}
})().catch(error=>{console.error(error);process.exitCode=1;});
