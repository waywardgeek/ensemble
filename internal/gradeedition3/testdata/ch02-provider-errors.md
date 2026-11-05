# Chapter2 provider failure supplement

This standalone public checker runs the submitted CLI against two owned loopback
HTTP providers. It uses Python3 standard-library modules only, no package install,
provider credentials, candidate internals or candidate test assembly. It currently
requires a POSIX host for owned subprocess-group cleanup. No native GUI or audio.

Use an existing interpreter, an absolute executable launcher and a pre-existing
absolute writable evidence directory. Run in an isolated environment. For example:

```sh
/absolute/python3 /absolute/graders/ch02-provider-errors.py \
  /absolute/candidate-launcher --work-root /absolute/owned/evidence
```

Do not pass a source directory or DLL directly: supply an executable that invokes
the pinned SDK/DLL and preserves the synthetic LLM variables supplied by the
checker. The checker launches candidates with a fresh environment and owned
home/temp/CLI directories. It makes one request per case and does not install or
fetch anything. Each case has a20-second process deadline and bounded server
socket I/O. Cleanup joins the process and server thread; retained result files
remain available under a newly created ch2-provider-errors directory.

The JSON contract is `chapter2-provider-errors-v1`. Require exit0, passed:true
and all four named cases: openai-valid-response, openai-empty-candidates,
gemini-valid-response and gemini-empty-candidates. The valid controls must preserve
the randomized response answer and record success without error. Empty candidate
arrays must exit1 without an assistant answer, preserve human/request facts, and
record exactly one error_occurred after request_sent with no response_ended.
Raw malformed-response canaries must not appear in stdout, stderr or History.
All cases require one request with no retry. Full per-case checks and candidate
exits are recorded; a crash is not a controlled refusal.

Exit1 means failed behavioral checks; argument/setup failures are not evidence
of a candidate defect. The checker retains stdout/stderr and actual logs. These
are synthetic data, not permission to collect private conversations. It awards
no historical chapter points and is not a live-provider, crash-recovery, general
security or student-acceptance certificate.

Initial sensitivity evidence uses the unchanged submitted C# student candidate:
valid controls pass, while both empty arrays fail controlled-failure-exit-not-crash.
A separate disposable copy with explicit candidate-array validation passes all
four cases. That copy validates the checker, not a revised student submission or
adopted runtime. Fresh student reconstruction remains required after the teaching
repair. The frozen chapter endpoints and original100-point grader are unchanged.

A second negative control changes only the failed candidate's process wrapper to
return1 instead of propagating its abort. Valid controls still pass, while both
malformed cases fail one-error-fact-no-success-fact. Thus a controlled exit alone
cannot satisfy the checker. This is an executable-wrapper control, not a newly
compiled runtime mutation.
