// Run with Node; this tests the shipped formatter without browser dependencies.
import assert from 'node:assert/strict';
import fs from 'node:fs';
import vm from 'node:vm';
const code=fs.readFileSync(new URL('../Gui/wwwroot/speech.js', import.meta.url), 'utf8');
const context=vm.createContext({});vm.runInContext(code,context);
const phrases=(text,offset,final)=>context.speechPhrases(text,offset,final);
const raw='# A heading\nwrapped **bold** text. Next `PascalCase` and snake_case with [a link](https://secret.invalid/path).\n\n```js\nconst SECRET = "do not read.";\n```\n\nFinal tail';
const expected=['A heading wrapped bold text.','Next Pascal Case and snake case with a link.','code block','Final tail'];
for(let width=1;width<=raw.length;width++){
 let text='',offset=0,heard=[];
 for(let at=0;at<raw.length;at+=width){text+=raw.slice(at,at+width);const next=phrases(text,offset,false);offset=next.offset;heard.push(...next.phrases);}
 const last=phrases(text,offset,true);heard.push(...last.phrases);
 assert.deepEqual(heard,expected,`fragment width ${width}`);
 assert.equal(last.offset,raw.length);
 assert.equal(phrases(text,last.offset,true).phrases.length,0);
}
assert.equal(phrases('one\ntwo',0,false).phrases.length,0);
assert.deepEqual([...phrases('1. first\n2. second\n\n',0,false).phrases],['first second']);
assert.deepEqual([...phrases('```js\nSECRET',0,true).phrases],['code block']);
console.log('PASS formatter: arbitrary fragmentation, fences before splitting, Markdown, identifiers, paragraphs, final cursor');
