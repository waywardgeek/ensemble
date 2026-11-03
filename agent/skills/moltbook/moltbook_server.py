#!/usr/bin/env python3
"""
Moltbook MCP Server — tools for reading and posting on Moltbook.

Lobster math challenges are returned to the AI to solve; call moltbook_verify
with the verification_code and your computed answer.

Adapted by Cadence from CodeRhapsody's Moltbook skill; original preserved.
No pip dependencies — uses only stdlib.
"""

import json
import sys
import os
import urllib.request
import urllib.error
import urllib.parse

BASE_URL = "https://www.moltbook.com/api/v1"
CREDS_PATH = os.path.expanduser("~/.config/moltbook/credentials.json")


MAX_RESPONSE_BYTES = 2 * 1024 * 1024


class ApiError(Exception):
    """Fixed, credential-free diagnostic suitable for a tool result."""


class NoRedirect(urllib.request.HTTPRedirectHandler):
    def redirect_request(self, req, fp, code, msg, headers, newurl):
        raise ApiError("Moltbook redirect refused; credentials were not forwarded")


def load_api_key():
    try:
        with open(CREDS_PATH, encoding="utf-8") as stream:
            text = stream.read(65537)
        if len(text) > 65536:
            raise ValueError()
        key = json.loads(text)["api_key"]
        if not isinstance(key, str) or not key.strip() or any(c in key for c in "\r\n"):
            raise ValueError()
        return key
    except (OSError, ValueError, KeyError, TypeError):
        raise ApiError("Moltbook credentials are missing or invalid") from None


def api_request(method, path, body=None):
    api_key = load_api_key()
    data = json.dumps(body).encode() if body is not None else None
    req = urllib.request.Request(BASE_URL + path, data=data, method=method,
        headers={"Authorization": f"Bearer {api_key}", "Content-Type": "application/json"})
    try:
        # Ensemble owns observation/Job lifetime. This synchronous server does
        # not promise that cancelling a Job aborts an already accepted HTTP action.
        with urllib.request.build_opener(NoRedirect()).open(req) as response:
            raw = response.read(MAX_RESPONSE_BYTES + 1)
        if len(raw) > MAX_RESPONSE_BYTES:
            raise ApiError("Moltbook response exceeds the size limit")
        result = json.loads(raw)
        if not isinstance(result, dict):
            raise ApiError("Moltbook returned an invalid response")
        if result.get("error") or result.get("success") is False:
            raise ApiError("Moltbook reported an API failure")
        return result
    except urllib.error.HTTPError as error:
        code = error.code
        error.close()
        raise ApiError(f"Moltbook HTTP {code}; response details withheld") from None
    except (urllib.error.URLError, OSError):
        raise ApiError("Moltbook network request failed; outcome may be unknown") from None
    except (ValueError, UnicodeError):
        raise ApiError("Moltbook returned invalid JSON") from None


def component(value):
    return urllib.parse.quote(value, safe="")


def validate_args(name, args):
    if not isinstance(args, dict):
        raise ApiError("Tool arguments must be an object")
    schema = TOOLS[name]["schema"]
    if set(args) - set(schema["properties"]) or any(k not in args for k in schema.get("required", [])):
        raise ApiError("Unknown or missing tool argument")
    for key, value in args.items():
        kind = schema["properties"][key]["type"]
        if kind == "string" and (not isinstance(value, str) or not value.strip() or len(value) > 100000):
            raise ApiError("Invalid text argument")
        if kind == "integer" and (type(value) is not int or not 1 <= value <= 100):
            raise ApiError("Limit must be an integer from 1 to 100")
    if "sort" in args and args["sort"] not in ("best", "new", "old"):
        raise ApiError("Sort must be best, new, or old")


# ── Tool handlers ──────────────────────────────────────────────────────────────

def moltbook_status(_args):
    """GET /home — karma, notification count, and activity on your posts."""
    data = api_request("GET", "/home")
    account = data.get("your_account", {})
    activity = data.get("activity_on_your_posts", [])
    result = {
        "account": account.get("name") or account.get("username"),
        "karma": account.get("karma"),
        "unread_notifications": account.get("unread_notification_count"),
        "activity": [
            {
                "post_id": a["post_id"],
                "post_title": a["post_title"],
                "new_notifications": a["new_notification_count"],
                "latest_commenters": a["latest_commenters"],
                "preview": a["preview"],
            }
            for a in activity
        ],
    }
    return json.dumps(result, indent=2)


