namespace Writing;

// Value-like document, not a relationship-bearing model object. All tools share
// this one text value; there are no Agent-owned mutable manuscript copies.
// It needs synchronization because independent Jobs may outlive their turns.
public sealed partial class Draft
{
    private readonly object gate = new();
    private string text = "";
    public partial string Read();
    public partial string Replace(string replacement);
    public partial int WordCount();
}
