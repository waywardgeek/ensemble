# The Right Way to Write Complex Code

## From C and C++ to Go, Rust, and C#: architecture for humans, and for autonomous AI implementation

**Bill Cox**  
**Working paper — publication draft**

### Abstract

Complex software needs a coherent data model: one authoritative representation of each fact, explicit relationships between objects, complete construction and destruction, and a shared vocabulary that algorithms can use without depending on one another's implementations. I used these principles in electronic-design-automation software for decades before language models could write code. They are rules for human programmers first and AI collaborators second.

This paper traces the experiences behind those rules and the ways programming languages help or obstruct them. My criticisms concern exposed memory representation in C, inheritance-centered modeling in C++, Go's coupling of methods to the packages that declare their receiver types, and Rust's machinery for mutable graphs with back-pointers and multiple parents. The issue is not whether a sufficiently determined programmer can implement a graph in each language. The issue is how directly the language expresses the architecture, and how much effort is spent accommodating the language instead of solving the application problem.

AI adds a practical distinction: supervised pair programming is not the same task as autonomous execution of an agentic codebook. I use Go successfully with continuous expert supervision. That does not establish its suitability for an AI working without that supervision. For the latter task, I now prefer C#: ordinary object references, garbage collection, compile-time partial declarations and implementations, and mature libraries let us express the intended architecture with less machinery. C# is not perfect. It is a pragmatic implementation choice for an architecture that came first.

*This is an engineering position paper based on my experience, architectural reasoning, and the project episode described below. It is not a controlled benchmark of programming languages or coding agents. Prepared with AI assistance from my earlier dictated account and subsequent clarifications; the first-person narrative is edited prose, not a verbatim transcript. Sources and evidence limits appear at the end.*

# Part I — The problems we keep teaching our languages

These are rules for human programmers first. I enforced these rules, and others beyond the scope of this document, for 25 years, with excellent results. They guided the design, implementation, and review of complex systems long before LLMs could write code. Applying them to an AI collaborator is a later use of an established engineering discipline, not the reason that discipline exists.

I start with four recurring architectural problems. I call them bugs because of the damage they do to complex programs, not because they are necessarily mistakes in a compiler implementation.

1. **The C bug: exposed memory representation.** The logical model of an object becomes entangled with its physical layout. C-compatible interfaces carry that representation into later languages, making it harder to change how data is stored without changing how the program accesses it.
2. **The C++ bug: treating inheritance as the answer to relationships.** A hierarchy of types is not a substitute for the graph of parents, children, and cooperating objects that actually makes up a complex system. Nor does inheritance naturally solve the problem of adding temporary, algorithm-specific data to shared objects.
3. **The “Go stupidity”: coupling methods to the package that declares the type.** Go prohibits import cycles while requiring methods to live in the package that declares their receiver's base type. This obstructs a shared vocabulary whose operations are implemented by separate functional modules. The workarounds change the architecture or add machinery that exists to accommodate the language.
4. **The “Rust stupidity”: making ordinary graph relationships an ownership problem.** Multiple parents and back-pointers are fundamental to my data model. I object to the ownership, borrowing, reference-counting, weak-reference, or handle machinery required to express and mutate those relationships—not to a supposed impossibility of implementing graphs in Rust.

These are my architectural criticisms. They are not claims that C cannot implement a structure of arrays, that every C++ program must use inheritance, or that a graph is impossible in Go or Rust. The problem is the design pressure: what is easy to express, what becomes conventional, and what programmers—and now language models—keep doing by default.

This document concerns **large, maintainable systems running on one computer**, such as an electronic-design-automation tool or Ensemble. It is not a prescription for distributed systems.

## 1. Where these habits came from

In the 1960s, memory was scarce and expensive. Programmers worked close to the machine, often in assembly. Modern cache hierarchies and the workloads that expose their costs were not the setting in which those early conventions developed.

C emerged around the turn of the following decade. Dennis Ritchie dates its development to 1969–1973, with its most creative period in 1972.[1] It gave systems programmers a practical way to move beyond machine-specific assembly while retaining control over memory and efficient execution. That was an enormous achievement.

But useful decisions can become permanent assumptions. A C structure has a concrete representation that other code can observe and depend on. Once that representation crosses a binary interface, changing it is no longer just an internal implementation decision. Compatibility spreads the convention across languages and generations of software.

That is why I sometimes call the C linker the root of all evil. The complaint is about the representation contract carried through C-compatible binary interfaces: a choice about how bytes are arranged becomes a constraint on future designs. It is not that the linker itself prevents every alternative layout. We implemented an alternative in C. We could do so because we had kept access to the data abstract.

At the same time, database research was teaching important lessons about representing information and avoiding redundancy. I took a central lesson from relational normalization, including third normal form: **do not keep independently maintained copies of the same fact.** That is the architectural lesson I apply here, rather than a formal definition of third normal form.

### Null, reclamation, and the missing reverse relationship

I have also used the phrase “Lisp bugs” for null-like values, memory reclamation, and structures that provide a way down from a parent without providing a way back from a child.

These are related concerns, but they are not the same problem. Nullable types can make absence explicit. Garbage collection can reclaim unreachable memory. Neither tells a child how to remove itself from both of its parents, or how to find the application’s logger.

The question that matters for this architecture is:

> Who tells every parent when a child is deleted?

If the answer is “some other part of the program will remember,” the data model is incomplete. We knew how to represent and maintain these relationships long before today’s languages and frameworks. A memory manager does not remove the need to do so.