def moltbook_feed(args):
    """GET /feed — list recent posts from the platform."""
    limit = args.get("limit", 10)
    data = api_request("GET", f"/feed?limit={limit}")
    posts = data.get("posts", [])
    result = [
        {
            "id": p.get("id"),
            "title": p.get("title"),
            "author": p.get("author_name"),
            "submolt": p.get("submolt_name"),
            "score": p.get("score"),
            "preview": p.get("content_preview", "")[:200] if p.get("content_preview") else "",
        }
        for p in posts
    ]
    return json.dumps(result, indent=2)


def moltbook_read_post(args):
    """GET /posts/{id} — full post content."""
    post_id = component(args["post_id"])
    data = api_request("GET", f"/posts/{post_id}")
    post = data.get("post", data)
    return json.dumps({
        "id": post.get("id"),
        "title": post.get("title"),
        "author": post.get("author_name"),
        "content": post.get("content"),
        "score": post.get("score"),
        "comment_count": post.get("comment_count"),
        "created_at": post.get("created_at"),
    }, indent=2)


def moltbook_comments(args):
    """GET /posts/{id}/comments — comments on a post."""
    post_id = component(args["post_id"])
    sort = args.get("sort", "best")
    limit = args.get("limit", 20)
    data = api_request("GET", f"/posts/{post_id}/comments?sort={sort}&limit={limit}")
    comments = data.get("comments", [])

    def fmt(c):
        return {
            "id": c.get("id"),
            "author": c.get("author", {}).get("name") if isinstance(c.get("author"), dict) else c.get("author"),
            "content": c.get("content"),
            "score": c.get("score"),
            "created_at": c.get("created_at"),
            "replies": [fmt(r) for r in c.get("replies", [])],
        }

    return json.dumps([fmt(c) for c in comments], indent=2)


def moltbook_comment(args):
    """POST /posts/{id}/comments — post a comment.

    Returns the verification challenge so the AI can solve the lobster math.
    Call moltbook_verify with the verification_code and your answer to publish.
    """
    post_id = component(args["post_id"])
    content = args["content"]
    data = api_request("POST", f"/posts/{post_id}/comments", {"content": content})
    comment = data.get("comment", {})
    verification = comment.get("verification", {})
    if not comment.get("id"):
        raise ApiError("Moltbook comment response lacks an ID; outcome unknown, do not retry automatically")
    return json.dumps({
        "comment_id": comment.get("id"),
        "status": comment.get("verification_status"),
        "challenge": verification.get("challenge_text"),
        "verification_code": verification.get("verification_code"),
        "expires_at": verification.get("expires_at"),
        "instructions": "Solve the math challenge and call moltbook_verify with the verification_code and answer (2 decimal places, e.g. '57.00')",
    }, indent=2)


def moltbook_verify(args):
    """POST /verify — submit answer to the lobster math challenge."""
    verification_code = args["verification_code"]
    answer = args["answer"]
    data = api_request("POST", "/verify", {
        "verification_code": verification_code,
        "answer": str(answer),
    })
    return json.dumps(data, indent=2)


def moltbook_mark_read(args):
    """POST /notifications/read-by-post/{id} — mark notifications as read."""
    post_id = component(args["post_id"])
    data = api_request("POST", f"/notifications/read-by-post/{post_id}")
    return json.dumps(data, indent=2)


def moltbook_delete_comment(args):
    """DELETE /comments/{id} — delete a comment (e.g. when challenge expires)."""
    comment_id = component(args["comment_id"])
    data = api_request("DELETE", f"/comments/{comment_id}")
    return json.dumps(data, indent=2)


def moltbook_follow(args):
    """POST /agents/{username}/follow — follow an agent by username."""
    username = component(args["username"])
    data = api_request("POST", f"/agents/{username}/follow")
    return json.dumps(data, indent=2)


def moltbook_search(args):
    """GET /search?q=... — search for posts and agents."""
    q = args["query"]
    limit = args.get("limit", 10)
    data = api_request("GET", f"/search?q={urllib.parse.quote(q)}&limit={limit}")
    results = data.get("results", [])
    return json.dumps([
        {
            "type": r.get("type"),
            "id": r.get("id"),
            "title": r.get("title"),
            "content": (r.get("content") or "")[:200],
            "author": r.get("author", {}).get("name") if r.get("author") else None,
            "post_id": r.get("post_id") or None,
            "url": r.get("url"),
        }
        for r in results
    ], indent=2)


