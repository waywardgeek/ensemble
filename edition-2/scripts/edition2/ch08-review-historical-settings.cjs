#!/usr/bin/env node
// Run the complete frozen client with controlled DOM/socket objects. This checks
// the historical encoding story; it makes no actual-browser or provider claim.
const vm=require('vm'),crypto=require('crypto'),cp=require('child_process');
const revision='fd93089b8f53bdde42f9846c9aa36d5d0e6a9051',file='solutions/ch09/web/gui/gui.js';
const source=cp.execFileSync('git',['show',revision+':'+file],{encoding:'utf8'}),nodes=new Map();
function node(id){
 if(!nodes.has(id))nodes.set(id,{value:'',checked:false,textContent:'',style:{},classList:{add(){},remove(){},toggle(){}},addEventListener(){}});
 return nodes.get(id);
}
let socket;
class Socket{constructor(){socket=this;}send(){}}
Socket.OPEN=1;
const scope={document:{getElementById:node,querySelector:node,querySelectorAll:()=>[],documentElement:node('html'),addEventListener(){},removeEventListener(){}},window:{},location:{protocol:'http:',host:'local'},WebSocket:Socket,ArtifactScroll:class{handleMessage(){}},TTS:{enabled:true,rate:1,cancel(){}},setTimeout(){}};
vm.runInNewContext(source,scope);
socket.onmessage({data:JSON.stringify({type:'current_settings',settings:{theme:'dark',tts_enabled:true,temperature:1}})});
const enabled=scope.TTS.enabled;
node('set-temperature').value=-5;
socket.onmessage({data:JSON.stringify({type:'settings_changed',settings:{theme:'dark'}})});
if(enabled!==true||scope.TTS.enabled!==false||node('set-tts-enabled').checked!==false||node('set-temperature').value!==-5)throw Error('historical control differs');
console.log(JSON.stringify({historical_commit:revision,source:file,sha256:crypto.createHash('sha256').update(source).digest('hex'),scope:'Whole frozen gui.js executed in vm with controlled DOM/socket; no browser or live provider claim',positive_enabled:enabled,after_omitted_false:scope.TTS.enabled,checkbox:node('set-tts-enabled').checked,temperature_after_omitted_zero:node('set-temperature').value},null,2));
