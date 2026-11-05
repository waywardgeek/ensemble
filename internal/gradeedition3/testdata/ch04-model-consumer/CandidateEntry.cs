using Ensemble;

// API spelling and the later streaming override are compilation choices, not a
// replacement Agent loop. Both branches call the candidate's real public owner.
internal static class CandidateEntry
{
    public static Configuration Configuration(string endpoint, string key)
    {
#if EARLY_AGENT_ASK
        return new Configuration(Vendor.Anthropic, "claude-3-haiku-20240307", endpoint, key);
#else
        return new Configuration(Vendor.Anthropic, "claude-3-haiku-20240307", endpoint, key,
            disableStreaming: true);
#endif
    }

    public static Task<string> Run(Agent agent, string prompt, CancellationToken cancellation)
    {
#if EARLY_AGENT_ASK
        return agent.AskAsync(prompt, cancellation);
#else
        return agent.SendAsync(prompt, cancellation);
#endif
    }
}
