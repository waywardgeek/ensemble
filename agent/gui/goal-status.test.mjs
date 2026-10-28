import test from 'node:test';
import assert from 'node:assert/strict';
import { renderGoalStatus } from './web/gui/goal-status.js';

test('goal status is accessible text, not an inferred completion or browser timer',()=>{
 const element={textContent:'',title:'',setAttribute(name,value){this[name]=value;}};
 renderGoalStatus(element,null);assert.equal(element.textContent,'Goal: loading');
 renderGoalStatus(element,null,'Goal: disconnected');assert.equal(element.textContent,'Goal: disconnected');
 renderGoalStatus(element,{file:'',phase:'none'});assert.equal(element.textContent,'No goal');
 const base={file:'<GOAL.md>',idle_seconds:480,enabled:true};
 for(const phase of ['working','waiting','paused','error']){
  renderGoalStatus(element,{...base,phase,error:phase==='error'?'file unavailable':'',next_reminder:'2030-01-01T00:00:00Z'});
  assert.equal(element.textContent,`Goal: ${phase}`);
  assert.ok(element.title.includes('File: <GOAL.md>'));
  assert.equal(element['aria-label'],element.title);
  assert.equal(element.title.includes('Next reminder:'),phase==='waiting');
  assert.equal(element.title.includes('Error: file unavailable'),phase==='error');
 }
 renderGoalStatus(element,{...base,enabled:false,phase:'working'});
 assert.equal(element.textContent,'Goal: working · reminders off');
 renderGoalStatus(element,{...base,enabled:false,phase:'off'});
 assert.equal(element.textContent,'Goal: reminders off');
 assert.ok(!element.title.includes('completed'));
});