I regard explicit treatment of nullable values as an important response to the null problem. It is not the main subject here. The harder issue for these systems is maintaining the relationships between objects throughout their useful lives.

## 2. What I learned building place-and-route software

Early in my career, I worked on algorithms for placing and routing designs in field-programmable gate arrays. At QuickLogic, we hired a programmer from Cadence who brought an architecture that served us extraordinarily well.

We called its shared data model the **database**. It was not a relational database server. It was the common vocabulary of the system: the objects that all the algorithms manipulated, the facts about them, and the relationships between them. Today I call that module `Common`.

If data had to survive a particular module’s execution, or had to be understood by more than one module, its representation belonged in that shared model. A router and a timing analyzer did not invent different authoritative versions of the same wire.

The most important document for a new programmer described those data structures. Once the programmer understood them, they understood the vocabulary of the whole system.

### Hub and spoke

Each functional module depended on the common definitions. Functional modules did not import and depend directly on one another. A top-level program—what we called `main`, whether it supplied a GUI or a command-line interface—orchestrated the modules.

The dependency structure was simple:

```text
main / application composition
    ├── placement module ──→ common
    ├── routing module ────→ common
    ├── timing module ─────→ common
    └── other modules ─────→ common
```

The arrows show **code dependencies**, not ownership relationships between runtime objects. `common` does not import the functional modules. The functional modules share a vocabulary without acquiring a web of package dependencies.

That separation let us concentrate on algorithms. From about 1990 to 1995, I could work principally on algorithms rather than repeatedly redesigning the underlying structures.

## 3. The data is a graph, not a tree

A graph gives a simple example. A node has incoming and outgoing edges. Each edge knows its source node and its destination node. Those are two distinct parent relationships. Constructing an edge takes both parents and establishes both memberships.

There is nothing exceptional about multiple parents. Complex systems contain them everywhere. **One authoritative representation of a fact does not mean one owner for every object.**

For a relationship-bearing object, the rule is:

> If A points to B as part of a relationship, B has the corresponding reverse relationship to A.

Do not soften this to “use back-pointers where appropriate.” Within the shared relationship model, the reverse relationship is part of the model, not an optional optimization. The explicit value-object distinction below determines which objects are governed by this rule.

### Construction and destruction are relationship operations

An edge constructor receives its source and destination nodes and joins the appropriate collections. Its destructor removes the edge from both collections.

A node destructor destroys its incoming and outgoing edges as required by the model, then removes the node from its owning graph. Shared traversal paths must not cause the same edge to be destroyed twice. Destruction must leave no parent retaining a relationship to an object that has been logically removed.

This is not a demand for a single-owner hierarchy. An object knows all of its parents. It detaches from all of them, and it destroys strongly owned children according to the meaning of those relationships.

The responsibility is divided deliberately:

- The child knows which parents must be notified.
- The child calls the appropriate declared operation on each parent.
- Each parent modifies its own relationship collection.

The child must not reach into arbitrary parent collections and edit them itself. Equally, it must not leave unrelated code to discover and repair those relationships later.

In a garbage-collected language such as Go or C#, logical destruction is distinct from reclaiming the object’s storage. Garbage collection can recover memory when objects become unreachable; it does not perform these application-level relationship operations for us. In this paper, “destructor” means an explicit logical destruction operation, not a C# finalizer.

### Back-pointers are also how code finds what it needs

Even an object that is never destroyed needs its parent relationships.

Suppose a function receives only an object and encounters a condition worth logging. How does it find the logger? It should be able to navigate through the object’s actual parent relationships to the application context and the relevant operation.

It should not need a copied logger field, a global lookup, a callback bundle, or a new context argument threaded through every intervening function merely because the relationships are missing.

I am not claiming that every global in every program is forbidden; I have used globals myself. But I require programmers, human or AI, to correct missing relationships rather than use globals as the default workaround for an object’s missing context.

**When legitimate access is difficult, inspect the data model and its declared operations first.** Adding plumbing to the caller can conceal a mistake several levels upstream.

## 4. The important exception: value objects

Not every reference represents a relationship that must be maintained in both directions.

I also use **value objects**: vectors, matrices, symbols, and similar values. I do not necessarily expect such a value to know every object that refers to it. In Lyric I can manage those values with reference counting; in Go or C# I can rely on garbage collection.

Every value-object type must be explicitly identified as such in a comment. The exception must be visible in the model, not inferred later because reverse relationships turned out to be inconvenient.

Some types need a deliberate decision. A compiler expression is a good example. Expressions form trees, but statements and other structures also refer to them. We can model expressions as relationship-bearing objects with back-pointers and explicit destruction, or explicitly choose value-object treatment with managed reclamation. The choice is not always obvious, but we must make it consistently.

The value-object exception does not excuse missing reverse relationships for the rest of the shared model. Nor does a type become a value object merely because the language happens to use garbage collection.

### Temporary local references

There is a separate, narrow case: a module may run an algorithm with a temporary cache containing pointers to objects that cannot be destroyed while that algorithm uses them. Those temporary references need not become relationships in `common`.

That is a local lifetime property of the algorithm. It is not permission to create a persistent shared relationship without its reverse link, or to keep a second authoritative copy of shared state.

## 5. Representation and the C bug

We represented many relationships with intrusive linked lists: the links lived in the child objects rather than in separately allocated list nodes. An object participating in different relationships had the link fields those relationships required. In my measurements on these systems, this was an efficient choice. The relevant evidence is the actual workload, not a blanket assumption that arrays or lists are always faster.

