import {Page} from './page.js';
const page = new Page(document.querySelector('main'));
window.addEventListener('pagehide', () => page.close());
