// Actual pixels from this tab only. The caller owns gesture, routing and draft
// admission; this object owns the picker result, video, tracks and canvas.
export class PageCapture {
    #devices;
    #active = null;
    constructor(mediaDevices = navigator.mediaDevices) { this.#devices = mediaDevices; }
    get busy() { return this.#active !== null; }
    get supported() {
        return isSecureContext && typeof this.#devices?.getDisplayMedia === "function"
            && typeof this.#devices?.setCaptureHandleConfig === "function"
            && typeof MediaStreamTrack.prototype.getCaptureHandle === "function"
            && typeof HTMLVideoElement.prototype.requestVideoFrameCallback === "function";
    }
    capture() {
        if (this.busy) return Promise.reject(new Error("A page capture is still closing."));
        if (!this.supported) return Promise.reject(new Error(
            "This browser cannot verify capture of this page. Take a screenshot and attach the image instead."));
        const handle = crypto.randomUUID();
        let wake;
        const state = { handle, error: null, tracks: [], video: null, canvas: null,
            unwait: null, done: null, stopped: new Promise(resolve => { wake = resolve; }), wake: () => wake() };
        this.#active = state;
        let request;
        try {
            this.#devices.setCaptureHandleConfig({ handle, exposeOrigin: true, permittedOrigins: [location.origin] });
            // Invoke synchronously in the user's gesture, without an intervening await.
            // Browser hints do not establish identity; the returned handle does.
            request = this.#devices.getDisplayMedia({ audio: false,
                video: { displaySurface: "browser" }, preferCurrentTab: true,
                selfBrowserSurface: "include", surfaceSwitching: "exclude" });
        } catch (error) { request = Promise.reject(error); }
        state.done = this.#run(state, request);
        return state.done;
    }
    cancel() {
        const state = this.#active;
        if (!state) return Promise.resolve();
        this.#invalidate(state, new DOMException("Page capture cancelled.", "AbortError"));
        // A browser picker and canvas encoding are not abortable APIs. Join their
        // eventual completion; any late stream is stopped and no image is returned.
        return state.done.then(() => {}, () => {});
    }
    #invalidate(state, error) {
        state.error ??= error;
        state.wake();
        state.unwait?.();
        for (const track of state.tracks) track.stop();
    }
    #check(state, track) {
        if (state.error) throw state.error;
        if (track.readyState !== "live") throw new Error("Captured tab stopped before the image was ready.");
        const identity = track.getCaptureHandle?.();
        if (track.getSettings().displaySurface !== "browser"
            || identity?.handle !== state.handle || identity?.origin !== location.origin)
            throw new Error("Capture refused: select this Ensemble tab, not another tab, window or screen.");
    }
    async #run(state, request) {
        let track, changed, ended, playing, framed;
        try {
            const stream = await request;
            state.tracks = Array.from(stream.getTracks()); // Own original membership.
            if (state.error) throw state.error;
            const videoTracks = state.tracks.filter(item => item.kind === "video");
            if (state.tracks.length !== 1 || videoTracks.length !== 1)
                throw new Error("Capture requires exactly one video track and no audio.");
            track = videoTracks[0];
            this.#check(state, track);
            changed = () => {
                try { this.#check(state, track); }
                catch (error) { this.#invalidate(state, error); }
            };
            ended = () => this.#invalidate(state, new Error("Captured tab stopped before the image was ready."));
            track.addEventListener("capturehandlechange", changed);
            track.addEventListener("ended", ended);
            const video = document.createElement("video");
            state.video = video;
            video.muted = true;
            video.playsInline = true;
            video.srcObject = new MediaStream(videoTracks);
            // Register before play: a static tab may deliver only its first frame.
            framed = new Promise((resolve, reject) => {
                let frame;
                const finish = error => {
                    clearTimeout(timer);
                    video.cancelVideoFrameCallback(frame);
                    state.unwait = null;
                    if (error) reject(error); else resolve();
                };
                const timer = setTimeout(() => finish(new Error("No captured video frame arrived.")), 5000);
                frame = video.requestVideoFrameCallback(() => finish());
                state.unwait = () => finish();
            });
            playing = video.play();
            await Promise.race([Promise.all([playing, framed]), state.stopped]);
            this.#check(state, track);
            const width = video.videoWidth, height = video.videoHeight;
            if (!Number.isInteger(width) || !Number.isInteger(height) || width <= 0 || height <= 0
                || width * height > 24_000_000)
                throw new Error("Captured image exceeds the 24-million-pixel limit.");
            const canvas = document.createElement("canvas");
            state.canvas = canvas;
            canvas.width = width;
            canvas.height = height;
            const context = canvas.getContext("2d");
            if (!context) throw new Error("Image capture is unavailable.");
            context.drawImage(video, 0, 0);
            const blob = await new Promise(resolve => canvas.toBlob(resolve, "image/png"));
            this.#check(state, track);
            if (!blob || !blob.size || blob.size > 32 * 1024 * 1024)
                throw new Error("Captured PNG is empty or exceeds 32 MiB.");
            return { file: new File([blob], "ensemble-page.png", { type: "image/png" }), width, height };
        } finally {
            if (track) {
                track.removeEventListener("capturehandlechange", changed);
                track.removeEventListener("ended", ended);
            }
            state.unwait?.();
            for (const owned of state.tracks) owned.stop();
            if (state.video) { state.video.pause(); state.video.srcObject = null; }
            await Promise.allSettled([playing, framed].filter(Boolean));
            if (state.canvas) { state.canvas.width = 0; state.canvas.height = 0; }
            if (this.#active === state) this.#active = null;
        }
    }
}
