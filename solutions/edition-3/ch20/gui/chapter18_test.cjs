'use strict';
const assert=require('node:assert/strict'),fs=require('node:fs');
const f=require('./client_round1_test.cjs')();
const frames=JSON.parse(fs.readFileSync(process.argv[2],'utf8'));
let selected=false,afterState=false;
function labels(){
 assert.equal(f.$('model').textContent,'new-private-model');
 assert.equal(f.$('agent-root').title,'new-private-model');
 assert.equal(f.$('model-name').value,'new-private-model');
 assert.match(f.$('cache-meter').title,/Selected model: new-private-model; last producing model: claude-sonnet-4-6/);
}
for(const frame of frames){
 f.run('socket.onmessage({data:'+JSON.stringify(JSON.stringify(frame))+'})');
 if(frame.type==='usage'&&frame.meter.selected_model==='new-private-model')selected=true;
 if(selected){labels();f.run('controls()');labels();}
 if(selected&&frame.type==='state_changed')afterState=true;
}
assert(selected&&afterState,'missing actual model/status then state sequence');
console.log('Chapter 18 production model projection passed');