At QuickLogic, I recall roughly 40% of the memory being used to represent pointers in the common structures. As designs grew beyond the comfortable limits of a 32-bit address space, moving to 64-bit pointers increased that burden.

The programmer who introduced the architecture insisted that data access be abstract. Our C macros preserved that separation. It let us change the representation from **array of structures** to **structure of arrays** without making every algorithm depend on the new physical layout.

An array of structures stores the fields of one object together:

```text
object 0: field A, field B, field C
object 1: field A, field B, field C
```

A structure of arrays stores corresponding fields together:

```text
field A: object 0's A, object 1's A, ...
field B: object 0's B, object 1's B, ...
field C: object 0's C, object 1's C, ...
```

We also used 32-bit integer handles instead of 64-bit pointers where the object space allowed it. Across the tools we measured, I report an average improvement of about **10% in speed and 20% in memory use**. Smaller references reduced memory pressure, and an algorithm iterating over one field could load more relevant data into cache instead of fetching unrelated fields alongside it.

A more recent measurement appears in *The Lyric Book* and the Lyric repository. In the bootstrap-day results dated June 12, 2026, switching the Lyric compiler from AoS to SoA with `--soa` produced **10% faster execution and 14% less memory use, with zero source-code changes**. The README identifies the workload as the Lyric compiler compiling its own source on a **MacBook Air M2**.[3]

In that implementation, AoS uses pointer handles to a slab of contiguous objects; SoA uses 32-bit indices into parallel per-field arrays. The comparison therefore combines the effects of layout and handle representation. Field-wise traversal can fill cache lines with the field the algorithm needs, but this result does not isolate cache effects from the other changes.

These are measurements of particular workloads, not universal promises for every program. The older QuickLogic result and the newer Lyric result are separate measurements; the 20% and 14% memory reductions should not be conflated.

The lesson is not “rewrite every structure as SoA.” It is that the logical vocabulary should not unnecessarily lock every caller to one physical representation.

## 6. Inheritance is not a relationship model

C++ made inheritance a prominent way to organize behavior and types. But an edge belonging to two nodes is not a subclassing problem. Adding more levels to a type hierarchy does not establish the relationships that an edge needs to maintain.

There is another practical problem: algorithm-local extensions. A timing analyzer may need resistance, capacitance, or temporary calculations associated with wires. Those fields are not necessarily part of the shared wire vocabulary needed by the router and every other module.

Putting every algorithm’s temporary fields into the common wire type contaminates the shared model. Constructing a new inheritance hierarchy for each combination of algorithms is not the answer either.

With an AoS representation we used extension mechanisms such as opaque pointers. With a handle-based SoA representation, a module can keep local arrays indexed by the same object handles. The module owns the meaning and lifetime of those extra fields. Shared facts remain in the shared model; local scratch work remains local.

## 7. The “Go stupidity”: declarations and implementations belong in different places

I use Go in CodeRhapsody. This criticism is not coming from someone who has decided that nothing useful can be written in Go.

What I call the **“Go stupidity”** is a combination of two language restrictions:

1. Package imports cannot form cycles.
2. A method's receiver base type must be defined in the same package as the method.[6]

The second rule is often stated informally as “methods must be defined where the struct is declared.” The exact boundary is the package, not the file or Go module. Moving two files into the same package is allowed. It also gives up the package separation I am trying to express.

Here is what I want:

- `Common` declares the objects, their persistent data, their relationships, and the operations callable across module boundaries.
- Functional modules implement those operations.
- Functional modules depend on `Common`, not directly on one another.
- The top-level application constructs and connects the system.

If `Node` is declared in `Common`, I want the graph implementation to supply `Node.RemoveEdge`. It is still a method on the same object. There is no second node, adapter, or service object involved.

Go does not allow a separate package to add that method to the type declared in `Common`. The usual alternatives each sacrifice something:

- **Put the methods in Common.** Now the vocabulary module accumulates the algorithms it was meant to declare, not implement.
- **Replace methods with functions in implementation packages.** Direct calls across functional packages introduce implementation dependencies. In a sufficiently interconnected system, some arrangements also run into the prohibition on import cycles.
- **Introduce interfaces, implementation objects, factories, and forwarding.** The dependency graph can remain acyclic, but now we have runtime machinery whose purpose is to connect a declaration to its implementation.
- **Put everything in one Go package.** Files can be organized into roles, but Go package boundaries no longer express those roles as separately located functional source modules. That can be a valid choice for another project; it does not satisfy the separation I am asking for.

The objection is not that any one of these techniques is always wrong. The objection is that I have to choose one merely to organize declarations and implementations the way I intend.

### I am not asking for circular imports

This point is easy to miss. I do not want a web of circular package dependencies. My required dependency structure is deliberately hub and spoke.

The graph is in the **data**. A node knows its edges. An edge knows its endpoints. Objects have parents, and some objects have several parents. The declarations describe all of that in one shared vocabulary.

Separately, the **code** is organized into functional implementations that depend on those declarations.

I want to separate declarations from implementations without creating circular imports in the first place. Treating my request as a desire to disable dependency discipline misses the architecture.

### The interface workaround is not free merely because it is idiomatic

Our Go approach used interfaces declared in `common` and implementations in functional packages. It is possible to build real systems that way. I have done so.

But the model now has more to explain. Which concrete object owns the data? Which object carries the operation implementation? Is an interface an actual application abstraction, or merely the bridge between a declaration and a body? Is there one authoritative fact, or has state been copied into the implementation object? Does cleanup follow the real relationships or some additional service wiring?

