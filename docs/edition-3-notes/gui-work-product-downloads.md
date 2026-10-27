# Work-product downloads

The browser receives a download offer only after the agent calls
`offer_download` successfully. The user must ask for the offer; the tool
description instructs the model when to use it, but cannot guarantee model
compliance. Files are reopened inside the agent workspace and sandbox when
downloaded, so the served bytes can change after the offer.

Codex exercised both the combined main tree and the independent feature branch
using the real GUI and jobs with a fake external model provider. Playwright
with Firefox downloaded and compared the file bytes in streaming and
non-streaming conversations, checked failure and agent isolation, and verified
reload and process-restart replay. These are not live-model results.

Targeted scratch mutations confirmed that the tests reject a failed job offered
as successful, a missing failure flag, reuse of a job handle for an old offer,
removal of replayed offers, and loss of a successful offer when its result report
fails before the completion event. Earlier mutations also checked token redaction,
deleted-agent rejection, workspace rechecking, exact bytes and sandbox escapes.
