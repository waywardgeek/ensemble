using System.Text;
using Ensemble.Gateway;

namespace GatewayInterview;

public sealed partial class Interview
{
    public partial Interview()
    {
        answers = System.Threading.Channels.Channel.CreateBounded<Envelope>(8);
    }

    public partial void OnConnected(WelcomePayload welcome) { }

    public partial void OnDeliver(Envelope envelope)
    {
        // Do not block the sole receive loop: outgoing RPCs also need it.
        // Refusal throws before the Client writes an acknowledgment.
        if (!answers.Writer.TryWrite(envelope))
            throw new IOException("Interview answer inbox is full or closed.");
    }

    public partial void OnDisconnected(string reason)
    {
        answers.Writer.TryComplete(new IOException($"Gateway disconnected: {reason}"));
    }

    public partial void OnError(Exception error)
    {
        answers.Writer.TryComplete(error);
    }

    public partial async Task RunAsync(InterviewOptions options, CancellationToken cancellation)
    {
        if (options.Timeout <= TimeSpan.Zero)
            throw new ArgumentOutOfRangeException(nameof(options), "Question timeout must be positive.");
        await using Client client = await Client.ConnectAsync(options.Connection, this, cancellation);
        using (var setup = CancellationTokenSource.CreateLinkedTokenSource(cancellation))
        {
            setup.CancelAfter(options.Timeout);
            await client.DeclareContactableAsync(options.ReplyGroups, setup.Token);
        }

        string[] questions =
        [
            "What slows you down most in your coding workflow?",
            "What burns the most tokens in your work?",
            "Where do you need a human's help?",
            "What new tool would help you most?"
        ];
        List<string> responses = [];
        foreach (string question in questions)
        {
            using var deadline = CancellationTokenSource.CreateLinkedTokenSource(cancellation);
            deadline.CancelAfter(options.Timeout);
            string conversation = "interview-" + Guid.NewGuid().ToString("N");
            try
            {
                await client.SendAsync(new OutboundMessage(options.Subject, question,
                    ConversationID: conversation, ResponseMode: Protocol.ResponseAsync), deadline.Token);
                while (true)
                {
                    Envelope answer = await answers.Reader.ReadAsync(deadline.Token);
                    // Other authorized messages are deliveries, not answers to this
                    // question. Sender identity comes from the authenticated router.
                    if (answer.ConversationID != conversation || answer.Source.UserID != options.Subject)
                        continue;
                    string header = Protocol.ProvenanceHeader(answer.Source) + "\n\n";
                    if (!answer.Content.StartsWith(header, StringComparison.Ordinal))
                        throw new FormatException("Answer has no matching router provenance header.");
                    responses.Add(answer.Content[header.Length..]);
                    break;
                }
            }
            catch (OperationCanceledException) when (!cancellation.IsCancellationRequested && deadline.IsCancellationRequested)
            {
                throw new TimeoutException($"Question {responses.Count + 1} exceeded its deadline.");
            }
            Console.WriteLine($"Received answer {responses.Count} of {questions.Length}");
        }

        // Opinions belong to the subject. This report is not a benchmark or an
        // independent measurement of the claims made in those answers.
        var report = new StringBuilder("# Ensemble workflow interview\n\n## Friction points\n\n");
        for (int index = 0; index < 3; index++)
            report.Append(questions[index]).Append("\n\n").Append(responses[index]).Append("\n\n");
        report.Append("## Improvement opportunities\n\n").Append(responses[3])
            .Append("\n\n## Raw transcript\n\n");
        for (int index = 0; index < questions.Length; index++)
            report.Append($"### Question {index + 1}\n\n").Append(questions[index])
                .Append("\n\n").Append(responses[index]).Append("\n\n");
        // Create privately on the supported Unix platforms; an existing report's
        // permissions remain its owner's choice. No report is written on timeout.
        var fileOptions = new FileStreamOptions { Mode = FileMode.Create, Access = FileAccess.Write };
        if (!OperatingSystem.IsWindows())
            fileOptions.UnixCreateMode = UnixFileMode.UserRead | UnixFileMode.UserWrite;
        await using var file = new FileStream(options.ReportPath, fileOptions);
        await file.WriteAsync(Encoding.UTF8.GetBytes(report.ToString()), cancellation);
    }
}