A skilled programmer can answer those questions and enforce a coherent result. The point is that the questions were introduced by the workaround.

When a language forces additional machinery into the model, “that is idiomatic” is not a defense. I am judging the idiom against the requirements, not the requirements against the idiom.

## 8. The “Rust stupidity”: ordinary relationships become a special ownership problem

What I call the **“Rust stupidity”** is the difficulty of representing the mutable relationships that are ordinary in my applications.

My objects have back-pointers. Some have multiple parents. Deleting an object requires removing it from all those parents. An algorithm may need to navigate from an edge to either endpoint, from a child to an owning application, or from one node of a doubly linked list to its predecessor and successor.

These are not exceptional relationships. They are the model.

Rust's ownership and borrowing rules fit some representations naturally. A uniquely owned hierarchy with deletion cascading through its owned children is a particularly straightforward case. My objection is what happens when the application is not that hierarchy.

### There are two different graphs

The application's relationship graph answers questions such as:

- Which nodes are endpoints of this edge?
- Which objects contain this child?
- Where is the next list node, and where is the previous one?
- Which parent provides access to the application and its logger?

The memory-ownership structure answers different questions:

- What keeps this allocation alive?
- Who may mutate it at this point?
- What event permits its storage to be reclaimed?

I want the application's relationships to be direct and complete. I do not want to distort them into a single-owner tree merely to make the memory-ownership structure convenient.

Rust can represent graphs. That is not the dispute. The dispute is how much machinery and restriction it places between the programmer and the graph.

### The workarounds are the point of the criticism

Reference-counted pointers are one approach. For mutable shared data, examples often combine `Rc` with `RefCell`; other situations need other synchronization or ownership choices. But a cycle of strong reference-counted pointers keeps itself alive. The reference counts do not reach zero merely because the application no longer has an external path to the cycle.

Rust's own documentation demonstrates reference cycles and uses `Weak` references for a parent link that should not keep the parent alive.[7] That is a legitimate memory-management technique. It also means the programmer must distinguish ownership links from non-owning links, create and upgrade weak references, handle the possibility that the referent is gone, and obey the mutation rules of the chosen representation.

Arenas and handles offer another approach. They can avoid reference-counted ownership cycles. They also change access into lookup through a storage mechanism, and may require generations or other validity rules if objects can be removed and slots reused. A borrowed-reference design may work when the lifetimes and mutation patterns permit it. Unsafe code can implement additional representations, with the burden of proving their safety.

Listing these possibilities does not answer my objection. It describes the choices the language makes me manage.

My criticism is not a universal claim that reference counting is always slow, arenas are always bad, or every Rust graph needs the same representation. It is this:

> **A graph or a doubly linked list should be an ordinary data model. I should not have to turn its relationships into a separate ownership engineering project before I can work on the application.**

### Lifetimes do not perform logical destruction

Compile-time lifetime checking can prevent classes of invalid memory access. It does not remove a logically deleted edge from both endpoint collections. Reference counting does not do that either. Neither does garbage collection in C#.

Those are application operations. They have to be declared, implemented, and reviewed as such.

Even after choosing an acceptable Rust storage representation, I still need the complete relationship operations described in this paper. I have not replaced that work; I have added another set of constraints around it.

A programmer who chooses Rust for a concrete requirement may accept that tradeoff. For the systems and autonomous workflow discussed here, I do not want to pay it merely to express back-pointers and multiple parents.

## 9. What changes when the AI works without me

I use Go in CodeRhapsody, and I supervise the changes. I listen to Claude's reasoning throughout the work. When it starts drifting away from the architecture, I catch it and correct it.

That supervision is not an incidental detail. It is part of the development process that makes the result work.

Executing an **agentic codebook** is a different task. The codebook is a sequence of implementation instructions intended to let an AI build a working system chapter by chapter. The AI must carry the architecture forward while adding behavior, testing it, and maintaining earlier invariants. I am not supposed to be standing beside it correcting every local design choice.

A language that works with continuous expert correction is not automatically a good choice when that correction is absent.

### What actually happened to the Go chapters

We discarded 21 chapters of Go work after Astra spent approximately 30 hours unable to complete Chapter 21. The code was badly designed. The assistant kept insisting on idiomatic Go instead of consistently implementing the architecture I required.

This was not a case of me rejecting a successful implementation because I preferred a different style. The work did not reach a usable completion of the chapter.

I also found a bug in how I had configured the instructions: the assistant had seen my rules as file content, but they had not been installed in its system prompt as governing instructions. I had supplied the rules without giving them the intended authority in that setup.

That distinction matters. My diagnosis was not simply “Go cannot do this.” The assistant was failing to preserve the architecture, and the instruction setup was failing to establish the rules as the authority over its familiar defaults.

Calling the setup mistake mine does not make the produced code good. It explains one part of what went wrong and one part of what we had to change.

### Reading a rule is not the same as making it govern the work

An agent can read a document and still treat it as background material rather than as the constraints that determine whether an implementation is acceptable. In our setup, merely presenting the file was not equivalent to activating those rules in the instruction mechanism we intended to use.

The operational lesson is to make the architecture governing instructions, not merely discoverable documentation. It must remain authoritative across chapters, delegated work, and resumed sessions. A child agent that is doing implementation needs the applicable rules too; its parent cannot assume that a file somewhere in the repository will have the intended effect.

