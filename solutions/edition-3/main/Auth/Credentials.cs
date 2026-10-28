using System.Collections.Immutable;
using System.Text.Json;
using System.Text.Json.Nodes;

namespace Ensemble;

public sealed partial class Ensemble
{
    public partial Credentials? Credentials()
    {
        lock (credentialsGate)
        {
            return credentials;
        }
    }

    internal partial void AddCredentials(Credentials child)
    {
        lock (credentialsGate)
        {
            ObjectDisposedException.ThrowIf(destroyed, this);
            if (credentials is not null || child.parent is not null || child.destroyed)
            {
                throw new InvalidOperationException("Invalid credential store registration.");
            }
            credentials = child;
            child.parent = this;
        }
    }

    internal partial void RemoveCredentials(Credentials child)
    {
        lock (credentialsGate)
        {
            if (!ReferenceEquals(credentials, child) || !ReferenceEquals(child.parent, this))
            {
                throw new InvalidOperationException("Credential store is not owned here.");
            }
            credentials = null;
            child.parent = null;
        }
    }
}

internal sealed partial class AuthRegistration
{
    internal partial AuthRegistration(string label, string issuer, string? subject,
        string clientID, string? email, string? accessToken, string? refreshToken,
        string? idToken, DateTimeOffset? expiresAt, string? scope)
    {
        Label = label;
        Issuer = issuer;
        Subject = subject;
        ClientID = clientID;
        Email = email;
        AccessToken = accessToken;
        RefreshToken = refreshToken;
        IDToken = idToken;
        ExpiresAt = expiresAt;
        Scope = scope;
    }

    internal partial AuthAccount Account() =>
        new(Label, Issuer, Subject, ClientID, Email, ExpiresAt, AccessToken is not null,
            AccessToken is not null && AuthProtocol.HasPlanScope(Scope));

    internal partial AuthRegistration WithoutTokens() =>
        new(Label, Issuer, Subject, ClientID, Email, null, null, null, null, null);
}

internal sealed partial class AuthFile
{
    internal partial AuthFile(string hostID, ImmutableArray<AuthRegistration> registrations, string? selectedPlanAccount)
    {
        HostID = hostID;
        Registrations = registrations;
        SelectedPlanAccount = selectedPlanAccount;
    }
}

public sealed partial class Credentials
{
    public partial Credentials(Ensemble parent, string directory)
        : this(parent, directory, new Uri("https://auth.openai.com"))
    {
    }

    internal partial Credentials(Ensemble parent, string directory, Uri issuer)
    {
        ArgumentException.ThrowIfNullOrWhiteSpace(directory);
        if (!issuer.IsAbsoluteUri || issuer.AbsolutePath != "/" || issuer.Query.Length != 0
            || issuer.Fragment.Length != 0 || issuer.UserInfo.Length != 0
            || (issuer.Scheme != "https" && !(issuer.Scheme == "http" && issuer.Host == "127.0.0.1")))
        {
            throw new ArgumentException("Authentication requires an HTTPS issuer or explicit loopback test issuer.");
        }
        if (OperatingSystem.IsWindows())
        {
            // Do not pretend Unix modes secure a Windows file. A future Windows
            // host needs an explicit ACL implementation before storing tokens.
            throw new PlatformNotSupportedException("Protected credentials currently require Unix file permissions.");
        }
        this.directory = Path.GetFullPath(directory);
        this.issuer = issuer;
        http = new HttpClient(new HttpClientHandler { AllowAutoRedirect = false, UseCookies = false })
        {
            Timeout = Timeout.InfiniteTimeSpan
        };
        try
        {
            parent.AddCredentials(this);
        }
        catch
        {
            http.Dispose();
            lifetime.Dispose();
            throw;
        }
    }

    public partial Ensemble Parent() => parent ?? throw new ObjectDisposedException(nameof(Credentials));

