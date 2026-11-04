---
name: work-products
type: loadable
description: Offer specific Sandbox work-product files for deliberate human downloads.
tools: offer_work_product
---
Use offer_work_product to make a completed report, image or other requested work
product downloadable. Explain what you are offering. A pathname in ordinary prose
is not a download offer. Loading this skill grants offering authority; file-editing
access alone does not. Each offer requires an explicit successful tool result.

The human's click starts transfer without a second approval dialog. No automatic
transfer or public hosting is provided. The server rechecks the Agent's Sandbox
and reads current file contents at click time, not a frozen copy. A changed or
removed file can change the download or make it unavailable. Transfers are bounded
to 64 MiB, separately from model image-input and Job-output limits.

Each Agent may create at most 64 offers during its lifetime. At the cap, additional
offers refuse rather than silently evict earlier ones. Offers expire on Agent
destruction or application restart. Retained History describes old offers but
does not restore their authority. Do not promise permanent links or immutable bytes.

A running tool observation is not a successful offer. Preserve its Job handle
and collect its completed result. Never infer offer metadata from prose.