This is an instruction-delivery requirement, not a claim that a system prompt makes an AI infallible. We still have to inspect the produced code and test behavior. But the assistant must not substitute “idiomatic Go” for an explicit architectural requirement and regard that substitution as progress.

### Language fit still matters after the instruction bug is fixed

Fixing the instruction setup does not remove the language restrictions described earlier.

If following the architecture requires a special arrangement of interfaces, wrappers, and conventions, the autonomous implementer has to preserve that arrangement through every change. Familiar alternatives remain available, and locally convenient deviations can accumulate into a different system.

A language that expresses the intended structure directly reduces the number of such decisions. It does not guarantee success. It removes avoidable opportunities for failure.

We changed both the instruction setup and the implementation language. This episode does not isolate their individual contributions. It does establish the practical failure we experienced and the reasoning behind the changes we chose.

The question I care about is therefore more specific than “Which language is best?”:

> **Which language lets an AI execute this architecture autonomously, with the least need for expert intervention to prevent architectural drift?**

## 10. Why I now prefer C# for this work

C# is not a perfect language. I am not asking it to be one.

I want a language familiar to coding assistants, with mature libraries for the work we need to do, that does not put unnecessary barriers in the way of this architecture. C# is a pragmatic fit.

### Ordinary references express the actual relationships

A C# object can refer directly to its parents. An edge can have references to both endpoints. A child can reach the application through its actual ownership relationships. Multiple parents and cycles do not require a new storage or ownership representation just to exist.

The garbage collector can reclaim unreachable cycles. It does not require us to turn every back-pointer into a reference-counting exception.

I still require explicit logical destruction. A reachable parent retaining a deleted child is a model bug even if every memory access is legal. The garbage collector does not know whether an edge belongs in a node's collection or whether a wire should still be attached to a pin.

That is the division of responsibility I want: the program maintains the meaning of its relationships; the runtime reclaims unreachable managed memory.

When I describe a relationship as weak, I mean that destroying the parent detaches that relationship without cascade-deleting the child. That does not automatically call for a C# `WeakReference<T>`. These are application-level ownership rules, not a disguised request for a particular memory-management API.

### The compiler connects declarations to implementations

C# partial types let us put parts of the same type in different source files. Partial members let one part declare an operation and another part implement it. The compiler combines the parts; it does not create a wrapper object or use runtime dispatch to connect them.[4]

C# 14 adds partial constructors, allowing constructors to follow this separation too.[5] For this implementation, I use C# 14 on .NET 10, or a later compatible version.

The arrangement is direct:

```text
Common/
    Shared type declarations
    Persistent fields and relationships
    Partial operation and constructor declarations

Graph/
    Matching partial types
    Graph operation and constructor implementations

OtherFunctionalModules/
    Implementations of their declared operations

Application/
    Construction, connection, and orchestration
```

For example, Common declares the partial `Node` type, its fields, and a constructor signature such as `public partial Node(Graph parent);`. The graph module supplies the matching constructor body in another part of the same type. That body establishes the relationship through the parent's declared add operation.

There is one `Node`, one set of fields, and one implementation of each declared operation. We do not pass a second operations interface into the node merely because its constructor body lives in another source directory.

### Source modules, not separate assemblies

This mechanism has a clear boundary: all parts must compile into the same assembly and compiled module.[4] A directory is not a separate DLL. We are organizing functional source modules within one compiled system.

That is an acceptable tradeoff for this architecture. I am not trying to distribute each functional module as an independently compiled extension to an existing class.

It also means the compiler does not enforce every directory-level dependency rule for us. A shared namespace makes the Common vocabulary accessible; it does not authorize one implementation module to depend on another module's implementation-only helpers. We still enforce the architectural rule in the instructions and in review.

### Mandatory declarations, not optional hooks

The callable partial operations have explicit accessibility. Missing an implementation must be a compile error, not an optional hook whose calls can disappear.[4]

Common is not an excuse to accumulate algorithms. It declares the types, data, relationships, and cross-module operations. Functional modules supply the bodies. Required operations are concrete and direct unless an actual requirement calls for something else.

This is what I wanted from the beginning: a shared vocabulary without a shared dumping ground for implementations.

### Libraries and familiarity matter

The application also needs ordinary infrastructure: HTTP, WebSockets, JSON, processes, cancellation, asynchronous I/O, and concurrency. The .NET libraries provide established facilities for this work.[8][9]

C# also has a substantial body of documentation and existing code for language models to work from. I am not proposing an obscure language that requires teaching the assistant the ecosystem before it can implement the application.

Familiarity is a practical reason to choose it, not a measured guarantee that every model will produce correct C#. The governing architecture and inspection of the result still matter.

### What I am not claiming

Choosing C# does not give us Lyric's generated structure-of-arrays layout or its reported benchmark results. Ordinary C# reference objects have their own runtime representation and costs. A data-layout requirement has to be addressed explicitly; it does not disappear because the architectural separation is cleaner.

Nor does a successful C# build establish that the application is correct, that the AI preserved every relationship, or that a real external provider works. Those require the appropriate inspection and tests.

My claim is simpler:

> **I am choosing C# because it lets me write the architecture I intend, has the libraries the application needs, and is familiar to the AI doing the implementation. I can spend my effort teaching the AI how the system should work instead of teaching it workarounds for its implementation language.**

The architecture came first. C# is the current implementation choice that accommodates it, not the source of the architecture.

## 11. Simplification is not optional

The architecture exists to make correct programs easier to understand and maintain. It is not permission to add a framework around every operation.

