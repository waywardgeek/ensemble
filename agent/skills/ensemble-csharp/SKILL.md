---
name: ensemble-csharp
type: loadable
description: Bill's rules for complex C# systems: hub-and-spoke modules, a common vocabulary, multiple parents, complete relationships, and explicit destruction.
---
# Rules for complex C# systems

These are requirements Bill has enforced for human programmers for many years. They apply to LLM collaborators too. The human remains the codebase expert and an active pair programmer. Load this skill as instructions before coding; reading its file is not activation.

## 1. Use a hub-and-spoke implementation

One Common module declares the shared vocabulary: the core data types, their relationships, and every function callable from another module. Common is a leaf and imports no other module.

Functional modules include Common, not one another. They implement the operations declared in Common. The top-level orchestration module may import any module to construct and connect the system. Do not turn Common into a collection of algorithm implementations.

## 2. Keep each fact in one place

Do not duplicate authoritative data to make it easier to reach. This does not mean an object has only one owner: an object may have multiple parents and owners. Represent the actual relationships rather than forcing the model into a tree.

## 3. Maintain both directions of every relationship

Relationship-bearing objects retain back-references to all their parents. Each relationship must be represented consistently at both ends. Change a parent's relationship collection by calling its declared methods, not by editing the collection from outside.

Each relationship-changing operation must update both ends before it returns. A parent's add method adds the child and establishes the child's back-reference; its remove method removes the child and clears that back-reference. Validate the change before modifying either end, so a refused operation leaves both unchanged. Do not split one relationship change between a parent method and cleanup that its caller must remember to perform.

## 4. Construct objects with their strong owning parents

Constructors take references to the parents that strongly own the new object and establish the relationships through those parents' add methods, which update both directions. For example, an edge in a directed graph takes its `fromNode` and `toNode`.

In C#, use ordinary object references, not unsafe pointers. Do not add separate operations interfaces, callback bags, or copied services to the constructor when the parents already provide access to them.

## 5. Reach data and services through the relationships

Required data and operations must be straightforward to reach from the object that needs them, including logging code. Traverse the parent relationships. Declare the necessary access paths and operations in Common.

If access is awkward, correct the model or its declared operations. Do not compensate with duplicated fields, global searches, reconstructed identifiers, runtime type discovery, or callback plumbing.

## 6. Make logical destruction explicit and complete

Destroying an object must:

- Remove it from every parent relationship by calling the parents' remove methods. Each method also clears the corresponding child back-reference; the destructor must not duplicate that work.
- Recursively destroy its strongly owned children.
- Remove its weak child relationships without destroying those children.
- Use complete two-way relationship operations for every detach, and avoid destroying a multiply owned child more than once.

Strong ownership means deletion cascades. A weak relationship is detached without cascade-deleting the related object. For example, destroying a wire detaches it from its pins; it does not destroy a pin owned by a package. The pin remains, floating.

These are model relationships, not instructions to use `WeakReference<T>`. C# garbage collection reclaims memory; it does not perform this logical destruction. Do not rely on finalizers to maintain the model.

## 7. Declare value-like classes explicitly

Value-like classes, such as vectors, symbols, and matrices, do not require back-pointers to their holders. Rely on C# garbage collection to reclaim their memory when they are no longer reachable; do not add parent tracking or relationship-destruction machinery solely to clean them up.

Declare a type's value-like role explicitly so this exception is clear. It does not require using a C# value type: a value-like object may be implemented as a class. Do not use this exception to disguise relationship-bearing objects or duplicate authoritative mutable state. Short-lived algorithmic references are not automatically persistent model relationships.

## 8. Map the rules directly to C#

Use C# 14 on .NET 10 or a later compatible version, so constructor declarations can be separated from their implementations too.

- **Common owns declarations.** Put the shared class and struct declarations, all their persistent fields and relationships, and every operation callable from another module in Common source files. Declare callable methods and constructors as partial members with explicit accessibility and a semicolon, not a body. Do not put algorithms or accessor bodies in Common.
- **Functional modules own implementations.** Put the matching partial type sections and member bodies in the responsible module directories. Each declared partial member has exactly one implementation. Keep temporary algorithm variables local; do not hide additional persistent object state in implementation files.
- **Match the type identity.** All parts use the same namespace, type name, and generic parameters. Mark each part `partial` and keep its accessibility and other required modifiers consistent. Use matching member signatures and modifiers on the declaration and implementation.
- **Compile the parts together.** Use one assembly and compiled module for these source modules. A directory is not a separate DLL or Go-style package. A partial type cannot be extended from a separately compiled implementation assembly.
- **Keep required operations mandatory.** Use explicitly accessible partial methods so a missing implementation is a compile error. Do not substitute optional private partial hooks whose calls can disappear when no implementation is supplied.
- **Preserve the dependency rule.** Implementations use the vocabulary declared in Common, not another functional module's implementation-only types or helpers. Do not evade this with fully qualified names or global imports. A shared namespace makes the Common declarations directly visible; it does not authorize cross-module private dependencies.
- **Do not add runtime glue.** The compiler combines the parts into one type and one set of fields. No interface field, wrapper object, delegate, factory, or type conversion is needed to connect a declaration to its body. Use ordinary concrete types and direct, non-virtual methods unless an actual requirement needs otherwise. Do not model containment with inheritance.

