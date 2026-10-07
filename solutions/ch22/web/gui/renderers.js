// Renderers — markdown, ANSI, JSON, and diff rendering for artifacts.

const Renderers = {
  markdown(text) {
    if (typeof marked !== 'undefined') {
      try { return marked.parse(text); } catch(e) { /* fall through */ }
    }
    // Fallback: escape HTML and convert newlines
    return text.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;')
               .replace(/\n/g, '<br>');
  },

  ansi(text) {
    if (typeof AnsiUp !== 'undefined') {
      try {
        const au = new AnsiUp();
        return au.ansi_to_html(text);
      } catch(e) { /* fall through */ }
    }
    return text.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
  },

  json(text) {
    try {
      const obj = JSON.parse(text);
      const pretty = JSON.stringify(obj, null, 2);
      return `<pre>${pretty.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;')}</pre>`;
    } catch(e) {
      // Not valid JSON yet (streaming), show raw
      return `<pre>${text.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;')}</pre>`;
    }
  },

  truncate(text, maxLen) {
    if (text.length <= maxLen) return text;
    return text.substring(0, maxLen) + '…';
  }
};
