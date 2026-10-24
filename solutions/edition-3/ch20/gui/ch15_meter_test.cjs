'use strict';
// Production message handler/DOM projection, without browser or native audio.
const assert=require('node:assert/strict'),fixture=require('./client_round1_test.cjs');
assert.match(require('node:fs').readFileSync(require('node:path').join(__dirname,'web/gui/index.html'),'utf8'),/id="cache-meter"/);
const f=fixture();
const known={last:{input:10,output:4,cache_write:10,cache_read:80},session:{input:10,output:4,cache_write:10,cache_read:80},requests:1,last_hit_rate:.8,session_hit_rate:.8,cost:{usd:.0001,status:'known'},buckets:[{vendor:'anthropic',surface:'messages',model:'served-model',requests:1,usage:{input:10,output:4,cache_write:10,cache_read:80},cost:{usd:.0001,status:'known'}}],cache:{marker_count:4,marker_reason:'test policy',shared_segments:2,shared_bytes:20,first_difference:'messages[2]'}};
const receive=m=>f.run('socket.onmessage({data:'+JSON.stringify(JSON.stringify({type:'usage',agent_id:'agent',meter:m,total:m.session,input:m.requests?m.last.input:0,output:m.requests?m.last.output:0,cache_read:m.requests?m.last.cache_read:0,cache_write:m.requests?m.last.cache_write:0,hit_rate:m.requests?m.session_hit_rate:0,cost_usd:m.cost.status==='known'?m.cost.usd:null,priced:m.cost.status==='known'}))+'})');
receive(known);
assert.equal(f.$('cache-meter').textContent,'$0.0001 · In 100 · Cache 80 · Hit 80.0%');
assert.match(f.$('cache-meter').title,/served-model/);assert.match(f.$('cache-meter').title,/not provider cache hits/);assert.match(f.$('cache-meter').title,/uncached input 10/);
const partial={...known,last_hit_rate:null,session_hit_rate:null,last:{input:-1,output:5,cache_write:0,cache_read:-1},session:{input:-1,output:9,cache_write:10,cache_read:-1},cost:{usd:.0001,status:'partial'}};
receive(partial);assert.equal(f.$('cache-meter').textContent,'partial (known $0.0001) · In unknown · Cache unknown · Hit unknown');assert.match(f.$('cache-meter').title,/cache read unknown/);
const empty={...known,last:{input:-1,output:-1,cache_write:-1,cache_read:-1},session:{input:0,output:0,cache_write:0,cache_read:0},requests:0,last_hit_rate:null,session_hit_rate:null,cost:{usd:0,status:'known'},buckets:[],cache:{marker_count:0,marker_reason:'',shared_segments:0,shared_bytes:0}};
f.run('replay({client:"owner",session:"new-process",agents:[{id:"agent",state:'+JSON.stringify(f.state())+',config:{vendor:"anthropic",model:"fixture",surface:"messages"},meter:'+JSON.stringify(empty)+'}],events:[],partials:[]})');assert.equal(f.$('cache-meter').textContent,'$0.0000 · In 0 · Cache 0 · Hit 0.0%');
console.log('PASS Chapter 15 production meter handler/DOM: known, partial, reset');
