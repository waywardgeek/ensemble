using System.Runtime.CompilerServices;

// Isolated lease lifecycle fixtures; public GUI regression consumers stay public.
[assembly: InternalsVisibleTo("VoiceLeaseTests")]
