namespace Ensemble;

public sealed partial class Agent
{
    public partial Task SetModelAsync(string model)
    {
        var reply = new TaskCompletionSource<string>(TaskCreationOptions.RunContinuationsAsynchronously);
        Enqueue(new Mail(MailKind.SelectModel, Model: model, Reply: reply));
        return reply.Task;
    }

    public partial Task SetModelAsync(Model model)
    {
        ArgumentNullException.ThrowIfNull(model);
        var reply = new TaskCompletionSource<string>(TaskCreationOptions.RunContinuationsAsynchronously);
        Enqueue(new Mail(MailKind.SelectModel, ConfiguredModel: model, Reply: reply));
        return reply.Task;
    }

    private partial void SelectModel(Mail message)
    {
        try
        {
            ObjectDisposedException.ThrowIf(stopping, this);
            // Queued prompts count as admitted work too. A selection must not
            // jump ahead of a paused turn or a skill transaction still starting.
            if (active is not null || requests.Count != 0 || workerPending
                || transportPending || automaticPending || startupReplies.Count != 0
                || controlReplies.Count != 0)
            {
                throw new InvalidOperationException("Model selection requires an idle Agent between turns.");
            }
            lock (ingress)
            {
                ObjectDisposedException.ThrowIf(destroyed, this);
                if (engineSends != 0)
                {
                    throw new InvalidOperationException("Model selection cannot change an active Engine send.");
                }
                // Validation constructs a complete immutable candidate first.
                // The meter/lens stay on the existing Engine; neither History
                // nor saved preferences becomes a second selection authority.
                if (message.ConfiguredModel is { } model)
                {
                    Ensemble root = Parent();
                    // Refuse a foreign graph before taking its gate. All further
                    // operations use ingress -> this root's graph, never two roots.
                    if (!ReferenceEquals(model.graphGate, root.ModelGraphGate))
                    {
                        throw new ArgumentException("The selected Model belongs to another Ensemble.");
                    }
                    lock (root.ModelGraphGate)
                    {
                        ObjectDisposedException.ThrowIf(root.ModelGraphClosed, root);
                        if (model.destroyed || !ReferenceEquals(model.Vendor().Parent(), root))
                        {
                            throw new InvalidOperationException("The selected Model is unavailable.");
                        }
                        // An explicit choice can recover from a removed old Model.
                        // The old snapshot supplies Agent policy, never routing.
                        var candidate = new Configuration(configuration, model);
                        SaveModelSelection(candidate);
                        if (!ReferenceEquals(selectedModel, model))
                        {
                            selectedModel?.RemoveAgent(this);
                            model.AddAgent(this);
                        }
                        configuration = candidate;
                        ResetMemoryPressure();
                    }
                }
                else
                {
                    Configuration candidate = ValidatedConfiguration().WithModel(message.Model!);
                    SaveModelSelection(candidate);
                    configuration = candidate;
                    ResetMemoryPressure();
                }
            }
            message.Reply!.TrySetResult("");
        }
        catch (Exception error) when (error is ArgumentException or InvalidOperationException or IOException or UnauthorizedAccessException
            or System.Text.Json.JsonException or FormatException or InvalidDataException)
        {
            message.Reply!.TrySetException(error);
        }
    }

    private partial void SaveModelSelection(Configuration candidate)
    {
        if (dataDirectory is not null && candidate.Offering is { } offering)
        {
            // One durable authority for explicit configured selections. Legacy
            // launch records remain readable; they are not a second commit point.
            var saved = Memory().ReadSettings();
            Memory().WriteSettings(new AgentSettings(saved?.Memory ?? MemorySettings(),
                saved?.Recall ?? RecallSettings(), saved?.WorkingContext ?? WorkingContextLimits(), offering));
        }
        else
        {
            SaveChildRestart(selection: candidate);
        }
    }

    internal partial Configuration BeginEngineSend()
    {
        lock (ingress)
        {
            ObjectDisposedException.ThrowIf(destroyed, this);
            // Validate before incrementing: a refused admission has no send to
            // release the guard in Engine's finally.
            Ensemble root = Parent();
            lock (root.ModelGraphGate)
            {
                Configuration frozen = ValidatedConfiguration();
                ObjectDisposedException.ThrowIf(root.ModelGraphClosed, root);
                frozen.Authentication?.ValidateEndpoint(new Uri(frozen.BaseUrl), frozen.Target().Surface);
                engineSends++;
                if (goalConfiguration.Enabled) { mailbox.Writer.TryWrite(new Mail(MailKind.GoalRefresh)); }
                return frozen;
            }
        }
    }

    internal partial void EndEngineSend()
    {
        lock (ingress)
        {
            // No cancellation or failure may strand a selection lock. This is
            // released by Engine's finally, including credential-acquisition and
            // response parsing failures. It never cancels the request itself.
            engineSends--;
            if (goalConfiguration.Enabled) { mailbox.Writer.TryWrite(new Mail(MailKind.GoalRefresh)); }
        }
    }
}

public sealed partial class Agent
{
    public partial Model? SelectedModel()
    {
        lock (ingress)
        {
            ObjectDisposedException.ThrowIf(destroyed, this);
            lock (Parent().ModelGraphGate)
            {
                return selectedModel;
            }
        }
    }

    internal partial Configuration DocumentaryConfiguration()
    {
        lock (ingress)
        {
            // Persistence describes the last selection even when its live graph
            // is unavailable. This value is not admission to an Engine send.
            return configuration;
        }
    }

    private partial void AttachConfiguredModel()
    {
        if (configuration.Offering is not { } identity) return;
        lock (Parent().ModelGraphGate)
        {
            Model model = Parent().ResolveModel(identity);
            model.AddAgent(this);
            try
            {
                _ = ValidatedConfiguration();
            }
            catch
            {
                model.RemoveAgent(this);
                throw;
            }
        }
    }

    private partial void DetachConfiguredModel()
    {
        lock (Parent().ModelGraphGate)
        {
            selectedModel?.RemoveAgent(this);
        }
    }

    // Callers hold ingress, or are still constructing the unpublished Agent.
    // Graph operations never acquire ingress in the opposite direction.
    private partial Configuration ValidatedConfiguration()
    {
        if (configuration.Offering is not { } identity) return configuration;
        Ensemble root = Parent();
        lock (root.ModelGraphGate)
        {
            Model model = selectedModel
                ?? throw new InvalidOperationException("The selected Model is no longer available.");
            ModelEndpoint endpoint = model.Parent();
            ModelAuthentication binding = endpoint.Authentication();
            if (model.destroyed
                || !ReferenceEquals(model.Vendor().Parent(), root)
                || model.Identity() != identity
                || !ReferenceEquals(binding, configuration.Authentication)
                || model.Name != configuration.Model
                || model.Features != configuration.Features
                || model.ServingContextTokens != configuration.ServingContextTokens
                || endpoint.Address.AbsoluteUri.TrimEnd('/') != configuration.BaseUrl
                || endpoint.Surface != configuration.Surface)
            {
                throw new InvalidOperationException("The selected Model no longer matches its request configuration.");
            }
            return configuration;
        }
    }
}
