import {Page} from './page.js';
import {SpeechService} from './speech-service.js';

// One explicit owner per document, including embeddings with several Pages.
export class BrowserApplication {
  constructor(synthesis = globalThis.speechSynthesis, Utterance = globalThis.SpeechSynthesisUtterance) {
    this.synthesis = synthesis; this.Utterance = Utterance;
    this.pages = new Set(); this.service = new SpeechService(this); this.closed = false;
  }
  createPage(root, url) {
    if (this.closed) throw new Error('Browser application is closed');
    const page = new Page(this, root, url); this.pages.add(page); return page;
  }
  speech() { return this.service; }
  release(page) { this.pages.delete(page); }
  diagnostic(message) { console.warn('Ensemble browser:', message); }
  close() {
    if (this.closed) return;
    this.closed = true;
    for (const page of this.pages) page.close();
    this.service.close();
  }
}
