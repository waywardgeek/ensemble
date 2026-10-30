// Run LAST against --crossover-host. Uses real TCP backpressure, not a mocked
// response stream, then requests the fixture's ordinary mounted shutdown.
const net = require('node:net');
const fs = require('node:fs');
const assert = require('node:assert/strict');
const { bulkHandle } = JSON.parse(fs.readFileSync('.crossover/runtime/fixture.json'));
const request = `GET /assistant/jobs/output?agent=bulk&handle=${bulkHandle} HTTP/1.1\r\nHost: 127.0.0.1:18202\r\nX-Ensemble-Output: 1\r\nConnection: close\r\n\r\n`;
function slowReader() {
 return new Promise((resolve, reject) => {
  const socket = net.connect(18202, '127.0.0.1');
  socket.once('connect', () => socket.write(request));
  socket.once('error', reject);
  socket.setTimeout(5000, () => socket.destroy(new Error('slow-reader setup timeout')));
  let data = Buffer.alloc(0);
  function receive(chunk) {
   data = Buffer.concat([data, chunk]);
   const boundary = data.indexOf('\r\n\r\n');
   if (boundary < 0) return;
   socket.pause();
   socket.removeListener('data', receive);
   const headers = data.subarray(0, boundary).toString();
   assert.match(headers, /^HTTP\/1.1 200/);
   const length = Number(/Content-Length: (\d+)/i.exec(headers)[1]);
   resolve({socket, length, received: data.length-boundary-4});
  }
  socket.on('data', receive);
 });
}
(async () => {
 const abort = await slowReader();
 assert(abort.received < abort.length);
 abort.socket.destroy();
 const health = await fetch('http://127.0.0.1:18202/assistant/');
 assert.equal(health.status, 200);
 console.log('PASS aborted output reader leaves mounted host usable');
 const stalled = await slowReader();
 let bytes = stalled.received;
 const closed = new Promise((resolve, reject) => {
  stalled.socket.on('data', chunk => { bytes += chunk.length; });
  stalled.socket.once('close', resolve);
  // ECONNRESET is a valid interrupted response, not a successful full download.
  stalled.socket.once('error', error => { if(error.code !== 'ECONNRESET') reject(error); });
  stalled.socket.setTimeout(5000, () => {
   stalled.socket.destroy();
   reject(new Error('mounted shutdown did not close slow output'));
  });
 });
 fs.writeFileSync('.crossover/runtime/stop', 'slow-output shutdown test\n');
 // Leave the reader paused long enough to create actual server backpressure.
 await new Promise(resolve => setTimeout(resolve, 300));
 stalled.socket.resume();
 await closed;
 assert(bytes < stalled.length, 'shutdown must cancel, not finish the stalled full output');
 console.log(`PASS mounted shutdown cancelled slow output (${bytes}/${stalled.length} bytes); fixture process exit must also be joined`);
})().catch(error => { console.error(error); process.exitCode=1; });