    public partial async Task<ImmutableArray<AuthAccount>> AccountsAsync(CancellationToken cancellation)
    {
        await gate.WaitAsync(cancellation).ConfigureAwait(false);
        try
        {
            ObjectDisposedException.ThrowIf(destroyed, this);
            using var active = CancellationTokenSource.CreateLinkedTokenSource(cancellation, lifetime.Token);
            cancellation = active.Token;
            using var held = await LockFileAsync(cancellation).ConfigureAwait(false);
            return ReadFile().Registrations.Select(registration => registration.Account()).ToImmutableArray();
        }
        finally
        {
            gate.Release();
        }
    }

    public partial async Task<string?> SelectedPlanAccountAsync(CancellationToken cancellation)
    {
        await gate.WaitAsync(cancellation).ConfigureAwait(false);
        try
        {
            ObjectDisposedException.ThrowIf(destroyed, this);
            using var active = CancellationTokenSource.CreateLinkedTokenSource(cancellation, lifetime.Token);
            using var held = await LockFileAsync(active.Token).ConfigureAwait(false);
            return ReadFile().SelectedPlanAccount;
        }
        finally
        {
            gate.Release();
        }
    }

    internal partial async Task<string> HostIDAsync(CancellationToken cancellation)
    {
        await gate.WaitAsync(cancellation).ConfigureAwait(false);
        try
        {
            ObjectDisposedException.ThrowIf(destroyed, this);
            using var active = CancellationTokenSource.CreateLinkedTokenSource(cancellation, lifetime.Token);
            cancellation = active.Token;
            using var held = await LockFileAsync(cancellation).ConfigureAwait(false);
            var value = ReadFile();
            // The same lock covers initial creation: two processes cannot choose
            // different host IDs and then each publish its own OAuth registration.
            WriteFile(value);
            return value.HostID;
        }
        finally
        {
            gate.Release();
        }
    }

    internal partial async Task StoreAsync(AuthRegistration registration, CancellationToken cancellation)
    {
        Validate(registration);
        await gate.WaitAsync(cancellation).ConfigureAwait(false);
        try
        {
            ObjectDisposedException.ThrowIf(destroyed, this);
            using var active = CancellationTokenSource.CreateLinkedTokenSource(cancellation, lifetime.Token);
            cancellation = active.Token;
            using var held = await LockFileAsync(cancellation).ConfigureAwait(false);
            var prior = ReadFile();
            var values = prior.Registrations.ToBuilder();
            int index = -1;
            for (int i = 0; i < values.Count; i++)
            {
                var existing = values[i];
                if (existing.Label == registration.Label)
                {
                    // Reauthentication may rotate credentials, not silently bind
                    // a familiar account label to someone else's registration.
                    if (!SameRegistration(existing, registration))
                    {
                        throw new InvalidOperationException("Account label belongs to another registration.");
                    }
                    index = i;
                }
                else if (existing.Issuer == registration.Issuer && existing.ClientID == registration.ClientID)
                {
                    throw new InvalidOperationException("Issued client already belongs to another account label.");
                }
            }
            if (index < 0)
            {
                values.Add(registration);
            }
            else
            {
                values[index] = registration;
            }
            WriteFile(new AuthFile(prior.HostID, values.ToImmutable(), prior.SelectedPlanAccount));
        }
        finally
        {
            gate.Release();
        }
    }

    public partial async Task LogoutAsync(string label, CancellationToken cancellation)
    {
        ArgumentException.ThrowIfNullOrWhiteSpace(label);
        await gate.WaitAsync(cancellation).ConfigureAwait(false);
        try
        {
            ObjectDisposedException.ThrowIf(destroyed, this);
            using var active = CancellationTokenSource.CreateLinkedTokenSource(cancellation, lifetime.Token);
            cancellation = active.Token;
            using var held = await LockFileAsync(cancellation).ConfigureAwait(false);
            var prior = ReadFile();
            var values = prior.Registrations.ToBuilder();
            int index = -1;
            for (int i = 0; i < values.Count; i++)
            {
                if (values[i].Label == label)
                {
                    index = i;
                    break;
                }
            }
            if (index < 0)
            {
                throw new InvalidOperationException("Unknown account label.");
            }
            values[index] = values[index].WithoutTokens();
            // Publish local logout before any caller attempts remote revocation.
            // Failure here is an error, never a reported successful logout.
            WriteFile(new AuthFile(prior.HostID, values.ToImmutable(), prior.SelectedPlanAccount));
        }
        finally
        {
            gate.Release();
        }
    }

