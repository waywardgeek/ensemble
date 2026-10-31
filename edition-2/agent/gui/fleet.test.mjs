import assert from "node:assert/strict";
import http from "node:http";
import net from "node:net";
import { spawn } from "node:child_process";
import { mkdtemp, rm, writeFile } from "node:fs/promises";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { chromium } from "playwright";

// Only the provider is a fake: use the built GUI, real tools and recursive actors.
let releaseLeaf;
const leafGate=new Promise(resolve=>{releaseLeaf=resolve;});
const steps=new Map(),requests=[],frames=[];
const provider=http.createServer(async(request,response)=>{
 let body="";for await(const bytes of request)body+=bytes;
 const value=JSON.parse(body);requests.push(value);
 const content=value.messages.filter(m=>m.role==="user" && ["ROOT","ROOT_OTHER","PEER","TREE_TASK","CHILD_TASK","LEAF_TASK"].includes(m.content)).at(-1)?.content;
 const prompt=typeof content==="string"?content:content?.map(part=>part.text||"").join("");
 const step=steps.get(prompt)||0;steps.set(prompt,step+1);
 let tool="submit",args={result:`${prompt}_RESULT`};
 if(prompt==="TREE_TASK"||prompt==="CHILD_TASK"){
  if(step===0){tool="spawn_sub_agent";args={agent_id:"peer",task:prompt==="TREE_TASK"?"CHILD_TASK":"LEAF_TASK"};}
  if(step===1){tool="join_agents";args={agent_ids:["peer"],ai_callback_delay:30};}
 }
 if(prompt==="LEAF_TASK")await leafGate;
 response.setHeader("content-type","application/json");
 response.end(JSON.stringify({choices:[{message:{role:"assistant",content:`${prompt}_REPLY`,tool_calls:[{id:`call-${step}`,type:"function",function:{name:tool,arguments:JSON.stringify(args)}}]},finish_reason:"tool_calls"}],usage:{prompt_tokens:3,completion_tokens:2}}));
});
await new Promise(resolve=>provider.listen(0,"127.0.0.1",resolve));
const probe=net.createServer();await new Promise(resolve=>probe.listen(0,"127.0.0.1",resolve));
const port=probe.address().port;await new Promise(resolve=>probe.close(resolve));
const directory=await mkdtemp(join(tmpdir(),"ch26-browser-"));
const child=spawn(process.env.GUI_BINARY,["--port",String(port),"--no-stream"],{cwd:directory,env:{...process.env,LLM_VENDOR:"openai",LLM_MODEL:"gpt-4o-mini",LLM_API_KEY:"fake",LLM_BASE_URL:`http://127.0.0.1:${provider.address().port}`},stdio:["pipe","pipe","pipe"]});
let stderr="";child.stderr.on("data",bytes=>{stderr+=bytes;});child.stdout.resume();
const exited=new Promise(resolve=>child.on("exit",(code,signal)=>resolve({code,signal})));
const browser=await chromium.launch({headless:true,executablePath:process.env.CHROME_BIN||"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"});
try{
 for(let i=0;i<100;i++){try{if((await fetch(`http://127.0.0.1:${port}`)).ok)break;}catch{}await new Promise(resolve=>setTimeout(resolve,25));}
 const page=await browser.newPage(),errors=[];
 page.on("pageerror",error=>errors.push(error.message));
 page.on("websocket",socket=>socket.on("framereceived",frame=>frames.push(JSON.parse(frame.payload))));
 await page.goto(`http://127.0.0.1:${port}/?speech=transcript`);
 await page.locator('[data-agent-key="main"]').waitFor();
 async function prompt(text){await page.locator("#input").fill(text);await page.locator("#input-form button").click();}
 async function complete(text){await page.waitForFunction(text=>document.querySelector("#actions").textContent.includes(text),text);}
 await prompt("ROOT");await complete("ROOT_RESULT");
 await page.locator("#agent-name").fill("peer");await page.locator("#agent-model").fill("gpt-4o-mini");await page.locator("#create-agent button").click();
 await page.locator('[data-agent-key="peer"]').click();await prompt("PEER");await complete("PEER_RESULT");
 assert.ok(!(await page.locator("#artifacts").textContent()).includes("ROOT_REPLY"));
 const second=await browser.newPage();await second.goto(`http://127.0.0.1:${port}/?speech=transcript`);
 await second.locator('[data-agent-key="main"]').waitFor();
 await second.locator("#input").fill("ROOT_OTHER");await second.locator("#input-form button").click();
 await second.waitForFunction(()=>document.querySelector("#actions").textContent.includes("ROOT_OTHER_RESULT"));
 assert.ok(!(await page.locator("#artifacts").textContent()).includes("ROOT_OTHER_REPLY"),"another tab's root output leaked into peer conversation");
 await second.close();

 await page.locator('[data-agent-key="main"]').click();await page.waitForFunction(()=>document.querySelector("#artifacts").textContent.includes("ROOT_REPLY"));
 assert.ok(!(await page.locator("#artifacts").textContent()).includes("PEER_REPLY"));
 await page.locator('[data-agent-key="peer"]').click();await page.waitForFunction(()=>document.querySelector("#artifacts").textContent.includes("PEER_REPLY"));
 await prompt("TREE_TASK");await page.locator('[data-agent-key="peer/peer/peer"]').waitFor();
 await page.getByRole("button",{name:"Delete peer",exact:true}).click();await page.waitForFunction(()=>document.querySelector("#agent-status").textContent.includes("active children"));
 releaseLeaf();await complete("TREE_TASK_RESULT");
 await page.locator('[data-agent-key="peer/peer/peer"]').click();await page.waitForFunction(()=>document.querySelector("#artifacts").textContent.includes("LEAF_TASK_REPLY"));
 assert.ok(!(await page.locator("#artifacts").textContent()).includes("PEER_REPLY"));
 await page.getByRole("button",{name:"Delete peer",exact:true}).click();await page.waitForFunction(()=>!document.querySelector('[data-agent-key="peer"]'));
 await page.waitForFunction(()=>document.querySelector("#conversation-title").textContent==="main");
 assert.ok(frames.some(frame=>frame.type==="agent_activity"&&frame.activity.phase==="waiting"));
 assert.ok(frames.some(frame=>frame.type==="agent_state"));
 assert.ok(requests.every(request=>request.model==="gpt-4o-mini"));assert.deepEqual(errors,[]);
 console.log(JSON.stringify({browser:"pass",requests:requests.length,treeFrames:frames.filter(f=>f.type==="agent_tree").length,selected:frames.filter(f=>f.type==="agent_selected").map(f=>f.key),phases:[...new Set(frames.filter(f=>f.type==="agent_activity").map(f=>f.activity.phase))]}));
}catch(error){console.error(JSON.stringify({error:error.message,requests:requests.length,frames:frames.slice(-4).map(f=>({type:f.type,key:f.key,error:f.error})),stderr:stderr.slice(-1000)}));throw error;}finally{
 if(process.env.EVIDENCE_DIR)await writeFile(join(process.env.EVIDENCE_DIR,"browser-frames.json"),JSON.stringify({requests,frames},null,2));
 releaseLeaf();await browser.close();child.stdin.end();const timer=setTimeout(()=>child.kill("SIGTERM"),3000);await exited;clearTimeout(timer);
 await new Promise(resolve=>provider.close(resolve));await rm(directory,{recursive:true,force:true});if(child.exitCode!==0)console.error(stderr);
}