For example, `Common/Node.cs` declares the partial `Node` type, its data, and signatures such as `public partial Node(Graph parent);` and `public partial void Destroy();`. `Graph/Node.cs` declares the same partial type in the same namespace and supplies those bodies, including relationship registration and cleanup. The shared declarations remain readable in one place; the implementations remain in their functional modules.

## 9. Apply KISS and YAGNI

**KISS — Keep It Simple.** Implement the required behavior with the simplest clear design that respects these rules. Prefer direct operations on known objects and reuse suitable existing code. Remove unnecessary indirection, duplicated mechanisms, and layers that merely pass data along. Preserve useful explanations of non-obvious decisions; fewer lines alone do not mean a simpler design.

**YAGNI — You Aren't Gonna Need It.** Implement current requirements, not imagined future ones. Do not add speculative features, frameworks, registries, configuration options, extension points, or general-purpose abstractions. Add a mechanism when a concrete requirement justifies it, not because it might become useful later.

Simplicity does not excuse missing parent relationships, incomplete destruction, duplicated facts, or violations of the hub-and-spoke structure. Those are requirements, not optional machinery.

The human must understand and agree to changes in the shared vocabulary and architecture. Routine implementation within that agreement does not need per-function permission. Tests and working demonstrations do not excuse violations of the structural rules.

## 10. Use fakes, not mocks

Test real components and the behavior they actually implement. When an external dependency must be replaced, use a small working fake that implements the relevant contract, including the protocol and failure behavior needed by the test. Do not substitute mocks for the implementation being tested or use mocking frameworks and scripted internal-call expectations to manufacture a pass.

Assert meaningful behavior, state, and results. Keep fakes simple and faithful to the boundary they replace. Passing against a fake does not establish that a real external provider or service works.

## 11. Write exemplary teaching code

This code teaches students; clarity, comments, and readable layout are requirements, not optional polish. A student should be able to follow the model and explain why the program works, not merely reproduce its output.

- **Explain the vocabulary and contracts.** Document each module's responsibility and each shared type's role. Describe what callable operations do, their important preconditions, results, failure behavior, and state changes. Constructors and relationship methods must explain ownership and their complete two-way postconditions. State explicitly when a type is value-like.
- **Explain reasons where they matter.** Comment non-obvious algorithms, protocol conversions, ordering, resource lifetime, and deliberately rejected alternatives. Explain why a choice preserves a requirement. Do not substitute narration of obvious syntax or a distant README for an explanation needed beside the code. Tests should explain the behavior or failure they are proving.
- **Prefer readable steps to dense expressions.** Put each substantive operation on its own line. Split non-trivial assignments and transformations into meaningful intermediate steps with names that reveal their purpose. Do not chain assignments, pack multiple statements onto one line, or hide validation, mutation, and cleanup inside a compact expression.
- **Expand control flow.** Use braces and readable multi-line bodies for conditionals, loops, constructors, and error/cleanup paths. Keep their order visible. Use a short expression-bodied member only when the operation is genuinely obvious, such as returning one field; it must not conceal effects or an algorithm.
- **Lay out data and calls for reading.** Give shared fields their own lines and useful explanations of their meaning and ownership. Break long parameter lists, object initializers, and JSON construction across logical lines. Do not extract pass-through helpers or invent abstractions merely to make a method shorter.
- **Review as a teacher.** Before calling a chapter complete, read its declarations and main execution paths as a new student would. Can the student identify the authoritative facts, follow the parent paths, see both ends of each relationship change, and understand failures and cleanup without guessing? Correct missing explanations and dense code before advancing.

Keep comments accurate as the implementation changes. Preserve useful rationale. Fewer lines are not a goal: choose the simplest adequate design and give its explanation enough space. These standards also apply to examples and tests, not just production code.
