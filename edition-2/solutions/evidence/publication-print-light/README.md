# Black-on-white print code

Bill requested a light print theme for black ink on white paper. The five-line
CSS override changes only foreground and background colors for code and its
syntax spans in the print-specific header. Existing font emphasis, spacing,
wrapping and pagination remain. EPUB styling is separate and unchanged.

`check.py` is the one-off build/observation receipt, not a new product test suite.
It rebuilds only print HTML and the two PDFs, then uses a dedicated real Chrome
process with print media emulation. All 13,993 selected code/container elements
have computed black text and white backgrounds. Recorded string, comment,
keyword and operator samples preserve bold/italic styling. Removing only the
new rule in the browser DOM reveals the previous colored tokens and verifies
that every observed element retains its font weight/style and geometry. This
DOM comparison does not modify the source or saved artifacts.

`check.json` binds the actual commands, output hashes and all protected manuscript,
assembled Markdown and EPUB hashes. They are unchanged from the starting bytes;
Bill's user-owned preface was never edited. All 30 TOC heading pages match the
prior `11f5562a` / publication-preface-r2 map. Page counts remain 371 interior and
372 including the cover, at 6 by 9 inches. The coder visually inspected actual
PDF page 361, retained as `page-361.png`: code is black on white, bold type remains
visible, and the table/footer retain their layout. Independent review follows.

This is local print preparation, not a physical proof copy or permission to
publish. No EPUB rebuild, provider call, runtime test, push or upload occurred.
Earlier publication artifacts and receipts remain in Git history and their
original evidence directories.
