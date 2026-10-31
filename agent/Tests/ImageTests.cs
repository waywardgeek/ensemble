using Ensemble;
using SkiaSharp;
using Application = Ensemble.Ensemble;

internal static class ImageTests
{
    private static void Check(bool condition, string message)
    {
        if (!condition) throw new Exception(message);
    }

    private static byte[] Png(int width, int height)
    {
        using var bitmap = new SKBitmap(width, height, SKColorType.Rgba8888, SKAlphaType.Premul);
        bitmap.Erase(SKColors.Transparent);
        bitmap.SetPixel(0, 0, SKColors.Red);
        using var image = SKImage.FromBitmap(bitmap);
        using var data = image.Encode(SKEncodedImageFormat.Png, 100);
        return data.ToArray();
    }

    private static SKBitmap Decode(PreparedImage prepared)
    {
        string locator = prepared.Image.Ref.Locator;
        Check(prepared.Image.Mime == "image/png" && locator.StartsWith("data:image/png;base64,"),
            "Prepared image must carry real PNG bytes.");
        byte[] bytes = Convert.FromBase64String(locator[(locator.IndexOf(',') + 1)..]);
        Check(bytes.Length == prepared.OutputBytes && bytes.Length <= Engine.MaximumImageBytes,
            "Output byte accounting or delivery bound was wrong.");
        return SKBitmap.Decode(bytes) ?? throw new Exception("Output pixels could not be decoded.");
    }

    private static void Refuses(Action action, string message)
    {
        bool refused = false;
        try { action(); }
        catch (Exception error) when (error is InvalidDataException or IOException
            or UnauthorizedAccessException or ArgumentException) { refused = true; }
        Check(refused, message);
    }

    private static void CheckDelivery(Application app, PreparedImage prepared)
    {
        string expected = prepared.Image.Ref.Locator.Split(',', 2)[1];
        foreach (var sample in new[]
        {
            (Vendor.OpenAI, "gpt-6-astra", Surface.Responses),
            (Vendor.Anthropic, "claude-3-haiku-20240307", Surface.Messages),
            (Vendor.Gemini, "gemini-2.5-flash", Surface.GenerateContent)
        })
        {
            var agent = new Agent(app, new Configuration(sample.Item1, sample.Item2,
                "http://127.0.0.1:1", surface: sample.Item3));
            try
            {
                agent.History().Append(new Event(EventKind.MessageReceived, actor: Actor.Human,
                    parts: [new TextPart("before"), prepared.Image, new TextPart("after")]));
                string before = JsonCodec.WriteEvent(agent.History().Events()[0]);
                var wire = agent.Engine().Render(agent.History().View());
                string actual = sample.Item1 switch
                {
                    Vendor.OpenAI => wire["input"]![0]!["content"]![1]!["image_url"]!
                        .GetValue<string>().Split(',', 2)[1],
                    Vendor.Anthropic => wire["messages"]![0]!["content"]![1]!["source"]!["data"]!.GetValue<string>(),
                    _ => wire["contents"]![0]!["parts"]![1]!["inlineData"]!["data"]!.GetValue<string>()
                };
                Check(actual == expected, $"{sample.Item1} did not deliver the prepared bytes.");
                Check(JsonCodec.WriteEvent(agent.History().Events()[0]) == before,
                    "Rendering modified the accepted image.");
            }
            finally { agent.Destroy(); }
        }
    }

    private static void CheckOutputBound(Agent agent)
    {
        using var bitmap = new SKBitmap(2048, 2048, SKColorType.Rgba8888, SKAlphaType.Unpremul);
        var random = new Random(26);
        Span<byte> pixels = bitmap.GetPixelSpan();
        random.NextBytes(pixels);
        for (int i = 3; i < pixels.Length; i += 4) pixels[i] = 255;
        using var image = SKImage.FromBitmap(bitmap);
        using var encoded = image.Encode(SKEncodedImageFormat.Png, 100);
        byte[] source = encoded.ToArray();
        Check(source.Length > Agent.MaximumImageOutputBytes, "Fixture did not exceed delivery bytes.");
        var prepared = agent.PrepareImage(source.AsMemory());
        Check(prepared.Resized && prepared.Width < 2048 && prepared.Width == prepared.Height,
            "Oversized encoded output did not trigger proportional reduction.");
        using var decoded = Decode(prepared);
        Check(decoded.Width == prepared.Width, "Bounded output metadata did not match pixels.");
    }

