'use strict';
const assert=require('node:assert/strict'),fixture=require('./client_round1_test.cjs');
const tests={
 'unsent target survives production remote frame and disconnect'(){
  const f=fixture();f.run('applySettings(settings)');const input=f.$('context-target');
  input.value='75000';input.oninput?.();
  f.run('socket.onmessage({data:JSON.stringify({type:"settings_changed",settings:{...settings,theme:"light",context_target:90000}})})');
  assert.equal(input.value,'75000');assert.equal(f.run('settings.context_target'),90000);
  f.run('socket.readyState=3;socket.onclose()');assert.equal(input.value,'75000');
  input.onchange();assert.equal(input.value,'75000');assert.equal(f.run('settings.context_target'),90000);
 },
 'submitted target stays visible and accepted ack does not erase newer typing'(){
  const f=fixture();f.run('applySettings(settings)');const input=f.$('context-target');input.value='1';input.oninput?.();input.onchange();
  assert.equal(input.value,'1');assert.equal(f.run('settings.context_target'),0);
  assert.equal(f.sent.at(-1).settings.context_target,1);
  f.run('socket.onmessage({data:JSON.stringify({type:"settings_changed",settings:{...settings,context_target:20000}})})');assert.equal(input.value,'20000');
  input.value='80000';input.oninput?.();input.onchange();input.value='85000';input.oninput?.();
  f.run('socket.onmessage({data:JSON.stringify({type:"settings_changed",settings:{...settings,context_target:80000}})})');
  assert.equal(input.value,'85000');assert.equal(f.run('settings.context_target'),80000);
 },
 'invalid and refused edits stay editable without changing accepted settings'(){
  const f=fixture();f.run('applySettings(settings)');const input=f.$('context-target');input.value='1.5';input.oninput?.();input.onchange();
  assert.equal(input.value,'1.5');assert.equal(f.run('settings.context_target'),0);
  input.value='-1';input.oninput?.();input.onchange();
  f.run('socket.onmessage({data:JSON.stringify({type:"settings_error",data:"invalid target"})})');
  assert.equal(input.value,'-1');assert.equal(f.run('settings.context_target'),0);
 }
};
tests['negative submission survives unrelated floor broadcast before refusal']=()=>{
 const f=fixture(),input=f.$('context-target');
 f.run('applySettings({...settings,context_target:20000})');
 input.value='-1';input.oninput();input.onchange();
 assert.equal(f.sent.at(-1).settings.context_target,-1);
 f.run('socket.onmessage({data:JSON.stringify({type:"settings_changed",settings:{...settings,theme:"light",context_target:20000}})})');
 assert.equal(input.value,'-1');assert.equal(f.run('settings.context_target'),20000);
 f.run('socket.onmessage({data:JSON.stringify({type:"settings_error",data:"invalid target"})})');
 assert.equal(input.value,'-1');assert.equal(f.run('settings.context_target'),20000);
};
if(process.argv[2])tests['real settings WS frames preserve drafts through production handlers']=()=>{
 const frames=JSON.parse(require('node:fs').readFileSync(process.argv[2],'utf8')),f=fixture(),input=f.$('context-target');
 const receive=frame=>f.run('socket.onmessage({data:'+JSON.stringify(JSON.stringify(frame))+'})');
 receive(frames.snapshot);input.value='75000';input.oninput();receive(frames.unrelated);
 assert.equal(input.value,'75000');assert.equal(f.run('settings.theme'),'light');assert.match(f.$('settings-status').textContent,/not saved/);
 input.value='1';input.oninput();input.onchange();assert.equal(input.value,'1');receive(frames.accepted);assert.equal(input.value,'20000');assert.equal(f.run('settings.context_target'),20000);
 input.value='-1';input.oninput();input.onchange();assert.equal(f.sent.at(-1).settings.context_target,-1);
 receive(frames.unrelatedAtFloor);assert.equal(input.value,'-1');assert.equal(f.run('settings.context_target'),20000);
 receive(frames.refused);assert.equal(input.value,'-1');assert.equal(f.run('settings.context_target'),20000);
 f.run('socket.readyState=3;socket.onclose()');receive(frames.accepted);assert.equal(input.value,'-1');
};
let failed=0;for(const [name,test]of Object.entries(tests)){try{test();console.log('PASS '+name);}catch(e){failed++;console.error('FAIL '+name+': '+e.stack);}}process.exitCode=failed?1:0;
