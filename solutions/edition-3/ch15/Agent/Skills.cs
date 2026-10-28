namespace Ensemble;

public sealed partial class Agent
{
    public partial Skills? SkillManager() => skills;

    public partial void AddSkills(Skills child)
    {
        if (destroyed || child.destroyed || skills is not null || child.parent is not null)
        {
            throw new InvalidOperationException("Skills are already attached or Agent is destroyed");
        }
        skills = child;
        child.parent = this;
    }

    public partial void RemoveSkills(Skills child)
    {
        if (!ReferenceEquals(skills, child) || !ReferenceEquals(child.parent, this))
        {
            throw new InvalidOperationException("Skills are not owned here");
        }
        skills = null;
        child.parent = null;
    }

    public partial bool ToolEnabled(string name)
    {
        if (IsAutomaticTool(name))
        {
            return false;
        }
        var context = History().View();
        var active = context.Dialogue.SelectMany(entry => entry.Parts).OfType<SkillPart>().ToArray();
        return (active.Length == 0 && skills is null)
            || context.EnabledTools().Contains(name);
    }
}

public sealed partial class Agent
{
    private partial async Task StartupWorker(Mail request, Task work)
    {
        Exception? error = null;
        try
        {
            await work;
        }
        catch (Exception failure)
        {
            error = failure;
        }
        // Startup is lifecycle work, not a tool exchange. The actor owns the
        // reliable reply and keeps the mailbox alive until this arrives.
        mailbox.Writer.TryWrite(new Mail(MailKind.ControlDone, Reply: request.Reply, Error: error));
    }

    internal partial Task InitializeSkillsAsync()
    {
        var reply = new TaskCompletionSource<string>(TaskCreationOptions.RunContinuationsAsynchronously);
        Enqueue(new Mail(MailKind.InitializeSkills, Reply: reply));
        return reply.Task;
    }
}
