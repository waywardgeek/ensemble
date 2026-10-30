using System.Text.Json;

namespace Ensemble;

public sealed partial class Ensemble
{
    public partial CodexCredentials? CodexCredentials()
    {
        lock (credentialsGate)
        {
            return codexCredentials;
        }
    }

    internal partial void AddCodexCredentials(CodexCredentials child)
    {
        lock (credentialsGate)
        {
            ObjectDisposedException.ThrowIf(destroyed, this);
            if (codexCredentials is not null || child.parent is not null || child.destroyed)
            {
                throw new InvalidOperationException("Invalid Codex credential registration.");
            }
            codexCredentials = child;
            child.parent = this;
        }
    }

    internal partial void RemoveCodexCredentials(CodexCredentials child)
    {
        lock (credentialsGate)
        {
            if (!ReferenceEquals(codexCredentials, child) || !ReferenceEquals(child.parent, this))
            {
                throw new InvalidOperationException("Codex credentials are not owned here.");
            }
            codexCredentials = null;
            child.parent = null;
        }
    }
}

internal sealed partial class CodexGrant
{
    internal partial CodexGrant(string bearer, string account)
    {
        Bearer = bearer;
        Account = account;
    }
}

public sealed partial class CodexAuthorizationException
{
    public partial CodexAuthorizationException()
        : base("Codex authorization unavailable. Sign in or renew in Codex and check the selected auth file. No public-OAuth or API-key fallback was used.")
    {
    }
}

public sealed partial class CodexCredentials
{
    public partial CodexCredentials(Ensemble parent, string path)
    {
        ArgumentNullException.ThrowIfNull(parent);
        ArgumentException.ThrowIfNullOrWhiteSpace(path);
        this.path = Path.GetFullPath(path);
        try
        {
            parent.AddCodexCredentials(this);
        }
        catch
        {
            lifetime.Dispose();
            gate.Dispose();
            throw;
        }
    }

    public partial Ensemble Parent() => parent ?? throw new ObjectDisposedException(nameof(CodexCredentials));

    internal static partial void ValidateEndpoint(Uri endpoint)
    {
        // Compare the entire canonical URI, not a suffix or a trusted-looking
        // hostname substring. No loopback exemption exists in production.
        string address = endpoint.AbsoluteUri.TrimEnd('/');
        if (address != ServiceBase && address != ResponsesEndpoint)
        {
            throw new ArgumentException("Codex requires its pinned HTTPS Responses endpoint.");
        }
    }

    internal partial async Task<CodexGrant> ReadAsync(CancellationToken cancellation)
    {
        ObjectDisposedException.ThrowIf(destroyed, this);
        using var linked = CancellationTokenSource.CreateLinkedTokenSource(cancellation, lifetime.Token);
        await gate.WaitAsync(linked.Token);
        try
        {
            ObjectDisposedException.ThrowIf(destroyed, this);
            _ = Parent();
            try
            {
                // One open handle observes one file generation even if Codex
                // atomically replaces its path while this read is in progress.
                using var input = new FileStream(path, FileMode.Open, FileAccess.Read,
                    FileShare.ReadWrite | FileShare.Delete, 4096, FileOptions.Asynchronous);
                using var bytes = new MemoryStream();
                var buffer = new byte[8192];
                while (true)
                {
                    int count = await input.ReadAsync(buffer, linked.Token);
                    if (count == 0)
                    {
                        break;
                    }
                    if (bytes.Length + count > 1024 * 1024)
                    {
                        throw new InvalidDataException();
                    }
                    bytes.Write(buffer, 0, count);
                }
                using var document = JsonDocument.Parse(bytes.ToArray());
                var root = document.RootElement;
                RequireUniqueProperties(root);
                if (root.ValueKind != JsonValueKind.Object)
                {
                    throw new InvalidDataException();
                }
                // Earlier Codex files omitted auth_mode. An explicitly selected
                // API-key/external mode is never accepted as a ChatGPT grant.
                if (root.TryGetProperty("auth_mode", out var mode)
                    && (mode.ValueKind != JsonValueKind.String || mode.GetString() != "chatgpt"))
                {
                    throw new InvalidDataException();
                }
                if (!root.TryGetProperty("tokens", out var tokens)
                    || tokens.ValueKind != JsonValueKind.Object)
                {
                    throw new InvalidDataException();
                }
                string bearer = HeaderValue(tokens, "access_token");
                string account = HeaderValue(tokens, "account_id");
                linked.Token.ThrowIfCancellationRequested();
                return new CodexGrant(bearer, account);
            }
            catch (Exception error) when (error is IOException or InvalidDataException or UnauthorizedAccessException or JsonException)
            {
                // Never attach the original exception: it may quote private
                // JSON, a path, or operating-system diagnostic content.
                throw new CodexAuthorizationException();
            }
        }
        finally
        {
            gate.Release();
        }
    }

    private static string HeaderValue(JsonElement tokens, string name)
    {
        if (!tokens.TryGetProperty(name, out var property) || property.ValueKind != JsonValueKind.String)
        {
            throw new InvalidDataException();
        }
        string value = property.GetString()!;
        if (value.Length == 0 || value.Any(character => character < '!' || character > '~'))
        {
            throw new InvalidDataException();
        }
        return value;
    }

    private static void RequireUniqueProperties(JsonElement value)
    {
        if (value.ValueKind == JsonValueKind.Object)
        {
            AuthProtocol.UniqueObject(value);
            foreach (var property in value.EnumerateObject())
            {
                RequireUniqueProperties(property.Value);
            }
        }
        else if (value.ValueKind == JsonValueKind.Array)
        {
            foreach (var item in value.EnumerateArray())
            {
                RequireUniqueProperties(item);
            }
        }
    }

    public partial void Destroy()
    {
        try
        {
            lifetime.Cancel();
        }
        catch (ObjectDisposedException)
        {
            // Another destroy still owns or has finished the joined teardown.
        }
        gate.Wait();
        try
        {
            if (destroyed)
            {
                return;
            }
            destroyed = true;
            lifetime.Dispose();
            parent?.RemoveCodexCredentials(this);
        }
        finally
        {
            // Already-admitted waiters must observe cancellation/destruction,
            // not race a disposed semaphore. No file operation occurs here.
            gate.Release();
        }
    }
}
