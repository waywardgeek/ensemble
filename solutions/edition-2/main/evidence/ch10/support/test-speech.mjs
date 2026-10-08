import assert from 'node:assert/strict';
import {speechState,assertNoRestoredSpeech} from './speech-probe.mjs';
const page={closed:false,connected:true,lastCauses:'{"typing":false,"speaking":false}',speech:{queue:[],current:null,cursors:new Map(),playbackReady:true}};
const app={service:{pending:[],active:null},pages:new Set([page])};
const before=speechState(app,0);assertNoRestoredSpeech(before,speechState(app,0));
for(const mutate of [s=>s.nativeAdmissions++,s=>s.serviceAdmissions++,s=>s.servicePending++,s=>s.serviceActive=true,s=>s.pages[0].queue++,s=>s.pages[0].current=true,s=>s.pages[0].causes.speaking=true]){
 const state=structuredClone(before);mutate(state);assert.throws(()=>assertNoRestoredSpeech(before,state),/restored speech/);
}
app.pages.clear();assertNoRestoredSpeech(before,speechState(app,0));
console.log('speech owner probe: valid baseline, seven intended negatives, closed Page pass; no browser/native speech run');
