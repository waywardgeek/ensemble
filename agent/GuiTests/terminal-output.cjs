const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
(async () => {
    const source = fs.readFileSync(path.join(__dirname, '../skills/shared/terminal-output.js'), 'utf8');
    const { terminalRuns } = await import('data:text/javascript;base64,' + Buffer.from(source).toString('base64'));
    const bytes = text => new Uint8Array(Buffer.from(text));
    const plain = runs => runs.map(run => run.text).join('');
    assert.equal(plain(terminalRuns(bytes('<img src=x>\nhello'))), '<img src=x>\nhello');
    const styled = terminalRuns(bytes('before\x1b[1;31mred\x1b[0mafter'));
    assert.deepEqual(styled, [
        {text:'before', bold:false, foreground:null, background:null},
        {text:'red', bold:true, foreground:1, background:null},
        {text:'after', bold:false, foreground:null, background:null}]);
    assert.equal(plain(terminalRuns(bytes('\x1b]8;;https://invalid\x07label\x1b]8;;\x1b\\'))), 'label');
    assert.equal(plain(terminalRuns(bytes('a\x1b[2Jb\x1bPsecret\x1b\\c'))), 'abc');
    assert.equal(plain(terminalRuns(bytes('a\rb\r\nc'))), 'a\nb\nc');
    assert.equal(plain(terminalRuns(new Uint8Array([0x82,0xac,0x41]), true)), 'A');
    assert.equal(plain(terminalRuns(new Uint8Array([0x41,0xe2,0x82]))), 'A');
    assert.equal(plain(terminalRuns(bytes('A€'))), 'A€');
    assert.equal(plain(terminalRuns(bytes('a\x1b[31'))), 'a');
    assert.equal(terminalRuns(bytes('\x1b[38;2;1;31;0mplain'))[0].foreground, null);
    assert.equal(terminalRuns(bytes('\x1b[38;2;1;31;0mplain'))[0].bold, false);
    assert.equal(terminalRuns(bytes('plain'), true)[0].foreground, null);
    assert.throws(() => terminalRuns(new Uint8Array(65537)), /Invalid terminal window/);
    console.log('terminal window decoder: 13 checks passed');
})().catch(error => { console.error(error); process.exitCode = 1; });