When I refactor code, I remove unnecessary state, eliminate duplicated authority, and make operations direct. Reducing line count may be a consequence. It is not the definition of improvement.

In EDA work, a large reduction in code size often did accompany a faster program. Removing wrappers and duplicated mechanisms can also remove allocations and work. But an experiment in NetworkManager showed the limit of that intuition: roughly halving the code produced essentially no runtime improvement. The work was dominated by external operations such as Wi-Fi scanning. Simpler code was still a maintenance improvement, but it did not shorten the operation that mattered to the user.

The rule is not “shorter code is always faster.” It is:

> Remove what is not needed. Make the remaining design direct. Measure the behavior that matters.

Tests are necessary, but a passing test suite does not excuse a structure that violates the requirements. A wire that still points to a logically destroyed endpoint is wrong even if no current test happens to traverse that link. An implementation full of forwarding layers is not made necessary by the fact that a test accepts its output.

Conversely, a design that looks elegant on a diagram must still work. We test real components, inspect meaningful state, and exercise the failure and cleanup paths that the requirements make important.

## 12. The human remains the codebase expert

Autonomous execution does not mean that the assistant gets to invent the architecture.

The human agrees to the shared vocabulary, the relationships, and the responsibilities. Within that agreement, the assistant should be able to carry out routine implementation without asking permission for every function. A genuine change to the model or architecture is different. That is a point for discussion, not an opportunity to hide a redesign inside a passing patch.

For a codebook, the declarations are also part of the teaching material. A student should be able to identify the authoritative facts, follow an object's path to its parents and services, and understand what construction and destruction do. The student should not have to reverse-engineer a collection of adapters to discover the actual objects.

The practical rules that follow make those expectations explicit.

# Part II — The rules I require

These are requirements I established for human programmers and now apply to AI collaborators. They describe the intended architecture, not optional suggestions to be replaced with a language's fashionable idioms.

The human remains the codebase expert. Changes to the shared vocabulary and architecture require agreement. Routine implementation within that agreement does not require per-function permission.

## 1. Use a hub-and-spoke implementation

One **Common** module declares the shared vocabulary: the core data types, their relationships, and every operation callable from another module. Common is a leaf and imports no functional module.

Functional modules use Common, not one another's implementations. They implement the operations declared in Common. The top-level application may use any module to construct and connect the system.

Do not turn Common into a collection of algorithm implementations. Do not conceal a functional dependency with a fully qualified name or a global import.

## 2. Keep each fact in one place

Do not duplicate authoritative data to make it easier to reach. A field, collection, or state variable should have an identifiable source of authority.

This does not mean an object has only one owner. An object may have multiple parents and owners. Represent the actual relationships rather than forcing the model into a tree.

Complete forward and reverse links are two maintained directions of one relationship, not permission for independently maintained, contradictory versions of a fact. The relationship-changing operation is responsible for keeping them consistent.

## 3. Maintain both directions of every relationship

Relationship-bearing objects retain back-references to all their parents. Each relationship is represented consistently at both ends.

Change a parent's relationship collection by calling its declared methods, not by editing that collection from outside. A parent's add method adds the child and establishes the child's back-reference. Its remove method removes the child and clears the corresponding back-reference.

Each operation updates both ends before returning. Validate a proposed change before modifying either end, so a refused operation leaves both unchanged.

Do not split one relationship change between a parent method and cleanup the caller must remember to perform. Do not maintain the forward link in one module and hope another module repairs the reverse link later.

## 4. Construct objects with their strong owning parents

Constructors take references to the parents that strongly own the new object. They establish those relationships through the parents' add methods, which update both directions.

An edge in a directed graph, for example, takes its `fromNode` and `toNode`. The constructor's contract explains the memberships and back-references that exist when construction succeeds.

In C#, use ordinary object references, not unsafe pointers. Do not add separate operations interfaces, callback bags, or copied services to the constructor when the parents already provide access to them.

Construction involving more than one relationship must not leave an abandoned half-registered object after failure. Make validation and cleanup understandable in the implementation rather than delegating that responsibility to an undocumented caller convention.

## 5. Reach data and services through the relationships

Required data and operations must be straightforward to reach from the object that needs them, including logging code. Traverse the parent relationships. Declare the necessary access paths and operations in Common.

If access is awkward, correct the model or its declared operations. Do not compensate with duplicated fields, global searches, reconstructed identifiers, runtime type discovery, or callback plumbing.

This does not mean that every function needs a tour through the whole application. It means that a relationship-bearing object has the context its role requires, rather than depending on unrelated callers to rebuild that context each time.

## 6. Make logical destruction explicit and complete

Destroying an object must:

1. Remove it from **every parent relationship** by calling the parents' remove methods. Those methods clear the corresponding child back-references; the destructor does not duplicate that work.
2. Recursively destroy its **strongly owned children**.
3. Remove its **weak child relationships** without destroying those children.
4. Use complete two-way relationship operations for every detach, and avoid destroying a multiply owned child more than once.

Strong ownership means deletion cascades. A weak relationship is detached without cascade-deleting the related object.

For example, destroying a wire detaches it from its pins; it does not destroy a pin owned by a package. The pin remains, floating. A partial detach that removes the pin from the wire but leaves the wire in the pin's collection is not complete destruction.

These are model relationships, not instructions to use `WeakReference<T>`. C# garbage collection reclaims unreachable managed memory. It does not perform this logical destruction. Do not rely on finalizers to maintain the model.

