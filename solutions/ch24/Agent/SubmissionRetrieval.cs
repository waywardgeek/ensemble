using System.Text.Json;

namespace Ensemble;

public sealed partial class Agent
{
    public partial StoredSubmission RetrieveSubmission(string childID)
    {
        ValidateChildID(childID);
        string? address = Address();
        string directory;
        Sandbox files;
        lock (ingress)
        {
            ObjectDisposedException.ThrowIf(destroyed || destroying, this);
            Agent? child = children.FirstOrDefault(candidate => candidate.childID == childID);
            directory = child is null
                ? Path.Combine(DataDirectory(), "children", childID)
                : child.DataDirectory();
            files = Sandbox();
        }

        // Open once through the actual owner. Atomic replacement by the child
        // gives this reader one version, not a mixture of old and new bytes.
        // Check bytes while reading too: a length check alone is not a bound
        // when an external writer changes a file in place.
        using var stream = files.OpenFile(Path.Combine(directory, "submit.md"), FileMode.Open, FileAccess.Read);
        using var contents = new MemoryStream();
        byte[] buffer = new byte[8192];
        while (true)
        {
            int count = stream.Read(buffer, 0, buffer.Length);
            if (count == 0)
            {
                break;
            }
            if (contents.Length + count > SubmissionArtifactBytes)
            {
                throw new InvalidDataException("Submission artifact byte limit exceeded.");
            }
            contents.Write(buffer, 0, count);
        }
        try
        {
            // One envelope adds one depth level to the existing schema/payload
            // validator's depth-64 limit. Its validator remains authoritative.
            using var document = JsonDocument.Parse(contents.ToArray(), new JsonDocumentOptions { MaxDepth = 66 });
            JsonElement row = document.RootElement;
            if (row.ValueKind != JsonValueKind.Object)
            {
                throw new InvalidDataException("Submission artifact must be an object.");
            }
            var fields = new HashSet<string>(StringComparer.Ordinal);
            foreach (JsonProperty field in row.EnumerateObject())
            {
                if (!fields.Add(field.Name)
                    || field.Name is not ("version" or "agent_id" or "generation"
                        or "submit_schema" or "payload" or "parent_address"))
                {
                    throw new InvalidDataException("Submission artifact has duplicate or unknown fields.");
                }
            }
            if (!row.TryGetProperty("version", out JsonElement version)
                || !version.TryGetInt32(out int revision) || revision is not (1 or 2)
                || fields.Count != (revision == 1 ? 5 : 6)
                || (revision == 1 && fields.Contains("parent_address"))
                || !row.TryGetProperty("agent_id", out JsonElement id)
                || id.ValueKind != JsonValueKind.String || id.GetString() != childID
                || !row.TryGetProperty("generation", out JsonElement generation)
                || !generation.TryGetInt64(out long acceptedGeneration) || acceptedGeneration < 0
                || !row.TryGetProperty("submit_schema", out JsonElement schema)
                || !row.TryGetProperty("payload", out JsonElement payload))
            {
                throw new InvalidDataException("Submission artifact version, identity or fields are invalid.");
            }
            string? recordedParent = null;
            if (revision == 2)
            {
                JsonElement parentAddress = row.GetProperty("parent_address");
                if (parentAddress.ValueKind is not (JsonValueKind.String or JsonValueKind.Null))
                {
                    throw new InvalidDataException("Submission parent address must be a string or null.");
                }
                recordedParent = parentAddress.GetString();
                if (recordedParent != address)
                {
                    throw new InvalidDataException("Submission artifact belongs to another parent address.");
                }
            }
            var contract = new SubmissionContract(schema);
            JsonElement validated = contract.Validate(payload);
            return new StoredSubmission(childID, recordedParent, acceptedGeneration,
                contract.Schema(), validated);
        }
        catch (Exception error) when (error is JsonException or ArgumentException or InvalidOperationException)
        {
            // Do not echo corrupted file content or local paths into tool errors.
            throw new InvalidDataException("Submission artifact is malformed or violates its schema.", error);
        }
    }
}
