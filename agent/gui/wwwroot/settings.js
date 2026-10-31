"use strict";
// This dialog moves no state into a second UI model. The existing controls and
// their acknowledgement/save handlers remain the sole editors of settings.
(() => {
    const dialog = document.querySelector("#settings-dialog");
    const open = document.querySelector("#settings-open");
    const close = document.querySelector("#settings-close");

    open.addEventListener("click", () => {
        if (dialog.open) return;
        dialog.showModal();
        dialog.querySelector('[role="tab"][aria-selected="true"]').focus();
    });
    close.addEventListener("click", () => dialog.close());
    dialog.addEventListener("close", () => open.focus({ preventScroll: true }));
    dialog.addEventListener("shown.bs.tab", () => {
        dialog.querySelector(".settings-body").scrollTop = 0;
    });
})();
