// Public WebSocket lifecycle tests using a real Agent and local fake provider only.
const assert = require('node:assert/strict');
const fs = require('node:fs');
const os = require('node:os');
const path = require('node:path');
const http = require('node:http');
const { spawn } = require('node:child_process');
const { chromium } = require(process.env.PLAYWRIGHT_MODULE || 'playwright');
const dll = path.resolve(process.argv[2] || '');
if (!process.argv[2]) throw Error('Pass built DocumentChat.dll');
const scratch = fs.mkdtempSync(path.join(os.tmpdir(), 'document-chat-limits-'));
const doc = path.join(scratch, 'reference.txt');
fs.writeFileSync(doc, '[PDF PAGE 1]\nSynthetic reference.');
const requests = [];
const requestBytes = [];
let largeAnswers = false;
let heldClosed = 0;
const provider = http.createServer(async (req, res) => {
  let text = '';
  for await (const chunk of req) text += chunk;
  const body = JSON.parse(text);
  requests.push(body);
  requestBytes.push(Buffer.byteLength(text));
  if (JSON.stringify(body.input).includes('HOLD_REQUEST')) {
    res.on('close', () => heldClosed++);
    return; // Real transport must close this response on Agent destruction.
  }
  res.setHeader('content-type', 'application/json');
  res.end(JSON.stringify({ id:'resp_limits', object:'response', status:'completed', model:'gpt-6-astra',
    output:[{id:'msg_limits', type:'message', role:'assistant', status:'completed',
      content:[{type:'output_text', text:largeAnswers ? `ANSWER_${requests.length}_` + 'x'.repeat(18000) : 'SAFE_ANSWER', annotations:[]}]}],
    usage:{input_tokens:10, output_tokens:2, total_tokens:12} }));
});
const listen = server => new Promise(resolve => server.listen(0, '127.0.0.1', () => resolve(server.address().port)));
const sleep = ms => new Promise(resolve => setTimeout(resolve, ms));
async function until(predicate, label) {
  const end = Date.now() + 15000;
  while (Date.now() < end) { if (await predicate()) return; await sleep(25); }
  throw Error('Timed out: ' + label);
}
let browser;
let activeHost;
async function start(providerPort, extra = []) {
  const reservation = http.createServer();
  const port = await listen(reservation);
  await new Promise(resolve => reservation.close(resolve));
  const base = `http://127.0.0.1:${port}`;
  const child = spawn(process.env.DOTNET || 'dotnet', [dll, '--document', doc, '--port', String(port),
    '--base-url', `http://127.0.0.1:${providerPort}`, ...extra], {
    cwd:scratch, env:{PATH:'/usr/bin:/bin', HOME:scratch, TMPDIR:scratch,
      DOCUMENT_CHAT_API_KEY:'synthetic-only', DOTNET_CLI_HOME:scratch,
      DOTNET_SKIP_FIRST_TIME_EXPERIENCE:'1', DOTNET_GENERATE_ASPNET_CERTIFICATE:'false'} });
  let stderr = '';
  child.stdout.resume(); child.stderr.on('data', b => { stderr = (stderr + b).slice(-8192); });
  const exited = new Promise(resolve => child.on('exit', (code, signal) => resolve({code, signal})));
  const host = {base, child, exited, stderr:() => stderr};
  activeHost = host;
  await until(async () => {
    assert.equal(child.exitCode, null, 'early host exit: ' + stderr);
    try { return (await fetch(base + '/about')).ok; } catch { return false; }
  }, 'host readiness');
  const page = await browser.newPage();
  // /about has no application script, so only sockets explicitly created below exist.
  await page.goto(base + '/about');
  await page.evaluate(() => {
    window.clients = [];
    window.openClient = () => {
      const c = {messages:[], closed:false, socket:new WebSocket(location.origin.replace('http:', 'ws:') + '/chat')};
      c.socket.onmessage = e => c.messages.push(JSON.parse(e.data));
      c.socket.onclose = () => { c.closed = true; };
      clients.push(c); return clients.length - 1;
    };
  });
  host.page = page;
  return host;
}
async function stop(host) {
  host.child.kill('SIGTERM');
  let timer;
  let result;
  try {
    const timeout = new Promise(resolve => { timer = setTimeout(() => resolve(null), 15000); });
    result = await Promise.race([host.exited, timeout]);
  } finally { clearTimeout(timer); }
  if (!result) { host.child.kill('SIGKILL'); await host.exited; throw Error('Graceful shutdown failed'); }
  assert.equal(result.code, 0, 'host shutdown: ' + host.stderr());
  await host.page.close();
  activeHost = null;
}
async function open(host, ready = true) {
  const id = await host.page.evaluate(() => openClient());
  if (ready) await host.page.waitForFunction(i => clients[i].messages.some(m => m.type === 'ready'), id, {timeout:15000});
  return id;
}
async function closed(host, id) {
  await host.page.waitForFunction(i => clients[i].closed, id, {timeout:15000});
}
async function send(host, id, text) {
  await host.page.evaluate(({id,text}) => clients[id].socket.send(text), {id,text});
}
async function close(host, id) {
  await host.page.evaluate(i => clients[i].socket.close(), id);
  await closed(host, id);
}
(async () => {
  const providerPort = await listen(provider);
  browser = await chromium.launch({headless:true, channel:'chrome'});
  try {
    let host = await start(providerPort);
    const denied = await new Promise((resolve, reject) => {
      const req = http.get(host.base + '/chat', {headers:{Connection:'Upgrade', Upgrade:'websocket',
        'Sec-WebSocket-Version':'13', 'Sec-WebSocket-Key':'dGhlIHNhbXBsZSBub25jZQ==', Origin:'https://wrong.example'}}, res => {
        res.resume(); resolve(res.statusCode);
      });
      req.on('upgrade', (_, socket) => { socket.destroy(); reject(Error('Wrong origin accepted')); });
      req.on('error', reject); req.setTimeout(5000, () => req.destroy(Error('upgrade timeout')));
    });
    assert.equal(denied, 403);
    for (const invalid of ['', ' \t ', 'bad\0input', 'x'.repeat(4097), 'é'.repeat(2049)]) {
      const id = await open(host); await send(host, id, invalid); await closed(host, id);
    }
    const binary = await open(host);
    await host.page.evaluate(i => clients[i].socket.send(new Uint8Array([1,2,3])), binary);
    await closed(host, binary);
    assert.equal(requests.length, 0, 'invalid input reached provider');
    const slots = [];
    for (let i = 0; i < 8; i++) slots.push(await open(host));
    const excess = await open(host, false); await closed(host, excess);
    assert.equal(await host.page.evaluate(i => clients[i].messages.length, excess), 0);
    await close(host, slots.pop());
    // Close notification can precede joined destruction/slot release by a small interval.
    await sleep(100);
    slots.push(await open(host));
    for (const id of slots) await close(host, id);
    console.log('PASS: origin, malformed/oversized UTF-8, binary, eight-slot refusal and recovery; zero provider calls.');
    const held = await open(host); await send(host, held, 'HOLD_REQUEST_DISCONNECT');
    await until(() => requests.length === 1, 'held request arrival');
    await close(host, held);
    await until(() => heldClosed === 1, 'disconnect cancels actual HTTP request');
    const fresh = await open(host); await send(host, fresh, 'FRESH_VISITOR');
    await host.page.waitForFunction(i => clients[i].messages.some(m => m.type === 'answer'), fresh);
    assert.equal(requests.length, 2);
    assert(!JSON.stringify(requests[1]).includes('HOLD_REQUEST'), 'disconnected dialogue leaked');
    await close(host, fresh);
    const shutdown = await open(host); await send(host, shutdown, 'HOLD_REQUEST_SHUTDOWN');
    await until(() => requests.length === 3, 'shutdown request arrival');
    await stop(host);
    await until(() => heldClosed === 2, 'shutdown cancels actual HTTP request');
    console.log('PASS: disconnect and shutdown during inference cancel HTTP; fresh visitor has no prior dialogue; host exit zero.');
    host = await start(providerPort, ['--session-seconds', '2']);
    const expiring = await open(host); await send(host, expiring, 'HOLD_REQUEST_EXPIRY');
    await until(() => requests.length === 4, 'expiry request arrival');
    await closed(host, expiring);
    await until(() => heldClosed === 3, 'expiry cancels actual HTTP request');
    await stop(host);
    host = await start(providerPort, ['--context-bytes', '32768']);
    const tooSmall = await open(host, false); await closed(host, tooSmall);
    const messages = await host.page.evaluate(i => clients[i].messages, tooSmall);
    assert(messages.some(m => m.type === 'error' && m.text.includes('No document text was dropped')));
    assert.equal(requests.length, 4, 'context refusal reached provider');
    await stop(host);
    console.log('PASS: absolute expiry during inference and full-document context refusal; exactly four fake requests so far.');
    host = await start(providerPort);
    const capped = await open(host);
    for (let i = 0; i < 20; i++) {
      await send(host, capped, 'Question ' + i);
      await host.page.waitForFunction(({id,count}) => clients[id].messages.filter(m => m.type === 'answer').length === count,
        {id:capped, count:i + 1});
    }
    await send(host, capped, 'Question twenty one'); await closed(host, capped);
    assert.equal(requests.length, 24, 'twenty-first question reached provider');
    const overloaded = await open(host); await send(host, overloaded, 'HOLD_REQUEST_OVERLOAD');
    await until(() => requests.length === 25, 'overload held request arrival');
    await host.page.evaluate(i => {
      clients[i].socket.send('queued one');
      clients[i].socket.send('queued two');
      clients[i].socket.send('queued three');
    }, overloaded);
    await closed(host, overloaded);
    await until(() => heldClosed === 4, 'overload cancels active request');
    assert.equal(requests.length, 25, 'overloaded queued input reached provider');
    await stop(host);
    console.log('PASS: twenty-question cap and queue overload; exactly 25 fake requests; all hosts exited zero.');
    largeAnswers = true;
    const allowance = 140000;
    host = await start(providerPort, ['--context-bytes', String(allowance)]);
    const growing = await open(host);
    const first = requests.length;
    const answers = [];
    const reference = fs.readFileSync(doc, 'utf8');
    for (let i = 0; i < 10; i++) {
      // These characters expand in JSON; UTF-8 input length alone isn't wire size.
      const question = 'GROW_' + i + '_' + '<>&"\\'.repeat(700);
      await send(host, growing, question);
      await host.page.waitForFunction(({id,count}) => clients[id].closed ||
        clients[id].messages.filter(m => m.type === 'answer').length === count,
        {id:growing, count:i + 1});
      if (await host.page.evaluate(id => clients[id].closed, growing)) break;
      assert.equal(requests.length, first + i + 1);
      const body = requests.at(-1);
      const wire = JSON.stringify(body);
      assert(requestBytes.at(-1) <= allowance, 'admitted request exceeded byte allowance');
      const input = JSON.stringify(body.input);
      const escapedReference = JSON.stringify(reference).slice(1, -1);
      assert.equal(input.split(escapedReference).length - 1, 1, 'full reference missing or duplicated');
      for (const answer of answers) assert(input.includes(answer), 'prior answer dropped before refusal');
      assert(wire.includes('GROW_' + i), 'new question absent');
      answers.push(`ANSWER_${requests.length}_` + 'x'.repeat(18000));
    }
    assert(answers.length >= 2 && answers.length < 10, 'did not reach accumulated-context boundary');
    await closed(host, growing);
    const refused = await host.page.evaluate(id => clients[id].messages, growing);
    assert(refused.some(m => m.type === 'error' && m.text.includes('reached its limit')));
    assert.equal(requests.length, first + answers.length, 'refused follow-up reached provider');
    largeAnswers = false;
    const recovered = await open(host);
    await send(host, recovered, 'NEW_AFTER_CONTEXT_LIMIT');
    await host.page.waitForFunction(id => clients[id].messages.some(m => m.type === 'answer'), recovered);
    assert(!JSON.stringify(requests.at(-1)).includes('GROW_'), 'exhausted conversation leaked');
    await close(host, recovered);
    await stop(host);
    console.log(`PASS: accumulated context refused after ${answers.length} intact answers; escaped questions bounded; fresh session recovered. ${requests.length} total fake requests.`);

  } finally {
    try { if (activeHost) await stop(activeHost); }
    finally {
      if (browser) await browser.close();
      provider.closeAllConnections();
      await new Promise(resolve => provider.close(resolve));
      fs.rmSync(scratch, {recursive:true, force:true});
    }
  }
})().catch(error => { console.error(error); process.exitCode = 1; });
