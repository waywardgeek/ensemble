using System.Threading.Channels;
using Ensemble.Gateway;

namespace GatewayInterview;

/// <summary>Value-like application options, not another connection or policy store.</summary>
public sealed record InterviewOptions(ClientConfig Connection, string Subject,
    string[] ReplyGroups, string ReportPath, TimeSpan Timeout);

/// <summary>Application root for one interview. Owns only the bounded answer inbox.
/// The public Client borrows this callback boundary while RunAsync owns its lifetime;
/// neither object owns the other's model or an Ensemble actor.</summary>
public sealed partial class Interview : IHandler
{
    private readonly Channel<Envelope> answers;

    /// <summary>Create a fresh, single-use interview inbox.</summary>
    public partial Interview();

    /// <summary>Ask four fixed questions; publish a report only after four correlated
    /// answers. Handler acceptance is inbox admission, not report durability.</summary>
    public partial Task RunAsync(InterviewOptions options, CancellationToken cancellation = default);
    public partial void OnConnected(WelcomePayload welcome);
    public partial void OnDeliver(Envelope envelope);
    public partial void OnDisconnected(string reason);
    public partial void OnError(Exception error);
}