    public partial void Destroy()
    {
        // Cancel network/lock waits, then join through the operation gate before
        // disposing transport or detaching. Normal shutdown does not sign out.
        try
        {
            lifetime.Cancel();
        }
        catch (ObjectDisposedException)
        {
            // Another Destroy already disposed the lifetime; still join its gate.
        }
        gate.Wait();
        try
        {
            if (destroyed)
            {
                return;
            }
            destroyed = true;
            login = null;
            http.Dispose();
            lifetime.Dispose();
            Parent().RemoveCredentials(this);
        }
        finally
        {
            // Keep the semaphore usable for already-waiting callers, which will
            // observe destroyed rather than race ObjectDisposedException in Wait.
            gate.Release();
        }
    }

    private static bool SameRegistration(AuthRegistration left, AuthRegistration right) =>
        left.Issuer == right.Issuer && left.ClientID == right.ClientID
            && (left.Subject is null || left.Subject == right.Subject);

    private static void Validate(AuthRegistration value)
    {
        if (string.IsNullOrWhiteSpace(value.Label) || string.IsNullOrWhiteSpace(value.Issuer)
            || (value.Subject is not null && string.IsNullOrWhiteSpace(value.Subject))
            || string.IsNullOrWhiteSpace(value.ClientID)
            || value.ClientID == "dynamic_agent_client")
        {
            throw new InvalidDataException("Invalid credential registration metadata.");
        }
        bool hasTokens = value.AccessToken is not null;
        if (hasTokens)
        {
            if (value.Subject is null || string.IsNullOrWhiteSpace(value.AccessToken) || string.IsNullOrWhiteSpace(value.RefreshToken)
                || string.IsNullOrWhiteSpace(value.IDToken) || value.ExpiresAt is null
                || string.IsNullOrWhiteSpace(value.Scope))
            {
                throw new InvalidDataException("Incomplete credential grant.");
            }
        }
        else if (value.RefreshToken is not null || value.IDToken is not null
            || value.ExpiresAt is not null || value.Scope is not null)
        {
            throw new InvalidDataException("Incomplete signed-out credential record.");
        }
    }

    private partial async Task<FileStream> LockFileAsync(CancellationToken cancellation)
    {
        if (OperatingSystem.IsWindows())
        {
            throw new PlatformNotSupportedException("Protected credentials require Unix file permissions.");
        }
        // Only this dedicated directory is created. Never chmod a user's shared
        // configuration directory as a side effect of reading account status.
        Directory.CreateDirectory(directory, UnixFileMode.UserRead | UnixFileMode.UserWrite | UnixFileMode.UserExecute);
        CheckPrivate(directory, isDirectory: true);
        string path = Path.Combine(directory, "credentials.lock");
        while (true)
        {
            cancellation.ThrowIfCancellationRequested();
            CheckPrivate(path, isDirectory: false, allowMissing: true);
            try
            {
                var held = new FileStream(path, new FileStreamOptions
                {
                    Mode = FileMode.OpenOrCreate,
                    Access = FileAccess.ReadWrite,
                    Share = FileShare.None,
                    UnixCreateMode = UnixFileMode.UserRead | UnixFileMode.UserWrite
                });
                try
                {
                    CheckPrivate(path, isDirectory: false);
                    return held;
                }
                catch
                {
                    held.Dispose();
                    throw;
                }
            }
            catch (IOException error) when (
                error.HResult == (OperatingSystem.IsMacOS() || OperatingSystem.IsFreeBSD() ? 35 : 11)
                || error.HResult == unchecked((int)0x80070020))
            {
                // .NET exposes EWOULDBLOCK as errno 35 on macOS/BSD and 11 on
                // Linux; a normalized sharing violation is 0x80070020. Do not
                // mistake every IOException for contention. The OS releases
                // the lock if its owning process crashes.
                await Task.Delay(30, cancellation).ConfigureAwait(false);
            }
        }
    }

