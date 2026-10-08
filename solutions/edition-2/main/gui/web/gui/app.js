import {BrowserApplication} from './application.js';
const application = new BrowserApplication();
application.createPage(document.querySelector('main'));
window.addEventListener('pagehide', () => application.close(), {once: true});
