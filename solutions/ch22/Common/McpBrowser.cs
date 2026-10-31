namespace Ensemble;

// A host's one connected browser target. Mcp owns the registration; after a
// skill claims it, both this registration and the ordinary manager/skill own
// the client. Disconnect cascades to that client, never to the Agent or skill.
//
// A registration serves one protocol session. Closing a skill closes its I/O;
// a fresh connection requires a fresh registration, not reuse of old RPC IDs.
public sealed partial class McpBrowser
{
    internal Mcp? parent;
    internal bool destroyed;
    private readonly IMcpTransport transport;
    private readonly object gate = new();
    private McpClient? client;
    private bool claimed;
    private Task? destruction;

    // On successful construction, takes ownership of the connected transport.
    // A refused registration leaves transport cleanup with the caller.
    public partial McpBrowser(Mcp parent, IMcpTransport transport);
    public partial Mcp Parent();
    public partial McpClient? Client();
    internal partial McpClient Claim(Skills owner, string skill, string server);
    public partial void AddClient(McpClient child);
    public partial void RemoveClient(McpClient child);
    // Joinable cleanup closes even an unused transport and detaches both ends.
    public partial Task DestroyAsync();
    private partial Task DestroyCoreAsync();
}