    private static void CheckPrivate(string path, bool isDirectory, bool allowMissing = false)
    {
        if (OperatingSystem.IsWindows())
        {
            throw new PlatformNotSupportedException("Protected credentials require Unix file permissions.");
        }
        FileAttributes attributes;
        try
        {
            attributes = File.GetAttributes(path);
        }
        catch (FileNotFoundException) when (allowMissing)
        {
            return;
        }
        if ((attributes & FileAttributes.ReparsePoint) != 0
            || ((attributes & FileAttributes.Directory) != 0) != isDirectory)
        {
            throw new IOException("Credential storage must use ordinary private files and directories.");
        }
        UnixFileMode expected = UnixFileMode.UserRead | UnixFileMode.UserWrite;
        if (isDirectory)
        {
            expected |= UnixFileMode.UserExecute;
        }
        if (File.GetUnixFileMode(path) != expected)
        {
            throw new IOException("Credential storage permissions must be owner-only (directory 0700, files 0600).");
        }
    }

    private partial AuthFile ReadFile()
    {
        string path = Path.Combine(directory, "credentials.json");
        CheckPrivate(path, isDirectory: false, allowMissing: true);
        byte[] bytes;
        try
        {
            bytes = File.ReadAllBytes(path);
        }
        catch (FileNotFoundException)
        {
            return new AuthFile("urn:uuid:" + Guid.NewGuid().ToString("D"), []);
        }
        try
        {
            using var document = JsonDocument.Parse(bytes);
            var root = document.RootElement;
            // Version 1 had no selection. Strict fields remain version-specific;
            // malformed/newer files never silently become API defaults.
            if (root.ValueKind != JsonValueKind.Object || !root.TryGetProperty("version", out _))
            {
                throw new InvalidDataException("Invalid credential format.");
            }
            var encodedVersion = root.GetProperty("version");
            if (encodedVersion.ValueKind != JsonValueKind.Number
                || !encodedVersion.TryGetInt32(out int version) || version is not (1 or 2))
            {
                throw new InvalidDataException("Unsupported credential format.");
            }
            if (version == 1)
            {
                Fields(root, "version", "host_id", "accounts");
            }
            else
            {
                Fields(root, "version", "host_id", "accounts", "selected_plan_account");
            }
            string? selected = version == 1 ? null : Optional(root, "selected_plan_account");
            string hostID = Required(root, "host_id");
            var accounts = root.GetProperty("accounts");
            if (accounts.ValueKind != JsonValueKind.Array)
            {
                throw new InvalidDataException("Invalid credential accounts.");
            }
            var values = ImmutableArray.CreateBuilder<AuthRegistration>();
            foreach (var account in accounts.EnumerateArray())
            {
                Fields(account, "label", "issuer", "subject", "client_id", "email", "access_token",
                    "refresh_token", "id_token", "expires_at", "scope");
                string? timestamp = Optional(account, "expires_at");
                DateTimeOffset? expiry = null;
                if (timestamp is not null)
                {
                    if (!DateTimeOffset.TryParseExact(timestamp, "O", System.Globalization.CultureInfo.InvariantCulture,
                        System.Globalization.DateTimeStyles.None, out var parsed))
                    {
                        throw new InvalidDataException("Invalid credential expiry.");
                    }
                    expiry = parsed;
                }
                var value = new AuthRegistration(Required(account, "label"), Required(account, "issuer"),
                    Optional(account, "subject"), Required(account, "client_id"), Optional(account, "email"),
                    Optional(account, "access_token"), Optional(account, "refresh_token"), Optional(account, "id_token"),
                    expiry, Optional(account, "scope"));
                Validate(value);
                if (values.Any(prior => prior.Label == value.Label
                    || (prior.Issuer == value.Issuer && prior.ClientID == value.ClientID)))
                {
                    throw new InvalidDataException("Duplicate credential registration.");
                }
                values.Add(value);
            }
            if (selected is not null && !values.Any(account => account.Label == selected))
            {
                throw new InvalidDataException("Invalid saved funding selection.");
            }
            return new AuthFile(hostID, values.ToImmutable(), selected);
        }
        catch (JsonException)
        {
            // The parser's context may include private bytes. Do not preserve it
            // as an inner exception or echo a damaged credential record.
            throw new InvalidDataException("Invalid credential JSON.");
        }
    }