Logical cleanup must also respect active work and external resources where the application has them. A task still using a child is not made safe by clearing the child's parent reference. Explain the ordering needed by the actual component rather than hiding it behind the word “dispose.”

## 7. Declare value-like classes explicitly

Value-like classes, such as vectors, symbols, and matrices, do not require back-pointers to their holders. Rely on C# garbage collection to reclaim their memory when they are no longer reachable. Do not add parent tracking and destruction machinery solely to reclaim such values.

Declare a type's value-like role explicitly. It may be implemented as a class; this is a modeling distinction, not a requirement to use a C# value type.

Do not use this exception to disguise relationship-bearing objects or duplicate authoritative mutable state. Short-lived algorithmic references are not automatically persistent model relationships, but persistent shared state does not become temporary merely because tracking its ownership is inconvenient.

## 8. Map the rules directly to C#

Use **C# 14 on .NET 10**, or a later compatible version, so constructor declarations can be separated from their implementations too.

### Common owns declarations

Put shared class and struct declarations, all their persistent fields and relationships, and every operation callable from another module in Common source files.

Declare callable methods and constructors as partial members with explicit accessibility and a semicolon, not a body. Do not put algorithms or accessor bodies in Common.

### Functional modules own implementations

Put matching partial type sections and member bodies in the responsible functional module directories. Each declared partial member has exactly one implementation.

Keep temporary algorithm variables local. Do not hide additional persistent object state in implementation files.

### Match the type identity

All parts use the same namespace, type name, and generic parameters. Mark each part `partial` and keep its accessibility and other required modifiers consistent. Declaration and implementation signatures must match.

These are parts of one type, not different types that happen to have the same name.

### Compile the parts together

Use one assembly and compiled module for these source modules. A directory is not a separate DLL or Go-style package. A partial type cannot be extended from a separately compiled implementation assembly.

### Keep required operations mandatory

Use explicitly accessible partial methods so a missing implementation is a compile error. Do not substitute optional private partial hooks whose calls can disappear when no implementation is supplied.

### Preserve the dependency rule

Implementations use the vocabulary declared in Common, not another functional module's implementation-only types or helpers. A shared namespace does not remove this restriction.

### Do not add runtime glue

The compiler combines the parts into one type and one set of fields. No interface field, wrapper object, delegate, factory, or type conversion is needed to connect a declaration to its body.

Use ordinary concrete types and direct, non-virtual methods unless an actual requirement needs otherwise. Do not model containment with inheritance.

## 9. Apply KISS and YAGNI

**KISS — Keep It Simple.** Implement the required behavior with the simplest clear design that respects these rules. Prefer direct operations on known objects and reuse suitable existing code. Remove unnecessary indirection, duplicated mechanisms, and layers that merely pass data along.

**YAGNI — You Aren't Gonna Need It.** Implement current requirements, not imagined future ones. Do not add speculative frameworks, registries, configuration options, extension points, or general-purpose abstractions. Add a mechanism when a concrete requirement justifies it.

Simplicity does not excuse missing parent relationships, incomplete destruction, duplicated facts, or violations of the hub-and-spoke structure. Those are requirements, not optional machinery.

Preserve useful explanations of non-obvious decisions. Fewer lines alone do not establish a simpler design.

## 10. Use fakes, not mocks

Test real components and the behavior they actually implement.

When an external dependency must be replaced, use a small working fake that implements the relevant contract, including the protocol and failure behavior needed by the test. Do not substitute mocks for the implementation being tested or use scripted internal-call expectations to manufacture a pass.

Assert meaningful behavior, state, and results. Relationship tests should verify both ends. Destruction tests should verify what remains as well as what disappears. Failure tests should check that refused mutations leave the model unchanged.

Keep fakes simple and faithful to the boundary they replace. Passing against a fake does not establish that a real external provider or service works.

## 11. Write exemplary teaching code

Clarity, comments, and readable layout are requirements. A student should be able to follow the model and explain why the program works, not merely reproduce its output.

### Explain vocabulary and contracts

Document each module's responsibility and each shared type's role. Describe what callable operations do, their important preconditions, results, failure behavior, and state changes.

Constructors and relationship methods explain ownership and their complete two-way postconditions. State explicitly when a type is value-like.

### Explain reasons where they matter

Comment non-obvious algorithms, protocol conversions, ordering, resource lifetime, and deliberately rejected alternatives. Explain why a choice preserves a requirement.

Do not substitute narration of obvious syntax or a distant README for an explanation needed beside the code. Tests should explain the behavior or failure they prove.

### Prefer readable steps

Put each substantive operation on its own line. Split non-trivial assignments and transformations into meaningful intermediate steps with names that reveal their purpose.

Do not chain assignments, pack multiple statements onto one line, or hide validation, mutation, and cleanup inside a compact expression.

### Expand control flow

Use braces and readable multi-line bodies for conditionals, loops, constructors, and error or cleanup paths. Keep their order visible.

A short expression-bodied member is appropriate only when the operation is genuinely obvious, such as returning one field. It must not conceal effects or an algorithm.

### Lay out data and calls for reading

Give shared fields their own lines and useful explanations of their meaning and ownership. Break long parameter lists, object initializers, and JSON construction across logical lines.

Do not extract pass-through helpers or invent abstractions merely to make a method shorter.

### Review as a teacher

Before calling a chapter complete, read its declarations and main execution paths as a new student would. Can the student identify the authoritative facts, follow the parent paths, see both ends of each relationship change, and understand failures and cleanup without guessing?

