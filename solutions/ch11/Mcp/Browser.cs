namespace Ensemble;

public sealed partial class Mcp
{
    public partial McpBrowser? Browser()
    {
        lock (gate)
        {
            return browser;
        }
    }

    public partial void AddBrowser(McpBrowser child)
    {
        lock (gate)
        {
            if (destroyed || child.destroyed || child.parent is not null || browser is not null)
            {
                throw new InvalidOperationException("MCP already has a browser, or the registration is owned/destroyed.");
            }
            browser = child;
            child.parent = this;
        }
    }

    public partial void RemoveBrowser(McpBrowser child)
    {
        lock (gate)
        {
            if (!ReferenceEquals(browser, child) || !ReferenceEquals(child.parent, this))
            {
                throw new InvalidOperationException("Browser belongs to another MCP manager.");
            }
            browser = null;
            child.parent = null;
        }
    }

    internal partial McpClient ClaimBrowser(Skills owner, string skill, string server)
    {
        var connection = Browser()
            ?? throw new InvalidDataException("This skill requires a connected browser.");
        return connection.Claim(owner, skill, server);
    }
}

public sealed partial class McpBrowser
{
    public partial McpBrowser(Mcp parent, IMcpTransport transport)
    {
        ArgumentNullException.ThrowIfNull(parent);
        ArgumentNullException.ThrowIfNull(transport);
        this.transport = transport;
        parent.AddBrowser(this);
    }

    public partial Mcp Parent() => parent ?? throw new ObjectDisposedException(nameof(McpBrowser));

    public partial McpClient? Client()
    {
        lock (gate)
        {
            return client;
        }
    }

    internal partial McpClient Claim(Skills owner, string skill, string server)
    {
        lock (gate)
        {
            if (destroyed || claimed)
            {
                throw new InvalidDataException("The browser connection is closed or already claimed; reconnect before loading.");
            }
            // The constructor attaches through each parent's methods. Nothing
            // grants tools until the skill's existing transactional commit.
            return new McpClient(Parent(), server, transport, owner, skill, this);
        }
    }

    public partial void AddClient(McpClient child)
    {
        lock (gate)
        {
            if (destroyed || claimed || child.destroyed || child.browserOwner is not null
                || !ReferenceEquals(child.Parent(), Parent())
                || !ReferenceEquals(child.transport, transport)
                || child.SkillOwner() is null)
            {
                throw new InvalidOperationException("Browser client is closed, foreign, or already owned.");
            }
            client = child;
            child.browserOwner = this;
            claimed = true;
        }
    }

    public partial void RemoveClient(McpClient child)
    {
        lock (gate)
        {
            if (!ReferenceEquals(client, child) || !ReferenceEquals(child.browserOwner, this))
            {
                throw new InvalidOperationException("Client belongs to another browser.");
            }
            client = null;
            child.browserOwner = null;
            // Remain consumed. A dead protocol session cannot silently become
            // a different session on a still-connected browser tab.
        }
    }

    public partial Task DestroyAsync()
    {
        lock (gate)
        {
            destroyed = true;
            destruction ??= Task.Run(DestroyCoreAsync);
            return destruction;
        }
    }

    private partial async Task DestroyCoreAsync()
    {
        try
        {
            var owned = Client();
            if (owned is not null)
            {
                await owned.DestroyAsync();
            }
            else
            {
                await transport.DisposeAsync();
            }
        }
        finally
        {
            parent?.RemoveBrowser(this);
        }
    }
}
