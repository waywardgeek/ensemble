# Fresh Flint reviewer: useful discoveries, incomplete coverage

The [procedure](../setup/independent-review.md) was followed for one fresh confined
reviewer on the exact free Flint Chat route. The [unedited report](reviewer-01-report.md)
is a typed model submission, not a parent-written diagnosis. The reviewer received
nine frozen public/source files, not the author's self-review, behavioral grade,
parent findings or private grader/reference source. It activated ensemble-coding.
No diagnosis-specific intervention, code execution or second review was supplied.

The parent verified the saved authority and working policy. The reviewer submitted,
was shut down and joined, and the original spawn receipt was collected. All nine
input hashes matched after review. It claims complete reading; that is its coverage
statement, not an independent audit of every delivered excerpt. Some progress
observations contained clipped chapter dumps and were not treated as reviewed.

## Supported findings

- F1: Agent attaches at Agent/Agent.cs:8 before environment validation at10–15.
  A thrown validation exception leaves the parent retaining a partial Agent.
- F2: Common/Owners.cs:60 constructs a default-timeout HttpClient; Request has no
  caller cancellation parameter and header/body waits pass no token. The chapter
  explicitly requires no fixed generation deadline and caller cancellation.
- F3: Agent/Agent.cs:12–13 throws without a base URL, whereas the chapter describes
  an Anthropic default for chat. The defaulting defect is supported.
- F4: Agent/Agent.cs:72–75 returns the actual retained List as IReadOnlyList. That
  neither creates a stable snapshot nor prevents a caller casting back to List.
  The chapter expressly requires an immutable snapshot.

These are independent discoveries in this review, not recognition of supplied
answers. They do not prove general review reliability or repair ability.

## Qualifications and omissions

The report puts the HttpClient allocation at Engine/Engine.cs:60; its correct
location is Common/Owners.cs:60. Its exact model-diagnostic wording demand is not
clearly mandatory: the chapter describes the reference wording, while safe failure
before a request is the consequential behavior. The base-URL finding stands without
that wording demand. Its comment preference considers the functional file alone;
Common contains Engine's class/ownership documentation. Absence of comments in one
partial file does not alone establish a teaching-standard violation. The report
also ends with an unused placeholder JSON example; the actual typed report remains
valid but the placeholder contributes nothing.

Major omissions include constructor bodies in Common/Values.cs, Engine's separate
caller attachment, duplicate AddExchange calls in Engine and Application, and
missing root/error-path destruction. These were explicit review dimensions in the
primary. The source also defaults missing usage members to zero, despite the report
correctly quoting the chapter's distinction between missing accounting and measured
zero. The report's four findings are therefore not an exhaustive repair list.

## Lesson and next decision

Fresh eyes helped: Flint can produce concrete, useful counterexamples on fixed
source. But enumerating review instructions does not guarantee systematic coverage,
just as the author's self-review checklist did not guarantee correct construction.
The skill was informed by previous failures, so this single case is not a blind
skill-design benchmark or proof of causal improvement.

No skill retuning, student repair, second grade or Chapter2 launch followed this
result. A possible next step is one bounded Flint correction task with reviewer
findings and additional parent findings explicitly attributed, followed by concrete
checks. That would be an assisted repair experiment, not evidence that Flint alone
found every required correction. Bill decides that scope.
