using System.Collections.Immutable;
using System.Net;
using System.Net.Http.Headers;
using System.Text.Json;

namespace Ensemble;

public sealed partial class Credentials
{
    internal static partial void ValidatePlanEndpoint(Uri endpoint)
    {
        bool localTest = endpoint.Scheme == "http"
            && (endpoint.Host == "127.0.0.1" || endpoint.Host == "[::1]");
        if ((endpoint.Scheme != "https" && !localTest) || endpoint.UserInfo.Length != 0
            || endpoint.Query.Length != 0 || endpoint.Fragment.Length != 0)
        {
            throw new ArgumentException("Plan endpoint requires HTTPS or literal loopback HTTP, without user info, query, or fragment");
        }
    }

    public partial async Task<ImmutableArray<PlanModel>> ModelsAsync(string label,
        string baseUrl, CancellationToken cancellation)
    {
        ArgumentException.ThrowIfNullOrWhiteSpace(label);
        if (!Uri.TryCreate(baseUrl, UriKind.Absolute, out var endpoint))
        {
            throw new ArgumentException("Catalog base URL must be absolute.");
        }
        ValidatePlanEndpoint(endpoint);
        // Accept the same service base, version base, or complete Responses URL
        // as inference. Never derive this address from an untrusted response.
        string address = endpoint.AbsoluteUri.TrimEnd('/');
        if (address.EndsWith("/v1/responses", StringComparison.Ordinal))
        {
            address = address[..^"/responses".Length];
        }
        if (!address.EndsWith("/v1", StringComparison.Ordinal))
        {
            address += "/v1";
        }
        address += "/models";

        await gate.WaitAsync(cancellation).ConfigureAwait(false);
        try
        {
            ObjectDisposedException.ThrowIf(destroyed, this);
            using var active = CancellationTokenSource.CreateLinkedTokenSource(cancellation, lifetime.Token);
            cancellation = active.Token;
            string bearer = await ResolveBearerAsync(label, cancellation).ConfigureAwait(false);
            using var request = new HttpRequestMessage(HttpMethod.Get, address);
            request.Headers.Authorization = new AuthenticationHeaderValue("Bearer", bearer);
            request.Headers.Accept.Add(new MediaTypeWithQualityHeaderValue("application/json"));
            // Credentials' handler never follows redirects or retains cookies.
            // Discovery has no inference, lens, accounting, or History side effects.
            using var response = await http.SendAsync(request, HttpCompletionOption.ResponseHeadersRead,
                cancellation).ConfigureAwait(false);
            if (response.StatusCode != HttpStatusCode.OK)
            {
                // No API-key retry, no credential erasure, and no raw error body.
                throw new HttpRequestException("Plan model discovery was refused.", null, response.StatusCode);
            }
            var document = await ReadAuthJsonAsync(response.Content, cancellation).ConfigureAwait(false);
            if (!document.TryGetProperty("models", out var models) || models.ValueKind != JsonValueKind.Array)
            {
                throw new InvalidDataException("Invalid plan model catalog.");
            }
            var result = ImmutableArray.CreateBuilder<PlanModel>();
            var slugs = new HashSet<string>(StringComparer.Ordinal);
            foreach (var model in models.EnumerateArray())
            {
                AuthProtocol.UniqueObject(model);
                string? visibility = AuthProtocol.OptionalString(model, "visibility");
                if (visibility != "list")
                {
                    continue;
                }
                string? slug = AuthProtocol.OptionalString(model, "slug");
                string? displayName = AuthProtocol.OptionalString(model, "display_name");
                if (string.IsNullOrWhiteSpace(slug) || string.IsNullOrWhiteSpace(displayName)
                    || !slugs.Add(slug))
                {
                    throw new InvalidDataException("Invalid or duplicate visible plan model.");
                }
                result.Add(new PlanModel(slug, displayName));
            }
            // Publish only after validating the whole list. Do not sort, cache,
            // or infer local transport/media/reasoning support from availability.
            return result.ToImmutable();
        }
        finally
        {
            gate.Release();
        }
    }
}