Correct missing explanations and dense code before advancing. These standards apply to examples and tests, not just production code.

## 12. Inspect the actual work

Read the code before judging it. A generated summary is not the implementation. A passing build is not proof of correct relationships. A test result is evidence only for the behavior the test actually exercises.

For autonomous work, inspect the model as well as the output. Did persistent state remain in Common? Did a child acquire another service field instead of using its parents? Did an implementation module start depending on another implementation module? Did destruction become a sequence of half-detaches? Did an interface appear because the application needed one, or because the assistant drifted back to a familiar pattern?

When something fails, distinguish setup failures, product defects, and limitations of the test. Preserve useful failure evidence. Do not respond to a failing architectural requirement by quietly redefining success.

This is how the rules stay operational rather than becoming a preface nobody follows.

# Conclusion — The architecture comes first

I do not need a programming language to persuade me that my data is a tree when it is a graph. I do not need runtime adapters merely to put a declaration in one source module and its implementation in another. I do not need an AI to replace requirements with familiar idioms and call that good engineering.

I need the objects, facts, relationships, and operations to be explicit. I need each relationship change to be complete. I need the implementations to remain understandable as the system grows.

Those needs existed when I was building place-and-route software. AI has not made them obsolete. Autonomous implementation makes adherence to them more important, because the expert is no longer correcting every local choice as it happens.

I use Go with close supervision. I understand that graphs can be implemented in Rust. Neither observation answers the question I am asking for an autonomous codebook.

For this work, I prefer C#. It lets us express the intended architecture directly enough that the language is not the main obstacle. Its libraries cover the infrastructure, its object references fit the model, and its partial declarations and implementations remove machinery that the architecture never needed.

**C# is not the philosophy. It is the tool I currently choose to implement the philosophy.**

The philosophy is older and simpler: understand the model, keep each fact in one place, represent every relationship completely, and write code that another person can understand.

# Sources and scope of the claims

## Historical account and project experience

The historical narrative is drawn from my earlier dictated account, the preserved draft of *The Right Way to Write Complex Code*, and the clarifications used to prepare this revision. Accounts of QuickLogic, the 64-bit transition, the later timing implementation, NetworkManager, supervised CodeRhapsody development, and the discarded Go chapters are my recollections and project reports. They are not presented as independently audited measurements.

The original source set included `The_Right_Way_to_Write_Complex_Code_Backup.docx`, `The_Right_Way_to_Write_Complex_Code.docx`, and `Your_Messages_Only_Conversation.docx`, followed by the Markdown reconstruction. One supplied Word export was incomplete. This paper does not invent the missing portion.

The language documentation supports the technical constraints described here. It does not endorse my architectural judgments. “Go stupidity” and “Rust stupidity” are my terms for those judgments.

## Primary technical sources

[1] Dennis M. Ritchie, **“The Development of the C Language.”** Source for the 1969–1973 development interval and the identification of 1972 as its most creative period.  
<https://www.nokia.com/bell-labs/about/dennis-m-ritchie/chist.html>

[2] Bill Cox / Wayward Geek, **Lyric repository and README.** Source for the documented generated relationships, handle representation, AoS/SoA options, and project-reported benchmarks.  
<https://github.com/waywardgeek/lyric>

[3] Lyric README, **“Where We Came From.”** Source for the quoted comparison between the earlier 64-bit transition and the reported 1997 timing implementation. This is a project account, not an independent replication.  
<https://github.com/waywardgeek/lyric#where-we-came-from>

[4] Microsoft, **“Partial Classes and Members (C# Programming Guide).”** Source for compile-time combination of partial types, the same-assembly and same-module requirement, and mandatory implementation of partial methods with explicit accessibility.  
<https://learn.microsoft.com/en-us/dotnet/csharp/programming-guide/classes-and-structs/partial-classes-and-methods>

[5] Microsoft, **“What's new in C# 14.”** Source for partial constructors and the C# 14/.NET 10 pairing.  
<https://learn.microsoft.com/en-us/dotnet/csharp/whats-new/csharp-14>

[6] **The Go Programming Language Specification**, “Method declarations” and “Import declarations.” Source for receiver base types belonging to the method's package and the prohibition on direct or indirect self-import.  
<https://go.dev/ref/spec#Method_declarations>  
<https://go.dev/ref/spec#Import_declarations>

[7] **The Rust Programming Language**, “Reference Cycles Can Leak Memory.” Source for strong reference-counting cycles and the example using a weak parent reference.  
<https://doc.rust-lang.org/book/ch15-06-reference-cycles.html>

[8] Microsoft, **`System.Net.Http` namespace.** Standard library HTTP facilities.  
<https://learn.microsoft.com/en-us/dotnet/api/system.net.http>

[9] Microsoft, **`System.Net.WebSockets` namespace.** Standard library WebSocket facilities.  
<https://learn.microsoft.com/en-us/dotnet/api/system.net.websockets>

## What would require additional evidence

This paper does not establish a universal performance ranking of C, C++, Go, Rust, C#, or TypeScript. It does not establish a comparative capability ranking of commercial coding agents from their source-code sizes.

A controlled claim about autonomous reliability would require comparable tasks, documented instruction delivery, defined completion criteria, and records of human intervention. Our Chapter 21 experience explains why we changed course; it does not substitute for that experiment.

The claim advanced here is architectural and practical: for this model and this workflow, direct language support for the intended design removes avoidable machinery and reduces what the autonomous implementer has to keep correct. The C# choice follows from that reasoning and remains subject to the quality of the actual implementation.