def moltbook_post(args):
    """POST /posts — create a new post on Moltbook.

    Returns a verification challenge. Call moltbook_verify to publish.
    """
    title = args["title"]
    content = args["content"]
    submolt_name = args.get("submolt_name", "general")
    # Look up submolt ID
    submolts = api_request("GET", "/submolts")
    submolt_id = None
    for s in submolts.get("submolts", []):
        if s["name"] == submolt_name:
            submolt_id = s["id"]
            break
    if not submolt_id:
        raise ApiError("Requested Moltbook community was not found")

    data = api_request("POST", "/posts", {
        "title": title,
        "content": content,
        "submolt_name": submolt_name,
        "submolt": submolt_id,
    })
    if "error" in data:
        return json.dumps(data, indent=2)
    post = data.get("post", data)
    verification = post.get("verification", {})
    if not post.get("id"):
        raise ApiError("Moltbook post response lacks an ID; outcome unknown, do not retry automatically")
    return json.dumps({
        "post_id": post.get("id"),
        "status": post.get("verification_status") or post.get("verificationStatus"),
        "challenge": verification.get("challenge_text"),
        "verification_code": verification.get("verification_code"),
        "expires_at": verification.get("expires_at"),
        "instructions": "Solve the math challenge and call moltbook_verify with the verification_code and answer (2 decimal places, e.g. '57.00')",
    }, indent=2)


def moltbook_upvote(args):
    """POST /posts/{id}/upvote — upvote a post."""
    post_id = component(args["post_id"])
    data = api_request("POST", f"/posts/{post_id}/upvote")
    return json.dumps(data, indent=2)


# ── MCP stdio server ───────────────────────────────────────────────────────────

TOOLS = {
    "moltbook_status": {
        "fn": moltbook_status,
        "schema": {
            "type": "object",
            "properties": {},
            "required": [],
        },
        "description": "Get your Moltbook karma, unread notification count, and recent activity on your posts.",
    },
    "moltbook_feed": {
        "fn": moltbook_feed,
        "schema": {
            "type": "object",
            "properties": {
                "limit": {"type": "integer", "description": "Number of posts to return (default 10)"},
            },
            "required": [],
        },
        "description": "Get recent posts from the Moltbook platform feed.",
    },
    "moltbook_read_post": {
        "fn": moltbook_read_post,
        "schema": {
            "type": "object",
            "properties": {
                "post_id": {"type": "string", "description": "Post UUID"},
            },
            "required": ["post_id"],
        },
        "description": "Read the full content of a Moltbook post by ID.",
    },
    "moltbook_comments": {
        "fn": moltbook_comments,
        "schema": {
            "type": "object",
            "properties": {
                "post_id": {"type": "string", "description": "Post UUID"},
                "sort": {"type": "string", "description": "Sort order: best, new, old (default: best)"},
                "limit": {"type": "integer", "description": "Number of comments (default 20)"},
            },
            "required": ["post_id"],
        },
        "description": "Read comments on a Moltbook post.",
    },
    "moltbook_comment": {
        "fn": moltbook_comment,
        "schema": {
            "type": "object",
            "properties": {
                "post_id": {"type": "string", "description": "Post UUID to comment on"},
                "content": {"type": "string", "description": "Comment text"},
            },
            "required": ["post_id", "content"],
        },
        "description": "Post a comment on Moltbook. Returns a lobster math challenge to solve — call moltbook_verify with your answer to publish.",
    },
    "moltbook_verify": {
        "fn": moltbook_verify,
        "schema": {
            "type": "object",
            "properties": {
                "verification_code": {"type": "string", "description": "Code returned by moltbook_comment"},
                "answer": {"type": "string", "description": "Math answer with 2 decimal places, e.g. '57.00'"},
            },
            "required": ["verification_code", "answer"],
        },
        "description": "Submit your answer to the Moltbook lobster math challenge to publish a pending comment.",
    },
    "moltbook_mark_read": {
        "fn": moltbook_mark_read,
        "schema": {
            "type": "object",
            "properties": {
                "post_id": {"type": "string", "description": "Post UUID whose notifications to mark read"},
            },
            "required": ["post_id"],
        },
        "description": "Mark all notifications on a Moltbook post as read.",
    },
    "moltbook_delete_comment": {
        "fn": moltbook_delete_comment,
        "schema": {
            "type": "object",
            "properties": {
                "comment_id": {"type": "string", "description": "Comment UUID to delete"},
            },
            "required": ["comment_id"],
        },
        "description": "Delete a Moltbook comment only with explicit authority. Never automatically delete and repost after an expired challenge.",
    },
    "moltbook_follow": {
        "fn": moltbook_follow,
        "schema": {
            "type": "object",
            "properties": {
                "username": {"type": "string", "description": "Username to follow (e.g. 'hewitt')"},
            },
            "required": ["username"],
        },
        "description": "Follow an agent on Moltbook by username.",
    },
    "moltbook_search": {
        "fn": moltbook_search,
        "schema": {
            "type": "object",
            "properties": {
                "query": {"type": "string", "description": "Search query"},
                "limit": {"type": "integer", "description": "Number of results (default 10)"},
            },
            "required": ["query"],
        },
        "description": "Search Moltbook for posts and agents.",
    },
    "moltbook_post": {
        "fn": moltbook_post,
        "schema": {
            "type": "object",
            "properties": {
                "title": {"type": "string", "description": "Post title"},
                "content": {"type": "string", "description": "Post body (markdown supported)"},
                "submolt_name": {"type": "string", "description": "Submolt name (default: general). Options: general, builds, ai, philosophy, memory, security, technology, todayilearned, etc."},
            },
            "required": ["title", "content"],
        },
        "description": "Create a new post on Moltbook. Returns a lobster math challenge — call moltbook_verify with your answer to publish.",
    },
    "moltbook_upvote": {
        "fn": moltbook_upvote,
        "schema": {
            "type": "object",
            "properties": {
                "post_id": {"type": "string", "description": "Post UUID to upvote"},
            },
            "required": ["post_id"],
        },
        "description": "Upvote a Moltbook post.",
    },
}


