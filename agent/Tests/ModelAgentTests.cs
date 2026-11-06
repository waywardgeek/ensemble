using Ensemble;
using Application = Ensemble.Ensemble;

internal static class ModelAgentTests
{
    private static void Check(bool condition, string reason)
    {
        if (!condition) throw new Exception(reason);
    }

    private static void Refuses(Action action)
    {
        try { action(); }
        catch (Exception error) when (error is InvalidOperationException or ArgumentException) { return; }
        throw new Exception("Expected configured relationship refusal.");
    }

    private static Model Configure(Application root)
    {
        var auth = new ModelAuthentication(root, "none");
        var vendor = new ModelVendor(root, "local", "Local");
        var endpoint = new ModelEndpoint(vendor, "test", "http://127.0.0.1:1",
            Surface.ChatCompletions, auth);
        return new Model(endpoint, "small", "same-wire", "Small", new ModelFeatures(Media.None));
    }

    public static void Membership(string temp)
    {
        var root = new Application(TextWriter.Null);
        var foreign = new Application(TextWriter.Null);
        try
        {
            Model model = Configure(root);
            Model foreignModel = Configure(foreign);
            var policy = new Configuration(Vendor.Anthropic, "legacy", "http://127.0.0.1:2", "synthetic");
            var projected = new Configuration(policy, model);
            var agent = new Agent(root, projected);
            var sibling = new Agent(root, projected);
            Check(ReferenceEquals(agent.SelectedModel(), model)
                && model.Agents().Length == 2 && model.Agents().Contains(agent),
                "Construction establishes both weak selection ends.");
            Check(ReferenceEquals(agent.Configuration(), projected),
                "Validated reads retain the frozen policy projection.");
            ModelSelectionStatus selection = agent.ModelSelectionStatus();
            Check(selection.Available && selection.Offering == model.Identity()
                && selection.LegacyChoices.IsEmpty && selection.ConfiguredChoices.Length == 1
                && selection.ConfiguredChoices[0].Available,
                "Selection status describes a configured choice without legacy route escape.");
            var otherAuth = new ModelAuthentication(root, "other-none");
            var otherEndpoint = new ModelEndpoint(model.Vendor(), "other", "http://127.0.0.1:3",
                Surface.ChatCompletions, otherAuth);
            var otherModel = new Model(otherEndpoint, "small", model.Name, "Other", new ModelFeatures(Media.None));
            selection = agent.ModelSelectionStatus();
            Check(selection.ConfiguredChoices.Length == 2
                && selection.ConfiguredChoices.Select(choice => choice.Offering).Distinct().Count() == 2,
                "Same wire names retain distinct configured selection identities.");
            otherAuth.Destroy();
            selection = agent.ModelSelectionStatus();
            Check(selection.Available && !selection.ConfiguredChoices.Single(
                choice => choice.Offering == otherModel.Identity()).Available,
                "Unavailable authentication disables only its own documented choice.");
            Refuses(() => new Agent(foreign, projected));
            Check(foreign.Agents().Length == 0 && foreignModel.Agents().Length == 0,
                "Same identity in another root cannot borrow the original binding; rollback removes memberships.");
            agent.Destroy();
            Check(model.Agents().Length == 1 && model.Agents()[0] == sibling,
                "Agent destruction detaches only its own weak selection.");

            ModelEndpoint endpoint = model.Parent();
            endpoint.RemoveModel(model);
            Refuses(() => sibling.Configuration());
            Refuses(() => sibling.BeginEngineSend());
            endpoint.AddModel(model);
            Configuration frozen = sibling.BeginEngineSend();
            try
            {
                Check(frozen.Offering == model.Identity(), "Reattached original relationship can admit a request.");
            }
            finally { sibling.EndEngineSend(); }
            // A refused Begin must not strand the send guard used by this setter.
            sibling.SetModelAsync(model.Name).GetAwaiter().GetResult();
            model.Destroy();
            Check(model.Agents().IsEmpty && sibling.SelectedModel() is null && !sibling.destroyed,
                "Model destruction clears weak links without destroying an Agent.");
            Refuses(() => sibling.Configuration());
            Refuses(() => sibling.BeginEngineSend());
            Refuses(() => sibling.Engine().Render(new Context()));
            ModelSelectionStatus unavailable = sibling.ModelSelectionStatus();
            Check(!unavailable.Available && unavailable.Offering == projected.Offering
                && unavailable.Model == projected.Model && unavailable.LegacyChoices.IsEmpty,
                "Model loss remains displayable without inference or legacy fallback.");
            sibling.History().Append(new Event(EventKind.MessageReceived, actor: Actor.Human,
                parts: [new TextPart("Conversation survives unavailable inference.")]));
            Check(sibling.History().View().Dialogue.Length != 0,
                "Lost selection preserves conversation rather than falling back to legacy routing.");
            string saved = Path.Combine(temp, "model-loss-save.json");
            sibling.History().Save(new PersistenceOptions(saved));
            Check(File.Exists(saved) && File.ReadAllText(saved).Contains("Conversation survives unavailable inference."),
                "Documentary persistence survives Model loss without admitting inference.");
        }
        finally
        {
            try { root.Destroy(); }
            finally { foreign.Destroy(); }
        }
    }
}
