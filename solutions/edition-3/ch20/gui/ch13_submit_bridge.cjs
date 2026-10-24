'use strict';
// Test-only controlled page: actual app.js handlers; Go supplies real WS frames.
const fixture=require('./client_round1_test.cjs'),readline=require('node:readline');
const f=fixture();
readline.createInterface({input:process.stdin}).on('line',line=>{
 try{
  const m=JSON.parse(line);
  if(m.op==='frame')f.run('socket.onmessage')({data:JSON.stringify(m.frame)});
  else if(m.op==='input'){f.$('input').value=m.text;f.$('input').oninput({isTrusted:true});}
  else if(m.op==='submit')f.$('composer').onsubmit({preventDefault(){}});
  else throw Error('unknown bridge operation');
  console.log(JSON.stringify({sent:f.sent.splice(0),channel:f.channel.read(),epoch:f.run('agents.get(selected)?.state.epoch'),status:f.channel.status(),utterances:f.utterances.map(u=>u.text)}));
 }catch(error){console.log(JSON.stringify({error:error.message}));}
});
