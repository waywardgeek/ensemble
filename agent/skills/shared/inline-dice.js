// Presentation of recorded additive facts only: no random source or tool access.
export default function inlineDice(document, record) {
    if (!record || typeof record.label !== "string" || record.label.length > 256
        || !Number.isSafeInteger(record.total) || !Array.isArray(record.terms)
        || record.terms.length === 0 || record.terms.length > 32)
        throw new TypeError("Invalid recorded dice result");
    let sum = 0;
    const explanation = [];
    for (const term of record.terms) {
        if (!term || typeof term.label !== "string" || term.label.length > 256
            || !Number.isSafeInteger(term.value))
            throw new TypeError("Invalid recorded dice term");
        sum += term.value;
        if (!Number.isSafeInteger(sum)) throw new TypeError("Dice total overflow");
        explanation.push(`${term.label}: ${term.value}`);
    }
    if (sum !== record.total) throw new TypeError("Recorded dice total disagrees with terms");
    const disclosure = document.createElement("details");
    disclosure.className = "inline-dice";
    const summary = document.createElement("summary");
    summary.textContent = `${record.label}: ${record.total}`;
    const facts = document.createElement("span");
    facts.className = "inline-dice-explanation";
    facts.textContent = ` (${explanation.join(" + ")} = ${record.total})`;
    disclosure.append(summary, facts);
    return disclosure;
}
