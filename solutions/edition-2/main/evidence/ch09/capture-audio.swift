// Short, explicit Chrome-only system audio receipt. No microphone or video output.
import Foundation
import ScreenCaptureKit
import AVFoundation
import CoreMedia

final class Recorder: NSObject, SCStreamOutput {
    let path: URL
    var file: AVAudioFile?
    var frames: UInt64 = 0
    init(path: String) { self.path = URL(fileURLWithPath: path) }
    func stream(_ stream: SCStream, didOutputSampleBuffer sample: CMSampleBuffer, of type: SCStreamOutputType) {
        guard type == .audio, let description = CMSampleBufferGetFormatDescription(sample), let format = AVAudioFormat(cmAudioFormatDescription: description) as AVAudioFormat? else { return }
        let count = CMSampleBufferGetNumSamples(sample)
        guard let buffer = AVAudioPCMBuffer(pcmFormat: format, frameCapacity: AVAudioFrameCount(count)) else { return }
        buffer.frameLength = AVAudioFrameCount(count)
        let status = CMSampleBufferCopyPCMDataIntoAudioBufferList(sample, at: 0, frameCount: Int32(count), into: buffer.mutableAudioBufferList)
        guard status == noErr else { return }
        do {
            if file == nil { file = try AVAudioFile(forWriting: path, settings: format.settings) }
            try file?.write(from: buffer)
            frames += UInt64(count)
        } catch { print("audio_write_failed"); }
    }
}
@main struct Main {
    static func main() async {
        do {
            let content = try await SCShareableContent.excludingDesktopWindows(false, onScreenWindowsOnly: true)
            guard let display = content.displays.first else { throw NSError(domain:"no_display",code:1) }
            let targetPID = Int32(CommandLine.arguments[2])!
            let apps = content.applications.filter { $0.processID == targetPID }
            guard !apps.isEmpty else { throw NSError(domain:"no_chrome_application",code:2) }
            let filter = SCContentFilter(display: display, including: apps, exceptingWindows: [])
            let config = SCStreamConfiguration()
            config.width = 2; config.height = 2
            config.capturesAudio = true; config.excludesCurrentProcessAudio = true
            config.sampleRate = 48000; config.channelCount = 2
            let recorder = Recorder(path: CommandLine.arguments[1])
            let queue = DispatchQueue(label:"ensemble.audio.receipt")
            let stream = SCStream(filter:filter, configuration:config, delegate:nil)
            try stream.addStreamOutput(recorder, type:.audio, sampleHandlerQueue:queue)
            try await stream.startCapture()
            print("capture_ready pid=\(targetPID) app=\(apps[0].applicationName) bundle=\(apps[0].bundleIdentifier) unix_ms=\(Date().timeIntervalSince1970 * 1000)"); fflush(stdout)
            try await Task.sleep(nanoseconds: 8_000_000_000)
            try await stream.stopCapture()
            queue.sync { print("captured_frames=\(recorder.frames)") }
        } catch { print("capture_failed: \(error.localizedDescription)"); exit(1) }
    }
}
