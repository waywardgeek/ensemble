// Run: node agent/GuiTests/recall-text.cjs
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const vm = require('node:vm');
const source = fs.readFileSync(path.join(__dirname, '../gui/wwwroot/app.js'), 'utf8');
const start = source.indexOf('function readable(part) {');
const end = source.indexOf('\n}', start) + 2;
assert(start >= 0 && end > start);
const context = vm.createContext({});
vm.runInContext(source.slice(start, end), context);
const text = '[Auto-recalled memories]\n> Exact supplied text <img src=x>\n> Second line.';
assert.equal(context.readable({type: 'recall', text, hashes: ['not display text']}), text);
assert.equal(context.readable({type: 'opaque', text: 'private signature'}), '');
assert.equal(context.readable({type: 'recall', text: ''}), '');
console.log('PASS recall readable text is exact; opaque metadata remains hidden');
