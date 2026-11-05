'use strict';
// Real public page, model turns and custom-tool Jobs; no application globals.
const { chromium } = require(process.env.PLAYWRIGHT_MODULE);
const assert = require('node:assert/strict');
const base = process.env.ARTIFACT_GUI_URL;
const origin = new URL(base).origin;
assert.equal(new URL(base).hostname, '127.0.0.1');
assert(!['8088', '8089', '8090'].includes(new URL(base).port));
(async () => {
  const browser = await chromium.launch({headless: true, channel: 'chrome'});
  try {
    const page = await browser.newPage({viewport: {width: 390, height: 844}, hasTouch: true, locale: 'en-US', timezoneId: 'Pacific/Honolulu'});
    page.setDefaultTimeout(20000);
    const errors = [], transfers = [];
    page.on('pageerror', error => errors.push(error.message));
    page.on('request', request => { if (request.url().includes('/work-products/download')) transfers.push(request.url()); });
    async function statusUntil(predicate) {
      for (let attempt = 0; attempt < 200; attempt++) {
        const state = await (await page.request.get(origin + '/grade/status')).json();
        assert.equal(state.failure, null, 'Actions provider oracle');
        if (predicate(state)) return state;
        await new Promise(resolve => setTimeout(resolve, 100));
      }
      throw new Error('Actions state did not reach expected boundary');
    }
    async function touchTarget(control, label) {
      const bounds = await control.boundingBox();
      assert(bounds && bounds.width >= 44 && bounds.height >= 44,
        label + ' has a visible 44-pixel touch target: ' + JSON.stringify(bounds));
    }
    async function checkRecordedTiming(card, fact) {
      assert(fact && typeof fact.recorded === 'string', 'canonical result has recorded time');
      const stamp = card.locator(':scope > time.recorded-time');
      assert.equal(await stamp.getAttribute('datetime'), fact.recorded,
        'display time equals canonical recorded fact, not browser arrival');
      const local = await page.evaluate(time => new Date(time).toLocaleString(), fact.recorded);
      assert.equal(await stamp.innerText(), 'Recorded ' + local, 'recorded time uses local timezone');
      assert((await stamp.getAttribute('title')).includes(fact.recorded), 'exact timestamp remains inspectable');
      const duration = card.locator(':scope > .observation-duration');
      if (fact.milliseconds === null) {
        assert.equal(await duration.count(), 0, 'unknown observation duration is not invented');
      } else {
        assert(fact.milliseconds >= 0, 'canonical observation duration is nonnegative');
        assert.equal(await duration.innerText(), 'Observation: ' +
          (fact.milliseconds / 1000).toLocaleString('en-US', {maximumFractionDigits: 3}) + ' s',
          'display duration equals recorded observation');
        assert.equal(await duration.getAttribute('title'), 'Dispatch to returned observation: ' +
          fact.milliseconds + ' ms; not background Job runtime', 'exact observation duration remains inspectable');
      }
    }
    // Replace only browser clipboard and speech boundaries, never renderer or protocol state.
    await page.addInitScript(() => {
      class Utterance { constructor(text) { this.text = text; } }
      const speech = {
        speaking: false, pending: false, current: null, utterances: [], cancellations: 0,
        getVoices() { return []; }, addEventListener() {},
        speak(item) {
          if (!item.text) return; // Chrome's silent activation is not narration.
          this.utterances.push(item.text); this.current = item; this.speaking = true;
          if (this.automaticCompletion) queueMicrotask(() => { if (this.current === item) this.finish(); });
        },
        cancel() { this.cancellations++; this.current = null; this.speaking = false; },
        finish() {
          const item = this.current; this.current = null; this.speaking = false;
          item?.onend?.();
        },
        drain() {
          for (let count = 0; this.current; count++) {
            if (count >= 1000) throw new Error('Synthetic speech queue did not drain');
            this.finish();
          }
        }
      };
      Object.defineProperty(window, 'speechSynthesis', {value: speech});
      Object.defineProperty(window, 'SpeechSynthesisUtterance', {value: Utterance});
      window.gradeSpeech = speech;
      window.gradeClipboard = {writes: [], denied: false};
      Object.defineProperty(navigator, 'clipboard', {configurable: true, value: {
        writeText: async text => {
          window.gradeClipboard.writes.push(text);
          if (window.gradeClipboard.denied) throw new DOMException('Denied', 'NotAllowedError');
        }
      }});
    });
    await page.goto(base);
    await page.getByRole('button', {name: 'Chat', exact: true}).click();
    const initial = await statusUntil(() => true);
    await page.locator('#prompt').fill('exercise Actions ' + initial.nonce + '\n```\nhuman code\n```');
    await page.locator('#send').click();
    await statusUntil(state => state.entered && state.running === 1);
    await page.locator('.pane-tab[data-pane="tools"]').click();
    const running = page.locator('[data-kind="tool_result"]');
    await running.waitFor();
    assert((await running.locator('.job-state').innerText()).includes('state when observed: running'),
      'real-running-observation must display running, not finished');
    assert.equal(await running.locator('.tool-state:not(.job-state)').innerText(), 'Result returned', 'returned observation is not completed work');
    const resultID = await running.getAttribute('data-part-id');
    await running.getByRole('button', {name: 'View call', exact: true}).click();
    const call = page.locator('[data-kind="tool_call"]:focus');
    assert.equal(await call.count(), 1, 'View call focuses a real call card');
    const callID = await call.getAttribute('data-part-id');
    assert((await call.innerText()).includes('held_action'), 'paired call is the actual held tool');
    await call.getByRole('button', {name: 'View latest result', exact: true}).click();
    assert.equal(await page.locator('[data-kind="tool_result"]:focus').getAttribute('data-part-id'), resultID,
      'call navigation returns to its actual observation');
    const keyboardCall = page.locator(`[data-part-id="${resultID}"]`).getByRole('button', {name: 'View call', exact: true});
    const keyboardResult = page.locator(`[data-part-id="${callID}"]`).getByRole('button', {name: 'View latest result', exact: true});
    await touchTarget(keyboardCall, 'View call');
    await touchTarget(keyboardResult, 'View latest result');
    await keyboardCall.focus();
    await keyboardCall.press('Enter');
    assert.equal(await page.locator('[data-kind="tool_call"]:focus').getAttribute('data-part-id'), callID,
      'Enter activates call navigation');
    await keyboardResult.focus();
    await keyboardResult.press('Space');
    assert.equal(await page.locator('[data-kind="tool_result"]:focus').getAttribute('data-part-id'), resultID,
      'Space activates result navigation');
    assert.equal(await page.locator('.work-product-links').count(), 0, 'running result has no offer control');
    const observed = (await statusUntil(state => state.running === 1)).observations[0];
    assert(observed.milliseconds > 0, 'admitted real observation has measured duration');
    await checkRecordedTiming(running, observed);
    assert(Number.isSafeInteger(observed.handle) && observed.handle > 0, 'canonical observation has a Job handle');
    const outputURL = new URL('jobs/output', base);
    outputURL.search = new URLSearchParams({agent: '', handle: String(observed.handle), download: '1'});
    const outputHeaders = {'X-Ensemble-Output': '1', Origin: origin};
    const runningOutput = await page.request.get(outputURL.href, {headers: outputHeaders});
    assert.equal(runningOutput.status(), 409, 'running Job output is not a completed download');
    const details = page.locator(`[data-part-id="${resultID}"]`).locator(':scope > details');
    const summary = details.locator(':scope > summary');
    await page.setViewportSize({width: 390, height: 480});
    await summary.click();
    await summary.focus();
    const readerNode = await summary.elementHandle();
    const receivedText = await details.locator('pre').textContent();
    assert(await details.evaluate(node => node.open), 'reader opens actual received payload');
    const readingPosition = await details.locator('pre').evaluate(async node => {
      const range = document.createRange();
      range.selectNodeContents(node);
      const selection = window.getSelection();
      selection.removeAllRanges();
      selection.addRange(range);
      const pane = document.querySelector('#tool-transcript');
      pane.scrollTop = pane.scrollHeight;
      await new Promise(resolve => requestAnimationFrame(() => requestAnimationFrame(resolve)));
      return {text: selection.toString(), top: pane.scrollTop,
        overflow: pane.scrollHeight - pane.clientHeight};
    });
    assert(readingPosition.text.length > 0, 'reader selects actual raw result text');
    assert(readingPosition.overflow > 0, 'reader fixture has genuine scroll overflow');
    await page.request.post(origin + '/grade/release-action');
    const completed = await statusUntil(state => state.requests === 3 && state.done === 1);
    assert.equal(completed.executions, 1);
    const oldResult = page.locator(`[data-part-id="${resultID}"]`);
    assert((await oldResult.locator('.job-state').innerText()).includes('state when observed: running'),
      'later completion must not rewrite the earlier observation');
    const terminal = page.locator('[data-kind="tool_result"]').filter({has: page.locator('.job-state', {hasText: 'state when observed: finished'})});
    await terminal.waitFor();
    assert(await readerNode.evaluate(node => node.isConnected && document.activeElement === node),
      'later-observation preserves the actual focused reader control');
    assert(await details.evaluate(node => node.open), 'later-observation preserves expanded payload');
    assert.equal(await details.locator('pre').textContent(), receivedText,
      'later-observation preserves complete original payload');
    const afterReading = await page.locator('#tool-transcript').evaluate(pane => ({
      text: window.getSelection().toString(), top: pane.scrollTop,
      overflow: pane.scrollHeight - pane.clientHeight
    }));
    assert.equal(afterReading.text, readingPosition.text,
      'later-observation preserves actual raw text selection');
    assert(afterReading.overflow > readingPosition.overflow,
      'new observations extend the scrollable transcript');
    assert(Math.abs(afterReading.top - readingPosition.top) <= 1,
      'later-observation preserves reader scroll instead of following output: ' +
      JSON.stringify({before: readingPosition, after: afterReading}));
    await page.evaluate(() => window.getSelection().removeAllRanges());
    await readerNode.dispose();
    await page.setViewportSize({width: 390, height: 844});
    assert.deepEqual(completed.observations[0], observed, 'later completion does not rewrite recorded timing');
    await checkRecordedTiming(oldResult, observed);
    await checkRecordedTiming(terminal, completed.observations[1]);
    assert.equal(await terminal.locator('.tool-state:not(.job-state)').innerText(), 'Result returned');
    assert((await terminal.textContent()).includes(completed.output), 'full pathname result is retained');
    assert.equal(await page.locator('.work-product-links').count(), 0, 'pathname prose grants no work-product control');
    assert.equal(transfers.length, 0, 'rendering pathname output transfers nothing');
    const ownedOutput = await page.request.get(outputURL.href, {headers: outputHeaders});
    assert.equal(ownedOutput.status(), 200, 'actual owning Agent can read completed Job output');
    assert((await ownedOutput.text()).includes(completed.output), 'owned Job download contains the actual result');
    outputURL.searchParams.set('agent', 'legacy-timing');
    const foreignOutput = await page.request.get(outputURL.href, {headers: outputHeaders});
    assert.equal(foreignOutput.status(), 404, 'foreign Agent cannot read another Agent Job by its handle');
    assert.equal(foreignOutput.headers()['content-disposition'], undefined,
      'foreign Job refusal is not an attachment');
    await page.getByRole('button', {name: 'Chat', exact: true}).click();
    const answer = page.locator('#transcript [data-role="assistant"][data-kind="text"]').filter({hasText: initial.final});
    await answer.waitFor();
    const snippets = answer.locator('.code-snippet');
    const expected = [
      '  界 😀 ' + initial.nonce + '\n\n\tend;  \n', '  quoted\n\n', 'nested\n',
      '```\n<script>window.snippetExecuted = true</script>\n',
      ('  long code 界 ' + initial.nonce + '\n').repeat(200), ''
    ];
    assert(expected[4].length > 1800, 'long fixture exceeds both preview limits');
    assert.equal(await answer.locator('[data-action="copy-code"]').count(), expected.length,
      'only closed assistant fences have copy controls');
    assert.equal(await page.locator('#transcript [data-role="human"]').filter({hasText: 'human code'}).count(), 1,
      'fenced human prompt is actually rendered');
    assert.equal(await snippets.nth(4).locator('details').getAttribute('open'), null,
      'full-payload copy starts with a collapsed long preview');
    assert.equal(await page.locator('#transcript [data-role="human"] [data-action="copy-code"]').count(), 0,
      'human fenced prompt has no assistant copy control');
    assert.deepEqual(await page.evaluate(() => window.gradeClipboard.writes), [], 'copy requires deliberate click');
    const draft = 'Unsent draft 界 ' + initial.nonce;
    await page.locator('#prompt').fill(draft);
    for (let index = 0; index < expected.length; index++) {
      const snippet = snippets.nth(index);
      const copy = snippet.locator('[data-action="copy-code"]');
      await touchTarget(copy, 'Copy');
      if (index < 2) {
        await copy.focus();
        await copy.press(index === 0 ? 'Enter' : 'Space');
      } else if (index === 2) await copy.tap();
      else await copy.click();
      assert.equal(await page.evaluate(() => window.gradeClipboard.writes.at(-1)), expected[index],
        'snippet logical payload exact at index ' + index);
      assert.equal(await snippet.locator('[role="status"]').innerText(), 'Copied');
    }
    assert.equal(await page.evaluate(() => window.snippetExecuted), undefined, 'hostile code stays literal');
    assert.equal(await answer.locator('script').count(), 0, 'literal markup is not a script node');
    await page.evaluate(() => { window.gradeClipboard.denied = true; });
    await snippets.nth(4).locator('[data-action="copy-code"]').click();
    const manual = snippets.nth(4).locator('textarea');
    assert.equal(await manual.inputValue(), expected[4], 'denied clipboard fallback contains full payload');
    assert(await manual.evaluate(node => node.readOnly && document.activeElement === node &&
      node.selectionStart === 0 && node.selectionEnd === node.value.length), 'manual payload is readonly and selected');
    assert.match(await snippets.nth(4).locator('[role="status"]').innerText(), /Clipboard unavailable/);
    assert.equal(await page.locator('#prompt').inputValue(), draft, 'copy and fallback preserve the unsent draft');
    const afterCopy = await statusUntil(() => true);
    assert.equal(afterCopy.requests, 3, 'copy does not send a model request');
    assert.equal(afterCopy.executions, 1, 'copy does not execute tools');
    await page.locator('#prompt').fill('');
    await page.locator('.pane-tab[data-pane="tools"]').click();
    assert.deepEqual(await page.evaluate(() => window.gradeSpeech.utterances), [],
      'ordinary rendering and copying do not start speech with narration off');
    await terminal.locator('[data-action="speak"]').click();
    const baselineSpeech = await page.evaluate(() => {
      window.gradeSpeech.drain(); return [...window.gradeSpeech.utterances];
    });
    assert(baselineSpeech.length > 0, 'manual speaker actually reaches browser speech engine');
    await page.evaluate(() => { window.gradeSpeech.utterances = []; });
    await terminal.locator('[data-action="speak"]').click();
    const activeSpeech = await page.evaluate(() => ({
      text: window.gradeSpeech.current?.text, cancellations: window.gradeSpeech.cancellations
    }));
    assert(activeSpeech.text, 'speech is active during reader operations');
    await terminal.getByRole('button', {name: 'View call', exact: true}).click();
    await page.locator('[data-kind="tool_call"]:focus').getByRole('button',
      {name: 'View latest result', exact: true}).click();
    await page.getByRole('button', {name: 'Chat', exact: true}).click();
    await page.evaluate(() => { window.gradeClipboard.denied = false; });
    await snippets.nth(0).locator('[data-action="copy-code"]').click();
    assert.deepEqual(await page.evaluate(() => ({
      text: window.gradeSpeech.current?.text, cancellations: window.gradeSpeech.cancellations
    })), activeSpeech, 'copy and pair navigation do not cancel or replace active speech');
    const afterSpeech = await page.evaluate(() => {
      window.gradeSpeech.drain(); return [...window.gradeSpeech.utterances];
    });
    assert.deepEqual(afterSpeech, baselineSpeech, 'copy and pair navigation do not duplicate queued speech');
    await page.reload();
    await page.locator('.pane-tab[data-pane="tools"]').click();
    await page.locator(`[data-part-id="${callID}"]`).waitFor();
    assert((await page.locator(`[data-part-id="${resultID}"] .job-state`).innerText()).includes('state when observed: running'),
      'replay preserves running snapshot');
    const replayCallLink = page.locator(`[data-part-id="${resultID}"]`).getByRole('button', {name: 'View call', exact: true});
    assert.equal(await replayCallLink.count(), 1, 'replay preserves the actual result-to-call relationship');
    await replayCallLink.click();
    assert.equal(await page.locator('[data-kind="tool_call"]:focus').getAttribute('data-part-id'), callID,
      'replay result navigation preserves exact original call identity');
    await page.locator(`[data-part-id="${callID}"]`).getByRole('button', {name: 'View latest result', exact: true}).click();
    assert.equal(await page.locator('[data-kind="tool_result"]:focus').getAttribute('data-part-id'), resultID,
      'replay call navigation preserves exact original result identity');
    const replay = await statusUntil(() => true);
    assert.equal(replay.executions, 1, 'replay cannot reexecute the callback');
    assert.equal(replay.requests, 3, 'replay cannot issue model requests');
    assert.deepEqual(replay.observations, completed.observations, 'replay keeps canonical timing');
    await checkRecordedTiming(page.locator(`[data-part-id="${resultID}"]`), observed);
    await checkRecordedTiming(page.locator('[data-kind="tool_result"]').filter({has:
      page.locator('.job-state', {hasText: 'state when observed: finished'})}), completed.observations[1]);
    await page.locator('.pane-tab[data-pane="agents"]').click();
    await page.locator('#agent-tree [data-agent="legacy-timing"]').click();
    await page.locator('.pane-tab[data-pane="tools"]').click();
    async function checkLegacyTiming() {
      const unknown = page.locator('[data-kind="tool_result"]').filter({hasText: 'legacy-unknown-' + initial.nonce});
      const zero = page.locator('[data-kind="tool_result"]').filter({hasText: 'legacy-zero-' + initial.nonce});
      await unknown.waitFor();
      await zero.waitFor();
      assert.equal(await page.locator(`[data-part-id="${callID}"], [data-part-id="${resultID}"]`).count(), 0,
        'selected legacy Agent cannot display previous Agent call or result');
      assert.equal(await page.locator('#tool-transcript').getByRole('button', {name: /^View (call|latest result)$/}).count(), 0,
        'unpaired legacy results cannot inherit previous Agent pair controls');
      assert.equal(await unknown.locator(':scope > time.recorded-time').count(), 0,
        'legacy browser must not invent recorded time');
      assert.equal(await unknown.locator(':scope > .observation-duration').count(), 0,
        'legacy browser must not invent zero duration');
      await checkRecordedTiming(zero, {recorded: '2004-03-02T01:02:03.4567890+05:30', milliseconds: 0});
    }
    await checkLegacyTiming();
    await page.reload();
    await page.locator('.pane-tab[data-pane="agents"]').click();
    await page.locator('#agent-tree [data-agent="legacy-timing"]').click();
    await page.locator('.pane-tab[data-pane="tools"]').click();
    await checkLegacyTiming();
    const afterLegacy = await statusUntil(() => true);
    assert.equal(afterLegacy.requests, 3, 'legacy snapshot selection and reload issue no model requests');
    assert.equal(afterLegacy.executions, 1, 'legacy snapshot selection and reload execute no tools');
    await page.locator('.pane-tab[data-pane="agents"]').click();
    await page.locator('#agent-tree [data-agent="snippet-stream"]').click();
    await page.getByRole('button', {name: 'Chat', exact: true}).click();
    await page.evaluate(() => { window.gradeSpeech.automaticCompletion = true; });
    await page.locator('.speech-toggle').click();
    assert(await page.locator('#narrate').isChecked(), 'automatic narration enabled through visible control');
    await page.locator('#prompt').fill('stream the snippet ' + initial.nonce);
    await page.locator('#send').click();
    const streamed = page.locator('#transcript [data-role="assistant"][data-kind="text"]').filter({hasText: 'stream-' + initial.nonce});
    await streamed.waitFor();
    assert.equal(await streamed.locator('[data-action="copy-code"]').count(), 0,
      'unfinished actual provider stream has no copy control');
    await page.request.post(origin + '/grade/stream-step/1');
    const streamCopy = streamed.locator('[data-action="copy-code"]');
    await streamCopy.waitFor();
    await streamed.locator('summary').click();
    await streamCopy.click();
    const streamPayload = ('  streaming 界 ' + initial.nonce + '\n').repeat(40);
    assert.equal(await page.evaluate(() => window.gradeClipboard.writes.at(-1)), streamPayload);
    await streamCopy.focus();
    const copyNode = await streamCopy.elementHandle();
    await page.request.post(origin + '/grade/stream-step/2');
    await streamed.getByText('later-' + initial.nonce, {exact: true}).waitFor();
    assert(await copyNode.evaluate(node => node.isConnected && document.activeElement === node),
      'real streaming preserves actual focused copy button');
    assert.equal(await streamed.locator('details').getAttribute('open'), '', 'real streaming preserves expanded snippet');
    assert.equal(await streamed.locator('[role="status"]').innerText(), 'Copied', 'real streaming preserves copy feedback');
    await copyNode.dispose();
    await page.evaluate(() => Object.defineProperty(navigator, 'clipboard', {configurable: true, value: undefined}));
    await streamCopy.click();
    const streamManual = streamed.locator('textarea');
    assert.equal(await streamManual.inputValue(), streamPayload, 'missing API gives exact manual payload');
    const manualNode = await streamManual.elementHandle();
    await page.request.post(origin + '/grade/stream-step/3');
    await streamed.getByText('finished-' + initial.nonce, {exact: true}).waitFor();
    assert(await manualNode.evaluate(node => node.isConnected && node.readOnly && document.activeElement === node &&
      node.selectionStart === 0 && node.selectionEnd === node.value.length),
      'real streaming preserves actual manual selection and focus');
    await manualNode.dispose();
    const streamDone = await statusUntil(state => state.streamCompleted === 1);
    assert.equal(streamDone.streamRequests, 1, 'streaming controls issue no extra provider request');
    assert.equal(streamDone.streamFailure, null, 'streaming provider finishes cleanly');
    assert.equal(streamDone.requests, 3);
    assert.equal(streamDone.executions, 1);
    await page.waitForFunction(marker => window.gradeSpeech.utterances.some(text => text.includes(marker)),
      'finished-' + initial.nonce);
    const narratedStream = await page.evaluate(() => [...window.gradeSpeech.utterances]);
    const spokenStream = narratedStream.join(' ');
    for (const marker of ['stream-', 'closed-', 'later-', 'finished-']) {
      assert.equal(spokenStream.split(marker + initial.nonce).length - 1, 1,
        'real streaming automatically narrates each prose marker once: ' + marker);
    }
    assert(!spokenStream.includes('streaming 界'), 'automatic narration does not read fenced code payload');
    for (const deny of [false, true]) {
      await page.evaluate(() => {
        window.gradeClipboard.pending = null;
        Object.defineProperty(navigator, 'clipboard', {configurable: true, value: {
          writeText: text => {
            window.gradeClipboard.writes.push(text);
            return new Promise((resolve, reject) => { window.gradeClipboard.pending = {resolve, reject}; });
          }
        }});
      });
      const writesBefore = await page.evaluate(() => window.gradeClipboard.writes.length);
      const oldSnippet = await streamed.locator('.code-snippet').elementHandle();
      await streamCopy.click();
      assert.equal(await streamed.locator('[role="status"]').innerText(), 'Copying…');
      await page.locator('.pane-tab[data-pane="agents"]').click();
      await page.locator('#agent-tree [data-agent="legacy-timing"]').click();
      await page.locator('#agent-tree [data-agent="snippet-stream"]').click();
      await page.getByRole('button', {name: 'Chat', exact: true}).click();
      await streamed.waitFor();
      assert(await oldSnippet.evaluate(node => !node.isConnected), 'Agent switch detaches original snippet');
      const freshStatus = await streamed.locator('[role="status"]').innerText();
      assert.equal(freshStatus, '', 'replayed snippet starts without old copy feedback');
      await page.evaluate(async deny => {
        const pending = window.gradeClipboard.pending;
        if (deny) pending.reject(new DOMException('Denied', 'NotAllowedError'));
        else pending.resolve();
        await new Promise(resolve => setTimeout(resolve, 0));
      }, deny);
      assert.equal(await streamed.locator('[role="status"]').innerText(), freshStatus,
        'delayed clipboard completion cannot change replayed feedback');
      assert.equal(await streamed.locator('textarea').count(), 0,
        'delayed clipboard denial cannot create replacement fallback');
      assert.equal(await oldSnippet.evaluate(node => node.querySelector('[role="status"]').textContent), 'Copying…',
        'detached clipboard completion does not update obsolete feedback');
      assert.equal(await page.evaluate(() => window.gradeClipboard.writes.length), writesBefore + 1,
        'delayed clipboard completion cannot recopy');
      assert.equal(await page.evaluate(() => window.gradeClipboard.writes.at(-1)), streamPayload,
        'delayed clipboard owns the original payload');
      await oldSnippet.dispose();
    }
    const afterClipboard = await statusUntil(() => true);
    assert.equal(afterClipboard.streamRequests, 1, 'clipboard replay issues no provider request');
    assert.equal(afterClipboard.executions, 1, 'clipboard replay executes no tool');
    assert.deepEqual(await page.evaluate(() => window.gradeSpeech.utterances), narratedStream,
      'selection replay and clipboard completion do not repeat automatic narration');
    await page.locator('.speech-toggle').click();
    assert(!(await page.locator('#narrate').isChecked()), 'narration explicitly disabled after stream check');
    await page.locator('.pane-tab[data-pane="agents"]').click();
    await page.locator('#agent-tree [data-agent="dispatch-states"]').click();
    await page.getByRole('button', {name: 'Chat', exact: true}).click();
    await page.locator('#prompt').fill('Run the two state steps.');
    await page.locator('#send').click();
    await page.getByRole('button', {name: 'Tools', exact: true}).click();
    const stateCalls = page.locator('#tool-transcript [data-kind="tool_call"]');
    await stateCalls.nth(1).waitFor();
    await statusUntil(state => state.stateExecutions === 1);
    assert.equal(await stateCalls.count(), 2, 'both finalized requests are visible');
    assert.equal(await stateCalls.nth(0).locator('.tool-state').innerText(), 'Dispatched · awaiting result',
      'first real batch call is dispatched, not merely requested or complete');
    assert.equal(await stateCalls.nth(1).locator('.tool-state').innerText(), 'Requested',
      'undispatched second batch call is only requested');
    assert.equal(await page.locator('#tool-transcript [data-kind="tool_result"]').count(), 0,
      'dispatch is not a returned observation');
    const stateIDs = await stateCalls.evaluateAll(nodes => nodes.map(node => node.dataset.partId));
    const stateRelease = await page.request.post(origin + '/grade/release-state');
    assert.equal(stateRelease.status(), 200, 'canonical held batch boundary agrees with visible states');
    await page.getByRole('button', {name: 'Chat', exact: true}).click();
    await page.getByText('states-complete-' + initial.nonce, {exact: true}).waitFor();
    await page.getByRole('button', {name: 'Tools', exact: true}).click();
    assert.deepEqual(await stateCalls.evaluateAll(nodes => nodes.map(node => node.dataset.partId)), stateIDs,
      'dispatch and results preserve original call identities');
    for (const call of await stateCalls.all())
      assert.equal(await call.locator('.tool-state').innerText(), 'Result returned');
    assert.equal(await page.locator('#tool-transcript [data-kind="tool_result"]').count(), 2);
    const statesDone = await statusUntil(state => state.stateRequests === 2);
    assert.equal(statesDone.stateExecutions, 2);
    assert.equal(statesDone.stateFailure, null);
    assert.deepEqual(errors, []);
    console.log(JSON.stringify({contract:'artifact-actions-v1', passed:true, checks:[
      'real-running-observation', 'call-result-navigation', 'historical-running-preserved',
      'terminal-pathname-not-grant', 'replay-no-execution', 'replay-exact-pair-and-agent-isolation', 'focused-expanded-reader-preserved', 'canonical-recorded-time-and-duration', 'timing-survives-replay', 'legacy-unknown-and-zero-browser', 'provider-snippets-exact-and-fallback', 'real-streaming-snippet-stability', 'delayed-clipboard-replay-fenced', 'raw-selection-and-scroll-preserved', 'job-output-owner-and-running-guards', 'requested-and-dispatched-batch-states', 'keyboard-touch-and-draft-preserved', 'speech-copy-and-pair-noninterference', 'streamed-narration-once-and-replay-silent'
    ]}));
  } finally { await browser.close(); }
})().catch(error => { console.error(error.stack); process.exitCode = 1; });
