const messages = document.querySelector('#messages');
const input = document.querySelector('#question');
const send = document.querySelector('#send');
const status = document.querySelector('#status');
let socket = null;
let ready = false;
function enabled(value) { ready = value; input.disabled = !value; send.disabled = !value; }
function add(role, text) {
  const article = document.createElement('article'); article.className = `message ${role}`;
  const title = document.createElement('strong'); title.textContent = role === 'user' ? 'You' : 'Measure X assistant';
  const body = document.createElement('div'); body.textContent = text;
  article.append(title, body); messages.append(article); messages.scrollTop = messages.scrollHeight;
}
function connect() {
  const previous = socket; socket = null; previous?.close();
  messages.replaceChildren(); input.value = ''; enabled(false); status.textContent = 'Starting a fresh conversation…';
  const url = new URL('./chat', location.href); url.protocol = location.protocol === 'https:' ? 'wss:' : 'ws:';
  const current = new WebSocket(url); socket = current;
  let failureReason = '';
  current.onmessage = event => {
    if (socket !== current) return;
    const message = JSON.parse(event.data);
    if (message.type === 'answer') { add('assistant', message.text); enabled(true); status.textContent = 'Ready for your next question.'; }
    else if (message.type === 'ready') { enabled(true); status.textContent = message.text; }
    else {
      if (message.type === 'error') failureReason = message.text;
      enabled(false); status.textContent = message.text;
    }
  };
  current.onclose = () => {
    if (socket !== current) return;
    enabled(false); messages.replaceChildren(); input.value = '';
    status.textContent = `${failureReason || 'Conversation closed or expired.'} Its chat has been discarded. Start a new conversation to continue.`;
  };
  current.onerror = () => {
    if (socket !== current) return;
    failureReason ||= 'Unable to connect. No question will be resent automatically.';
    status.textContent = failureReason;
  };
}
document.querySelector('#composer').addEventListener('submit', event => {
  event.preventDefault(); const question = input.value.trim();
  if (!ready || socket?.readyState !== WebSocket.OPEN || !question) return;
  if (new TextEncoder().encode(question).length > 4096) { status.textContent = 'Please keep the question under 4,096 UTF-8 bytes.'; return; }
  enabled(false); add('user', question); input.value = ''; status.textContent = 'Waiting for the answer…';
  try { socket.send(question); } catch { socket.close(); }
});
input.addEventListener('keydown', event => { if (event.key === 'Enter' && !event.shiftKey && !event.isComposing) { event.preventDefault(); document.querySelector('#composer').requestSubmit(); } });
document.querySelector('#new-session').addEventListener('click', connect);
document.querySelectorAll('.examples button').forEach(button => button.addEventListener('click', () => { if (ready) { input.value = button.textContent; input.focus(); } }));
window.addEventListener('pagehide', () => { socket?.close(); messages.replaceChildren(); input.value = ''; });
window.addEventListener('pageshow', event => { if (event.persisted) location.reload(); });
fetch('./about', { cache: 'no-store' }).then(response => { if (!response.ok) throw Error(); return response.json(); }).then(info => {
  document.querySelector('#model').textContent = `Model: ${info.model}. Exact application instructions:`;
  document.querySelector('#instructions').textContent = info.instructions;
  document.querySelector('#document-info').textContent = `Full reading copy: ${info.documentBytes.toLocaleString()} UTF-8 bytes. Session limit: ${Math.floor(info.sessionSeconds / 60)} minutes. ${info.ocrWarnings ? 'This reading copy contains recovery-OCR warnings.' : ''} Document SHA-256: ${info.documentSha256}`;
  connect();
}).catch(() => { status.textContent = 'The document information could not be loaded. No conversation was started.'; });
