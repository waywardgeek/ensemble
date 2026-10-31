namespace Ensemble;

// These immutable records are values, not relationship-bearing objects.
// They need no holder back-pointers or logical destruction.
public sealed record Configuration(string BaseUrl, string ApiKey, string Model);
public sealed record Message(string Role, string Content);
public readonly record struct Usage(long Input, long Output);