    private static void CheckOrientations(Agent agent)
    {
        using var bitmap = new SKBitmap(32, 16);
        using (var canvas = new SKCanvas(bitmap))
        {
            SKColor[] colors = [SKColors.Red, SKColors.Lime, SKColors.Blue, SKColors.White];
            for (int corner = 0; corner < 4; corner++)
            {
                using var paint = new SKPaint { Color = colors[corner] };
                canvas.DrawRect((corner % 2) * 16, (corner / 2) * 8, 16, 8, paint);
            }
        }
        using var image = SKImage.FromBitmap(bitmap);
        using var encoded = image.Encode(SKEncodedImageFormat.Jpeg, 100);
        byte[] jpeg = encoded.ToArray();
        using var original = SKBitmap.Decode(jpeg);
        SKColor[] expectedColors = [original.GetPixel(2, 2), original.GetPixel(29, 2),
            original.GetPixel(2, 13), original.GetPixel(29, 13)];
        int[][] mappings = [[0,1,2,3], [1,0,3,2], [3,2,1,0], [2,3,0,1],
            [0,2,1,3], [2,0,3,1], [3,1,2,0], [1,3,0,2]];
        for (int orientation = 1; orientation <= 8; orientation++)
        {
            // A minimal EXIF TIFF IFD with the orientation tag, no private metadata.
            byte[] exif = Convert.FromHexString(
                "FFE1002245786966000049492A0008000000010012010300010000000100000000000000");
            exif[28] = (byte)orientation;
            byte[] source = [.. jpeg.AsSpan(0, 2), .. exif, .. jpeg.AsSpan(2)];
            var prepared = agent.PrepareImage(source.AsMemory());
            bool swap = orientation >= 5;
            Check(prepared.Width == (swap ? 16 : 32) && prepared.Height == (swap ? 32 : 16)
                && !prepared.Resized, $"Orientation {orientation} dimensions were wrong.");
            using var upright = Decode(prepared);
            SKColor[] actual = [upright.GetPixel(2, 2), upright.GetPixel(upright.Width - 3, 2),
                upright.GetPixel(2, upright.Height - 3), upright.GetPixel(upright.Width - 3, upright.Height - 3)];
            for (int corner = 0; corner < 4; corner++)
            {
                Check(actual[corner] == expectedColors[mappings[orientation - 1][corner]],
                    $"Orientation {orientation} corner {corner} was wrong.");
            }
        }
    }

    public static void Preparation(string temp)
    {
        string root = Path.Combine(temp, "image-preparation");
        string inside = Path.Combine(root, "inside");
        Directory.CreateDirectory(inside);
        var app = new Application(TextWriter.Null);
        try
        {
            var config = new Configuration(Vendor.OpenAI, "gpt-6-astra", "http://127.0.0.1:1",
                surface: Surface.Responses, sandboxRoot: inside);
            var agent = new Agent(app, config, Path.Combine(inside, "jobs"));
            byte[] small = Png(7, 5);
            var same = agent.PrepareImage(small.AsMemory());
            Check(!same.Resized && same.Width == 7 && same.Height == 5, "Small image was enlarged.");
            using (var pixels = Decode(same))
            {
                Check(pixels.GetPixel(0, 0) == SKColors.Red, "Image pixel content was lost.");
                Check(pixels.GetPixel(6, 4).Alpha == 0, "Transparency was lost.");
            }
            byte[] large = Png(4096, 2048);
            string path = Path.Combine(inside, "large.png");
            File.WriteAllBytes(path, large);
            var resized = agent.PrepareImage("large.png");
            Check(resized.Resized && resized.OriginalWidth == 4096 && resized.OriginalHeight == 2048
                && resized.Width == 2048 && resized.Height == 1024, "Aspect-preserving dimensions were wrong.");
            using (var pixels = Decode(resized))
            {
                Check(pixels.Width == resized.Width && pixels.Height == resized.Height,
                    "Reported resize was not represented in the output pixels.");
            }
            Check(File.ReadAllBytes(path).SequenceEqual(large), "Preparation overwrote the source file.");
            File.WriteAllBytes(Path.Combine(root, "outside.png"), small);
            File.CreateSymbolicLink(Path.Combine(inside, "escape.png"), Path.Combine(root, "outside.png"));
            Refuses(() => agent.PrepareImage("escape.png"), "Outward symlink was accepted.");
            Refuses(() => agent.PrepareImage("../outside.png"), "Outside file was accepted.");
            Refuses(() => agent.PrepareImage(ReadOnlyMemory<byte>.Empty), "Empty image accepted.");
            Refuses(() => agent.PrepareImage("not an image"u8.ToArray().AsMemory()), "Nonimage accepted.");
            Refuses(() => agent.PrepareImage(small.AsMemory(0, small.Length / 2)), "Truncated image accepted.");
            Refuses(() => agent.PrepareImage(new byte[Agent.MaximumImageSourceBytes + 1].AsMemory()),
                "Source byte limit ignored.");
            using (var oversized = File.Create(Path.Combine(inside, "oversized.png")))
            {
                oversized.SetLength(Agent.MaximumImageSourceBytes + 1L);
            }
            Refuses(() => agent.PrepareImage("oversized.png"), "Oversized file accepted.");
            // A highly compressible real image still exceeds the decoded-pixel ceiling.
            byte[] tooManyPixels = Png(5000, 5000);
            Refuses(() => agent.PrepareImage(tooManyPixels.AsMemory()), "Decoded-pixel limit ignored.");
            byte[] animation = Convert.FromHexString(
                "47494638396101000100800000FF0000000000" +
                "21F90400000000002C0000000001000100000202440100" +
                "21F90400000000002C00000000010001000002024401003B");
            Refuses(() => agent.PrepareImage(animation.AsMemory()), "Animated image accepted.");
            CheckOrientations(agent);
            CheckOutputBound(agent);
            File.Delete(path);
            CheckDelivery(app, resized);
            agent.Destroy();
        }
        finally { app.Destroy(); }
    }
}