    private static void Fields(JsonElement value, params string[] names)
    {
        if (value.ValueKind != JsonValueKind.Object)
        {
            throw new InvalidDataException("Invalid credential object.");
        }
        var actual = value.EnumerateObject().Select(property => property.Name).ToArray();
        if (actual.Length != names.Length || actual.Distinct(StringComparer.Ordinal).Count() != actual.Length
            || actual.Any(name => !names.Contains(name, StringComparer.Ordinal)))
        {
            throw new InvalidDataException("Invalid credential fields.");
        }
    }

    private static string Required(JsonElement value, string name) =>
        Optional(value, name) is { } text && !string.IsNullOrWhiteSpace(text)
            ? text : throw new InvalidDataException("Missing credential field.");

    private static string? Optional(JsonElement value, string name)
    {
        var field = value.GetProperty(name);
        if (field.ValueKind == JsonValueKind.Null)
        {
            return null;
        }
        if (field.ValueKind != JsonValueKind.String)
        {
            throw new InvalidDataException("Invalid credential field type.");
        }
        return field.GetString();
    }

    private partial void WriteFile(AuthFile value)
    {
        if (OperatingSystem.IsWindows())
        {
            throw new PlatformNotSupportedException("Protected credentials require Unix file permissions.");
        }
        if (value.SelectedPlanAccount is not null
            && !value.Registrations.Any(account => account.Label == value.SelectedPlanAccount))
        {
            throw new InvalidDataException("Invalid saved funding selection.");
        }
        var accounts = new JsonArray();
        foreach (var registration in value.Registrations)
        {
            Validate(registration);
            accounts.Add(new JsonObject
            {
                ["label"] = registration.Label, ["issuer"] = registration.Issuer,
                ["subject"] = registration.Subject, ["client_id"] = registration.ClientID,
                ["email"] = registration.Email, ["access_token"] = registration.AccessToken,
                ["refresh_token"] = registration.RefreshToken, ["id_token"] = registration.IDToken,
                ["expires_at"] = registration.ExpiresAt?.ToString("O"), ["scope"] = registration.Scope
            });
        }
        byte[] bytes = JsonSerializer.SerializeToUtf8Bytes(new JsonObject
        {
            ["version"] = 2, ["host_id"] = value.HostID, ["accounts"] = accounts,
            ["selected_plan_account"] = value.SelectedPlanAccount
        });
        string target = Path.Combine(directory, "credentials.json");
        CheckPrivate(target, isDirectory: false, allowMissing: true);
        string temporary = Path.Combine(directory, ".credentials-" + Guid.NewGuid().ToString("N"));
        try
        {
            using (var output = new FileStream(temporary, new FileStreamOptions
            {
                Mode = FileMode.CreateNew, Access = FileAccess.Write, Share = FileShare.None,
                UnixCreateMode = UnixFileMode.UserRead | UnixFileMode.UserWrite
            }))
            {
                output.Write(bytes);
                output.Flush(flushToDisk: true);
            }
            // Same-directory replacement gives readers a complete old or new
            // record. No directory fsync/power-loss durability guarantee is made.
            File.Move(temporary, target, overwrite: true);
        }
        finally
        {
            File.Delete(temporary);
        }
    }
}
