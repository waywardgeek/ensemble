// Exercise the shipped page, not a second formatter. The browser driver is an
// explicitly supplied test dependency; Ensemble itself has no Node dependency.
import fs from 'node:fs/promises';
import path from 'node:path';
import net from 'node:net';
import { spawn } from 'node:child_process';
import { pathToFileURL } from 'node:url';
import assert from 'node:assert/strict';

function required(name) {
    const value = process.env[name];
    if (!value) throw new Error(`Set ${name}`);
    return value;
}
const dotnet = required('TTS_DOTNET');
const gui = required('TTS_GUI');
const chrome = required('TTS_CHROME');
const driver = await import(pathToFileURL(required('TTS_PUPPETEER')).href);
const puppeteer = driver.puppeteer || driver.default;
const work = required('WORKSPACE');
const output = required('TTS_LOG');
const [prompt, flag, typing] = process.argv.slice(2);
if (!prompt || (flag && (flag !== '--type-during-turn' || typing === undefined))) {
    throw new Error('Usage: tts-harness.sh PROMPT [--type-during-turn TEXT]');
}
const delay = milliseconds => new Promise(resolve => setTimeout(resolve, milliseconds));
async function port() {
    const server = net.createServer();
    await new Promise(resolve => server.listen(0, '127.0.0.1', resolve));
    const number = server.address().port;
    await new Promise(resolve => server.close(resolve));
    return number === 8088 ? port() : number;
}
const runtime = {};
for (const name of ['HOME', 'TMPDIR', 'PATH', 'DOTNET_ROOT', 'DOTNET_CLI_HOME']) {
    if (process.env[name]) runtime[name] = process.env[name];
}
const environment = { ...runtime, DOTNET_NOLOGO: '1', DOTNET_CLI_TELEMETRY_OPTOUT: '1',
    LLM_VENDOR: 'anthropic', LLM_MODEL: 'claude-3-haiku-20240307', LLM_API_KEY: 'local-speech-fixture',
    LLM_BASE_URL: required('LLM_BASE_URL') };
const number = await port();
const origin = `http://127.0.0.1:${number}`;
const args = [gui, '--port', String(number), '--no-save', '--settings', path.join(work, 'settings.json')];
if (process.env.STREAMING === 'off') args.push('--no-stream');
let host, browser, hostDone;
let stdout = '', stderr = '';
let interrupted = false;
const onSignal = () => { interrupted = true; };
process.on('SIGTERM', onSignal);
process.on('SIGINT', onSignal);
try {
    host = spawn(dotnet, args, { cwd: work, env: environment, stdio: ['ignore', 'pipe', 'pipe'] });
    host.stdout.on('data', chunk => { stdout += chunk; });
    host.stderr.on('data', chunk => { stderr += chunk; });
    hostDone = new Promise((resolve, reject) => {
        host.on('error', reject);
        host.on('exit', (code, signal) => resolve({ code, signal }));
    });
    let ready = false;
    for (let attempt = 0; attempt < 200 && !interrupted; attempt++) {
        if (host.exitCode !== null) throw new Error(`GUI exited during startup: ${stderr}`);
        try { ready = (await fetch(origin)).ok; } catch { /* Startup has not bound yet. */ }
        if (ready) break;
        await delay(50);
    }
    if (!ready || interrupted) throw new Error('GUI startup did not complete');
    browser = await puppeteer.launch({ executablePath: chrome, headless: true, env: runtime,
        userDataDir: path.join(work, 'profile'), args: ['--no-first-run', '--no-default-browser-check', '--mute-audio'] });
    const page = await browser.newPage();
    const errors = [];
    page.on('pageerror', error => errors.push(error.message));
    await page.goto(origin + '/?speech=transcript');
    await page.waitForFunction(() => typeof settings !== 'undefined' && settings && agentID, { timeout: 10000 });
    await page.evaluate(() => {
        window.harnessEnded = false;
        window.harnessTyping = null;
        socket.addEventListener('message', event => {
            if (JSON.parse(event.data).type === 'turn_ended') window.harnessEnded = true;
        });
        patchSettings({ tts_enabled: true });
    });
    await page.waitForFunction(() => settings.tts_enabled, { timeout: 10000 });
    await page.evaluate((text, draft) => {
        // Type after the first response fragment, not before prompt submission.
        if (draft !== null) {
            const observe = event => {
                const frame = JSON.parse(event.data);
                if (frame.type !== 'part_delta') return;
                socket.removeEventListener('message', observe);
                const input = document.querySelector('#prompt');
                input.value = draft;
                input.dispatchEvent(new Event('input', { bubbles: true }));
                window.harnessTyping = new Promise(resolve => setTimeout(() => {
                    input.value = '';
                    input.dispatchEvent(new Event('input', { bubbles: true }));
                    resolve();
                }, 100));
            };
            socket.addEventListener('message', observe);
        }
        document.querySelector('#prompt').value = text;
        document.querySelector('#compose').requestSubmit();
    }, prompt, typing ?? null);
    let ended = false;
    for (let attempt = 0; attempt < 1000 && !interrupted; attempt++) {
        ended = await page.evaluate(() => window.harnessEnded && activeSpeech === null && speechQueue.length === 0);
        if (ended) break;
        await delay(50);
    }
    if (!ended || interrupted) throw new Error('Turn did not finish before the harness deadline');
    await page.evaluate(async () => { if (window.harnessTyping) await window.harnessTyping; });
    assert.deepEqual(errors, []);
    const entries = await page.evaluate(() => window._ttsLog);
    await fs.writeFile(output, entries.map(entry => JSON.stringify(entry)).join('\n') + '\n');
} finally {
    try { if (browser) await browser.close(); }
    finally {
        if (host && hostDone) {
            if (host.exitCode === null) host.kill('SIGTERM');
            const timer = setTimeout(() => host.kill('SIGKILL'), 5000);
            const result = await hostDone.finally(() => clearTimeout(timer));
            await fs.writeFile(path.join(work, 'gui.stdout'), stdout);
            await fs.writeFile(path.join(work, 'gui.stderr'), stderr);
            assert.equal(result.code, 0, `GUI shutdown failed: ${JSON.stringify(result)} ${stderr}`);
        }
    }
}
