using System.Net;
using Ensemble.Gateway;

// Composition root only: Router owns policy, queues and upgraded connections.
// This executable owns the HTTP listener, periodic saver and console lifetime.
try
{
    return await Run(args);
}
catch (Exception error)
{
    Console.Error.WriteLine($"gateway: {error}");
    return 1;
}

static async Task<int> Run(string[] arguments)
{
    string configPath = "gateway.json";
    bool printKey = false;
    for (int index = 0; index < arguments.Length; index++)
    {
        switch (arguments[index])
        {
            case "--config" when index + 1 < arguments.Length:
                configPath = arguments[++index];
                break;
            case "--print-key":
                printKey = true;
                break;
            case "--help":
                Console.WriteLine("GatewayHost [--config gateway.json] [--print-key]");
                return 0;
            default:
                throw new ArgumentException("Expected --config PATH or --print-key; use --help for usage.");
        }
    }

    RouterConfig config = Router.ReadConfig(configPath);
    ArgumentNullException.ThrowIfNull(config.Gateway);
    if (string.IsNullOrWhiteSpace(config.Gateway.Listen) || string.IsNullOrWhiteSpace(config.Gateway.DataDirectory))
        throw new ArgumentException("gateway.listen and gateway.data_dir are required.");
    Keypair key = Keys.LoadOrCreateKeypair(Path.Combine(config.Gateway.DataDirectory, "identity.key"));
    // Identity output is not readiness: a launcher must also observe this process
    // alive and accepting connections, then pin this key in its Noise handshake.
    Console.WriteLine(Keys.EncodePublicKey(key.Public.AsSpan()));
    if (printKey)
        return 0;

    Router router = new(config, key);
    WebApplication? app = null;
    Task? saver = null;
    List<Exception> failures = [];
    try
    {
        IPEndPoint endpoint = await ListenEndpoint(config.Gateway.Listen);
        WebApplicationBuilder builder = WebApplication.CreateSlimBuilder(new WebApplicationOptions
        {
            Args = [],
            ApplicationName = typeof(Router).Assembly.FullName
        });
        // Operator JSON is the listener authority, not ambient ASPNETCORE_URLS
        // or an appsettings file from the caller's working directory.
        builder.Configuration.Sources.Clear();
        builder.Logging.ClearProviders();
        builder.WebHost.ConfigureKestrel(options =>
        {
            options.Limits.RequestHeadersTimeout = TimeSpan.FromSeconds(10);
            options.Listen(endpoint);
        });
        app = builder.Build();
        app.UseWebSockets();
        app.MapGet("/v1/ws", async context =>
        {
            var ids = context.Request.Query["agent_id"];
            if (ids.Count != 1 || !router.ContainsIdentity(ids[0] ?? ""))
            {
                context.Response.StatusCode = StatusCodes.Status403Forbidden;
                return;
            }
            if (!context.WebSockets.IsWebSocketRequest)
            {
                context.Response.StatusCode = StatusCodes.Status400BadRequest;
                return;
            }

            // AcceptAsync takes ownership even if admission closes after preflight.
            // Awaiting it keeps the HTTP request alive until all socket work joins.
            var socket = await context.WebSockets.AcceptWebSocketAsync();
            try
            {
                await router.AcceptAsync(ids[0]!, socket, context.RequestAborted);
            }
            catch (Exception error)
            {
                Console.Error.WriteLine($"gateway connection: {error.Message}");
            }
        });
        await app.StartAsync();
        CancellationToken stopping = app.Lifetime.ApplicationStopping;
        saver = SavePeriodically(router, stopping);
        try
        {
            await Task.Delay(Timeout.InfiniteTimeSpan, stopping);
        }
        catch (OperationCanceledException) when (stopping.IsCancellationRequested)
        {
            // SIGINT/SIGTERM requests normal shutdown through ConsoleLifetime.
        }
    }
    catch (Exception error)
    {
        failures.Add(error);
    }
    finally
    {
        Task? httpStop = null;
        if (app is not null)
        {
            app.Lifetime.StopApplication();
            // Begin unbinding before joining WebSockets. Awaiting HTTP stop first
            // could wait on upgraded requests whose lifetime belongs to Router.
            try
            {
                httpStop = app.StopAsync();
            }
            catch (Exception error)
            {
                failures.Add(error);
            }
        }
        if (saver is not null)
        {
            try
            {
                await saver;
            }
            catch (Exception error)
            {
                failures.Add(error);
            }
        }
        try
        {
            await router.DisposeAsync();
        }
        catch (Exception error)
        {
            failures.Add(error);
        }
        if (httpStop is not null)
        {
            try
            {
                await httpStop;
            }
            catch (Exception error)
            {
                failures.Add(error);
            }
        }
        if (app is not null)
        {
            try
            {
                await app.DisposeAsync();
            }
            catch (Exception error)
            {
                failures.Add(error);
            }
        }
    }
    if (failures.Count != 0)
        throw new AggregateException("Gateway host failed; all owned cleanup was attempted.", failures);
    return 0;
}

static async Task SavePeriodically(Router router, CancellationToken stopping)
{
    using PeriodicTimer timer = new(TimeSpan.FromSeconds(30));
    try
    {
        while (await timer.WaitForNextTickAsync(stopping))
        {
            try
            {
                router.Save();
            }
            catch (Exception error)
            {
                // A failed snapshot does not discard live queues or disable later
                // ticks. Final-save failure still makes process shutdown fail.
                Console.Error.WriteLine($"gateway persistence: {error.Message}");
            }
        }
    }
    catch (OperationCanceledException) when (stopping.IsCancellationRequested)
    {
        // Join before Router's final snapshot; no saver survives its owner.
    }
}

static async Task<IPEndPoint> ListenEndpoint(string text)
{
    // Match the reference host:port configuration (including :port and [IPv6]).
    // Resolve a hostname explicitly: Kestrel's URL wildcard behavior must not
    // silently turn a configured local hostname into an all-interface listener.
    int separator = text.LastIndexOf(':');
    if (separator < 0 || !int.TryParse(text[(separator + 1)..], out int port) || port < 0 || port > 65535)
        throw new ArgumentException("gateway.listen must be host:port.");
    string host = text[..separator];
    if (host.StartsWith('[') && host.EndsWith(']'))
        host = host[1..^1];
    if (host.Length == 0)
        return new(IPAddress.IPv6Any, port);
    if (IPAddress.TryParse(host, out IPAddress? address))
        return new(address, port);
    IPAddress[] addresses = await Dns.GetHostAddressesAsync(host);
    address = addresses.FirstOrDefault(value => value.AddressFamily == System.Net.Sockets.AddressFamily.InterNetwork)
        ?? addresses.FirstOrDefault()
        ?? throw new ArgumentException("gateway.listen hostname has no IP address.");
    return new(address, port);
}
