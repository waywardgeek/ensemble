def slugify(text):
    """Lowercase text and join whitespace-separated words with hyphens."""
    return "-".join(text.lower().split())
