"use strict";
// Transient layout only: the server still owns Agent state and preferences.
const workspace = document.querySelector("#workspace");
const minimumPanePixels = 160; // Enough to keep the composer and pane controls reachable.
const paneFor = side => document.querySelector(side === "left" ? "#left-pane" : "#tools-pane");
function paneLimits(side) {
    const width = workspace.getBoundingClientRect().width;
    const gutters = ["left", "right"].reduce((sum, name) =>
        sum + document.querySelector(`#${name}-splitter`).getBoundingClientRect().width, 0);
    const other = paneFor(side === "left" ? "right" : "left").getBoundingClientRect().width;
    return { minimum: minimumPanePixels / width * 100,
        maximum: Math.max(minimumPanePixels, width - gutters - other - minimumPanePixels) / width * 100 };
}
function panePercent(side) {
    return paneFor(side).getBoundingClientRect().width / workspace.getBoundingClientRect().width * 100;
}
function resizePane(side, value) {
    const limits = paneLimits(side);
    const bounded = Math.max(limits.minimum, Math.min(limits.maximum, value));
    // Avoid observer feedback from writing an unchanged, rounded geometry value.
    if (Math.abs(bounded - panePercent(side)) > .01)
        document.documentElement.style.setProperty(`--${side}-width`, `${bounded}%`);
}
function describeSplitters() {
    for (const side of ["left", "right"]) {
        const splitter = document.querySelector(`#${side}-splitter`);
        const limits = paneLimits(side);
        splitter.setAttribute("aria-valuemin", limits.minimum.toFixed(1));
        splitter.setAttribute("aria-valuemax", limits.maximum.toFixed(1));
        splitter.setAttribute("aria-valuenow", panePercent(side).toFixed(1));
        splitter.setAttribute("aria-valuetext", `${Math.round(paneFor(side).getBoundingClientRect().width)} pixels`);
    }
}
function fitPanes() {
    const em = parseFloat(getComputedStyle(document.documentElement).fontSize);
    const compact = window.innerWidth < 62 * em;
    document.body.classList.toggle("compact", compact);
    if (!compact) {
        // Re-clamp after a window/font change, not just after dragging. Each side
        // leaves room for the other side and the conversation; no fixed 35% cap.
        resizePane("left", panePercent("left"));
        resizePane("right", panePercent("right"));
        describeSplitters();
    }
}
new ResizeObserver(fitPanes).observe(document.documentElement);
new MutationObserver(fitPanes).observe(document.documentElement, { attributes: true, attributeFilter: ["style"] });
fitPanes();
for (const button of document.querySelectorAll("button[data-pane]")) {
    button.onclick = () => {
        workspace.dataset.pane = button.dataset.pane;
        for (const other of document.querySelectorAll("button[data-pane]"))
            other.setAttribute("aria-pressed", String(other === button));
    };
}
for (const side of ["left", "right"]) {
    const splitter = document.querySelector(`#${side}-splitter`);
    splitter.onkeydown = event => {
        if (!["ArrowLeft", "ArrowRight", "Home", "End"].includes(event.key)) return;
        event.preventDefault();
        const limits = paneLimits(side), direction = side === "left" ? 1 : -1;
        resizePane(side, event.key === "Home" ? limits.minimum : event.key === "End" ? limits.maximum
            : panePercent(side) + (event.key === "ArrowRight" ? 2 : -2) * direction);
        describeSplitters();
    };
    splitter.onpointerdown = event => {
        if (event.button === 0) splitter.setPointerCapture(event.pointerId);
    };
    splitter.onpointermove = event => {
        if (!splitter.hasPointerCapture(event.pointerId)) return;
        const bounds = workspace.getBoundingClientRect();
        const distance = side === "left" ? event.clientX - bounds.left : bounds.right - event.clientX;
        resizePane(side, distance / bounds.width * 100);
        describeSplitters();
    };
    splitter.onpointerup = event => {
        if (splitter.hasPointerCapture(event.pointerId)) splitter.releasePointerCapture(event.pointerId);
    };
}