def send(obj):
    sys.stdout.write(json.dumps(obj) + "\n")
    sys.stdout.flush()


def handle(request):
    if not isinstance(request, dict):
        return
    method = request.get("method")
    req_id = request.get("id")

    if method == "initialize":
        send({"jsonrpc": "2.0", "id": req_id, "result": {
            "protocolVersion": "2024-11-05",
            "capabilities": {"tools": {}},
            "serverInfo": {"name": "moltbook", "version": "1.1.0"},
        }})

    elif method == "tools/list":
        tools = []
        for name, t in TOOLS.items():
            tools.append({
                "name": name,
                "description": t["description"],
                "inputSchema": t["schema"],
            })
        send({"jsonrpc": "2.0", "id": req_id, "result": {"tools": tools}})

    elif method == "tools/call":
        params = request.get("params", {})
        if not isinstance(params, dict):
            send({"jsonrpc": "2.0", "id": req_id, "error": {"code": -32602, "message": "Invalid parameters"}})
            return
        name = params.get("name")
        args = params.get("arguments", {})
        if not isinstance(name, str) or name not in TOOLS:
            send({"jsonrpc": "2.0", "id": req_id, "error": {
                "code": -32601, "message": "Unknown tool",
            }})
            return
        try:
            validate_args(name, args)
            result = TOOLS[name]["fn"](args)
            send({"jsonrpc": "2.0", "id": req_id, "result": {
                "content": [{"type": "text", "text": result}],
            }})
        except Exception as error:
            # Never echo arbitrary exception text, HTTP bodies, or credential files.
            message = str(error) if isinstance(error, ApiError) else "Moltbook tool failed; details withheld"
            send({"jsonrpc": "2.0", "id": req_id, "result": {
                "isError": True, "content": [{"type": "text", "text": message}],
            }})

    elif method in ("notifications/initialized", "$/cancelRequest", "notifications/cancelled"):
        pass  # no-op

    else:
        if req_id is not None:
            send({"jsonrpc": "2.0", "id": req_id, "error": {
                "code": -32601, "message": f"Method not found: {method}",
            }})


def main():
    global CREDS_PATH
    import argparse
    parser = argparse.ArgumentParser(description="Moltbook stdio MCP bridge")
    parser.add_argument("--credentials", default=CREDS_PATH, help="Explicit existing credential file; never copied")
    CREDS_PATH = parser.parse_args().credentials
    for line in sys.stdin:
        line = line.strip()
        if not line:
            continue
        try:
            request = json.loads(line)
        except json.JSONDecodeError:
            continue
        handle(request)


if __name__ == "__main__":
    main()
