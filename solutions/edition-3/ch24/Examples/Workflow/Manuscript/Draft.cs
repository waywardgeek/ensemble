namespace Writing;

public sealed partial class Draft
{
    public partial string Read()
    {
        lock (gate)
        {
            return text;
        }
    }

    public partial string Replace(string replacement)
    {
        ArgumentNullException.ThrowIfNull(replacement);
        lock (gate)
        {
            text = replacement;
            return text;
        }
    }

    public partial int WordCount()
    {
        // A local immutable snapshot makes counting safe without retaining a
        // second authoritative manuscript or holding the lock while scanning.
        string snapshot = Read();
        return snapshot.Split((char[]?)null, StringSplitOptions.RemoveEmptyEntries).Length;
    }
}
