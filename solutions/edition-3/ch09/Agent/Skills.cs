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
        var active = History().View().Dialogue.SelectMany(entry => entry.Parts).OfType<SkillPart>().ToArray();
        return active.Length == 0 || active.Any(skill => skill.Tools.Contains(name, StringComparer.Ordinal));
    }
}
