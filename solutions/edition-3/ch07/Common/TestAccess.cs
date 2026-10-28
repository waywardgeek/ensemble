using System.Runtime.CompilerServices;

// Test the internal relationship methods' postconditions without weakening the
// public API or adding a production testing service.
[assembly: InternalsVisibleTo("Tests")]
