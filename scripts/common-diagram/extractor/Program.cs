using System.Security.Cryptography;
using System.Text.Json;
using Microsoft.CodeAnalysis;
using Microsoft.CodeAnalysis.CSharp;
using Microsoft.CodeAnalysis.CSharp.Syntax;

// Source-only inventory: no candidate assembly loading, application execution,
// initializers, or private runtime files. Partial declarations share one type ID.
if (args.Length != 2)
{
    throw new ArgumentException("Usage: Extractor <Common directory> <inventory.json>");
}
string root = Path.GetFullPath(args[0]);
var declarations = new List<TypeDeclarationSyntax>();
var sourceFiles = new List<object>();
foreach (string file in Directory.GetFiles(root, "*.cs").Order())
{
    string text = File.ReadAllText(file);
    var tree = CSharpSyntaxTree.ParseText(text, new CSharpParseOptions(LanguageVersion.Preview), file);
    var errors = tree.GetDiagnostics().Where(d => d.Severity == DiagnosticSeverity.Error).ToArray();
    if (errors.Length != 0)
    {
        throw new InvalidDataException(string.Join("\n", errors.Select(d => d.ToString())));
    }
    declarations.AddRange(tree.GetRoot().DescendantNodes().OfType<TypeDeclarationSyntax>());
    sourceFiles.Add(new { file = Path.GetFileName(file), sha256 = Convert.ToHexStringLower(SHA256.HashData(File.ReadAllBytes(file))) });
}
string TypeID(TypeDeclarationSyntax type)
{
    return string.Join(".", type.Ancestors().OfType<TypeDeclarationSyntax>().Reverse()
        .Select(t => t.Identifier.Text).Append(type.Identifier.Text));
}
object Location(SyntaxNode node)
{
    return new { file = Path.GetFileName(node.SyntaxTree.FilePath), line = node.GetLocation().GetLineSpan().StartLinePosition.Line + 1 };
}
string Comment(SyntaxNode node)
{
    return string.Join("\n", node.GetLeadingTrivia()
        .Where(t => t.IsKind(SyntaxKind.SingleLineCommentTrivia)
            || t.IsKind(SyntaxKind.SingleLineDocumentationCommentTrivia)
            || t.IsKind(SyntaxKind.MultiLineCommentTrivia))
        .Select(t => t.ToFullString().Trim()));
}
var groups = declarations.GroupBy(TypeID).OrderBy(g => g.Key).ToArray();
var known = groups.ToDictionary(g => g.Key, g => g.First().Identifier.Text);
var types = new List<object>();
foreach (var group in groups)
{
    TypeDeclarationSyntax first = group.First();
    bool isStatic = first.Modifiers.Any(SyntaxKind.StaticKeyword);
    var fields = new List<object>();
    void AddField(string name, TypeSyntax syntax, SyntaxNode site, string kind)
    {
        var references = new HashSet<string>();
        foreach (var identifier in syntax.DescendantNodesAndSelf().OfType<SimpleNameSyntax>())
        {
            string nameText = identifier.Identifier.Text;
            string[] candidates = known.Where(k => k.Value == nameText).Select(k => k.Key).ToArray();
            if (candidates.Length == 1)
            {
                references.Add(candidates[0]);
            }
            else if (candidates.Length > 1)
            {
                string? scoped = candidates.SingleOrDefault(k => k == group.Key + "." + nameText);
                if (scoped is null)
                {
                    throw new InvalidDataException($"Ambiguous type reference {group.Key}.{name}: {nameText}");
                }
                references.Add(scoped);
            }
        }
        // Multiplicity is structural, not a claim about a collection's contents
        // at a particular instant. Task<T> is explicitly marked as deferred.
        string[] collections = ["List", "HashSet", "Dictionary", "ImmutableArray", "ImmutableList", "ImmutableDictionary", "ImmutableHashSet", "IEnumerable", "IReadOnlyList", "IReadOnlyDictionary", "Channel", "Queue", "ConcurrentDictionary"];
        bool many = syntax.DescendantNodesAndSelf().OfType<ArrayTypeSyntax>().Any()
            || syntax.DescendantNodesAndSelf().OfType<GenericNameSyntax>().Any(n => collections.Contains(n.Identifier.Text));
        bool deferred = syntax.DescendantNodesAndSelf().OfType<GenericNameSyntax>().Any(n => n.Identifier.Text is "Task" or "TaskCompletionSource" or "Func" or "Action");
        fields.Add(new { name, type = syntax.ToString(), kind, many, deferred,
            references = references.Order().ToArray(), source = Location(site), comment = Comment(site) });
    }
    foreach (TypeDeclarationSyntax declaration in group)
    {
        foreach (FieldDeclarationSyntax field in declaration.Members.OfType<FieldDeclarationSyntax>())
        {
            if (field.Modifiers.Any(SyntaxKind.StaticKeyword) || field.Modifiers.Any(SyntaxKind.ConstKeyword))
            {
                continue;
            }
            foreach (VariableDeclaratorSyntax variable in field.Declaration.Variables)
            {
                AddField(variable.Identifier.Text, field.Declaration.Type, field, "field");
            }
        }
        foreach (PropertyDeclarationSyntax property in declaration.Members.OfType<PropertyDeclarationSyntax>())
        {
            // Auto-properties store state; computed accessors are access paths,
            // not another authoritative copy of the relationship.
            if (!property.Modifiers.Any(SyntaxKind.StaticKeyword) && property.ExpressionBody is null
                && property.AccessorList is not null
                && property.AccessorList.Accessors.All(a => a.Body is null && a.ExpressionBody is null))
            {
                AddField(property.Identifier.Text, property.Type, property, "auto-property");
            }
        }
        if (declaration is RecordDeclarationSyntax record && record.ParameterList is not null)
        {
            foreach (ParameterSyntax parameter in record.ParameterList.Parameters)
            {
                AddField(parameter.Identifier.Text, parameter.Type!, parameter, "record-property");
            }
        }
    }
    types.Add(new { id = group.Key, name = first.Identifier.Text, kind = first.Keyword.Text,
        isStatic, isAbstract = first.Modifiers.Any(SyntaxKind.AbstractKeyword),
        sources = group.Select(Location).ToArray(), comments = group.Select(Comment).Where(c => c.Length != 0).ToArray(),
        bases = group.SelectMany(t => t.BaseList?.Types.Select(b => b.Type.ToString()) ?? []).Distinct().ToArray(), fields });
}
var output = new { schema = 1, scope = "agent/Common/*.cs", files = sourceFiles, types };
File.WriteAllText(args[1], JsonSerializer.Serialize(output, new JsonSerializerOptions { WriteIndented = true }) + "\n");
Console.WriteLine($"Parsed {sourceFiles.Count} source files, merged {types.Count} types; inventory: {args[1]}");
