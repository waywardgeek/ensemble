package grade

// PlantedFact is the string the fake server puts in its round-1 *assistant*
// reply. The whole memory check rests on it: a program that forwards only user
// turns, or that starts a fresh conversation each round, cannot possibly have
// this string in its round-4 request, because the program never typed it —
// the server did.
const PlantedFact = "TANAGER-4417"

// MemoryProbeRound is the 1-based round whose request is inspected for the
// planted fact.
const MemoryProbeRound = 4

// Round is one scripted exchange.
type Round struct {
	User  string // what the grader writes to the program's stdin
	Reply string // what the fake server returns for that round's API call
}

// Script is the fixed five-round conversation every submission is graded on.
// It is deliberately boring: nothing here tests the model's intelligence,
// because there is no model. Every check is structural.
var Script = []Round{
	{
		User:  "Hello! I'm starting a new project. Give it a codename and remember it.",
		Reply: "Welcome aboard. Your project's codename is " + PlantedFact + ", and I'll remember it.",
	},
	{
		User:  "Good. What is the capital of France?",
		Reply: "The capital of France is Paris.",
	},
	{
		User:  "And the fourth planet from the Sun?",
		Reply: "Mars is the fourth planet from the Sun.",
	},
	{
		// The probe. The answer is irrelevant to grading; what matters is
		// what the program had to send in order to ask it.
		User:  "Now, what codename did you give my project?",
		Reply: "The codename I gave your project is " + PlantedFact + ".",
	},
	{
		User:  "Thanks. Anything else I should know?",
		Reply: "That is everything for now. Good luck with " + PlantedFact + ".",
	},
}

// Replies extracts the scripted assistant replies in order.
func Replies() []string {
	out := make([]string, len(Script))
	for i, r := range Script {
		out[i] = r.Reply
	}
	return out
}
