namespace Ensemble;

public sealed partial class Agent
{
    private partial void ValidateAdditionalImages(IEnumerable<Part> parts)
    {
        static bool HasImage(Part part) => part is BlobPart blob
            ? blob.Mime.StartsWith("image/", StringComparison.Ordinal)
            : part is ToolResultPart result && result.Parts.Any(HasImage);
        var proposed = parts.ToArray();
        // Text and recovery controls must remain admissible even when restored
        // context is already over budget. Never remove existing evidence here.
        if (!proposed.Any(HasImage)) return;
        var queued = requests.Where(request => request.Kind == MailKind.Submit && !request.Parts.IsDefaultOrEmpty)
            .SelectMany(request => request.Parts);
        try
        {
            Engine().ValidateImageBudget(History().View(), queued.Concat(proposed));
        }
        catch (Exception error) when (error is IOException or InvalidDataException
            or NotSupportedException or FormatException)
        {
            // Raw library submissions may contain unresolved or malformed media.
            // Preserve their established recorded-rendering refusal rather than
            // faulting the mailbox during admission. This is not a budget waiver:
            // rendering still validates capabilities, reads and hard limits.
        }
    }

    private partial ToolResultPart AdmitToolImages(ToolResultPart result)
    {
        try
        {
            ValidateAdditionalImages(result.Parts);
            return result;
        }
        catch (ImageBudgetExceededException error)
        {
            // The Job keeps its original output. Only this delivery is refused;
            // paired History must not claim that the model received the pixels.
            return new ToolResultPart(result.CallID,
                [new TextPart("Image delivery refused; no new images were delivered. " + error.Message)],
                isError: true, job: result.Job);
        }
    }
}
