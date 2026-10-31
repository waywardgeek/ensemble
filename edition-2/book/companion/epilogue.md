# Epilogue: What the student changed

The first attempt to revise the book failed before the student could provide the intended lesson. An author was placed ahead of the student to improve each assignment. A small persistence exercise became a larger system: 254 physical lines of production Go in the original implementation, 4,585 in its replacement, excluding tests and measured against each version's predecessor. The reviewer approved the expansion. Bill stopped the attempt and preserved it for inspection.

The [original epilogue](../edition-3/epilogue.md) records the problem plainly: the review had checked whether the student built the revised assignment correctly, not whether the revision still served the original task. The failure wasn't that large programs are necessarily bad. The assignment had changed before its supposed improvement was tested.

The restart reversed that order. The student used the original book and graders, with access to its own preceding implementation but not the old answers. A reviewer compared implementations after the student had built and tried each chapter. Only then did the teaching change. That preserved a stable instruction against which a misunderstanding could be explained.

## Corrections with a cause

The tool-loop assignment told the student to record text and tool calls, then continue until the model stopped requesting tools. The student also concatenated intermediate narration into the answer returned to the caller. The unchanged grader rejected it. Returning only the final response's text raised the recorded score from 55 to 90 out of 100; a separate inherited media check still failed. The corrected teaching distinguishes what belongs in conversation history from what the caller receives.

The memory assignment contained a different error. It described a compression trigger using cumulative input tokens. A lifetime total cannot fall when context is compressed. Bill clarified that the trigger should use input from the preceding request, including cached input once. The authors checked the accounting contracts and original implementation, recorded the assistance, and corrected the explanation. Here the student's difficulty exposed a wrong instruction, not a wrong implementation.

A Gateway check exposed a wrong test. A deliberately defective implementation returned the expected refusal to a child's key-registration request while still registering the forbidden key. The revised grader initially passed it. Trying that key afterward revealed the unauthorized access. The stronger check rejected the same defect, and the initial passing result remained in the record to explain why the test had changed.

These are three different repairs: clarify the result contract, correct the measurement, observe the forbidden consequence. Treating all three as “the model made a mistake” would discard the knowledge needed to prevent their recurrence.

## Beyond the construction tests

Daily use added another kind of evidence. As [Chapter 25](chapter-25.md) records, the second-edition crossover exposed an old executable behind a changed speech control, stale JavaScript after a server restart, and a total HTTP deadline that killed legitimate responses. A source edit, a build, an installation and an observed running change were separate events. Construction tests hadn't eliminated those integration failures.

That crossover advances beyond the original epilogue's then-unfinished plan. It doesn't establish a recurring autonomous improvement cycle, or new Edition 3 execution results. Likewise, the local tests behind [Goals](chapter-27.md) establish timer, admission and display behavior, not that a real model will reliably choose the right next action. A reminder provides another opportunity to inspect unfinished work; it doesn't supply the missing judgment.

The durable result is more specific than a successful demonstration. The next implementation can distinguish an answer from its history, current context from lifetime usage, a refusal message from denied access, and saved instructions from evidence of completed work. The codebook makes those distinctions executable. This companion preserves the failures and reasons behind them, so the next technical lead can change the implementation without having to rediscover why they matter.
