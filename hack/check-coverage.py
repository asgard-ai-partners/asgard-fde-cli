#!/usr/bin/env python3
"""Recompute the asgard-docs coverage row, and fail when the page disagrees.

`internal/corpus/wiki/index.md` states how much of the product documentation
this material is written from. That row is the one place a reader looks to
decide whether an absence means "not covered" or "not there", **so a wrong
denominator makes the material read better than it is** - and the two easiest
to confuse are the number of LINKS it writes and the number of PAGES there are.

**"Uncited" is not "unread" either.** 26 of the uncited pages are two families
this material points at by URL pattern rather than by link - the per-processor
reference and the SSE event pages - so this number measures how much is linked.
Treating it as a reading backlog overstates the backlog by those 26.

So it is computed rather than counted. The page names the commit it was
measured at, and this measures at that same commit - `git ls-tree` on the
clone, so pulling does not change the answer.

    hack/check-coverage.py          measures at the commit the page names
    hack/check-coverage.py --head   also measures at the clone's HEAD
    hack/check-coverage.py --drift  which cited pages have changed since the
                                    commit the citing document names

Needs $ASGARD_DOCS. See hack/sources.py.
"""

import pathlib
import re
import subprocess
import sys

sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent))
from sources import commit, resolve  # noqa: E402

ROOT = pathlib.Path(__file__).resolve().parent.parent
INDEX = ROOT / "internal/corpus/wiki/index.md"

# The three directories the page excludes on purpose. Kept here because the
# check has to count them, and stated on the page with the reason for each.
EXCLUDED = re.compile(r"asgard-builtin/message-template|help-community/release-notes/|superpowers/")

# The row this checks, and the four numbers in it.
ROW = re.compile(
    r"\|\s*asgard-docs\s*\|[^|]*\|\s*(\d+)\s*/\s*(\d+)\s*cited at `([0-9a-f]{7,})`"
    r";\s*(\d+)\s*published and uncited;\s*(\d+)\s*deliberately excluded")


def pages(docs: pathlib.Path, ref: str) -> set:
    """Every published page at one commit, as a slug.

    **`.md` and `.mdx` both.** Counting only `.mdx` gives 148 where the page
    says 162, which is what sent one reading of this down a false trail.
    """
    out = subprocess.run(["git", "-C", str(docs), "ls-tree", "-r", "--name-only", ref, "--", "docs"],
                         capture_output=True, text=True)
    if out.returncode != 0:
        sys.exit(f"{docs} has no commit {ref}. `git -C {docs} fetch` first.")
    got = set()
    for f in out.stdout.split():
        for ext in (".mdx", ".md"):
            if f.endswith(ext):
                got.add(f[len("docs/"):-len(ext)])
                break
    return got


SLUG = re.compile(r'^slug:\s*(\S+)', re.M)


def url_index(docs: pathlib.Path, ref: str) -> dict:
    """Live URL path -> page path, for every page at one commit.

    **A page's URL is its `slug:` frontmatter when it declares one**, and 14 of
    the 158 pages do. Matching a cited URL against the file path instead drops
    every one of them: four channel pages are `integration/LINE.mdx` served at
    `/integration/line`, eight processor pages are named for their family and
    served under the builder's name - `flow-entry.mdx` at `processor/entry` -
    and `product-suite/index.mdx` is served at `product-suite`.

    That undercount is why the coverage row read 71 rather than 77 for a week,
    and why it is worth reading the frontmatter rather than guessing at case and
    `index`, which is what the first fix here did.
    """
    out = {}
    for path in pages(docs, ref):
        for ext in (".mdx", ".md"):
            got = subprocess.run(["git", "-C", str(docs), "show", f"{ref}:docs/{path}{ext}"],
                                 capture_output=True, text=True)
            if got.returncode == 0:
                m = SLUG.search(got.stdout[:1500])
                if m:
                    url = m.group(1).strip().lstrip("/")
                    if url.startswith("docs/"):
                        url = url[len("docs/"):]
                elif path.endswith("/index"):
                    # A directory's index page is served at the directory, with
                    # no `index` in the URL, whether or not it says so.
                    url = path[: -len("/index")]
                else:
                    url = path
                out[url.rstrip("/")] = path
                break
    return out


def resolve_slug(slug: str, index: dict) -> str:
    """The page a cited URL names, or "" when it names no page.

    A miss here means "not a page" - a screenshot under `img/`, or a directory
    with no index, which is the 404 `../wiki/processors.md` warns about - and
    never "spelled differently".
    """
    return index.get(slug.rstrip("/"), "")


def cited_slugs(index: dict = None) -> set:
    """Every docs.asgard-ai.com page the material links to, as a slug."""
    bodies = []
    for pat in ("internal/corpus/**/*.md", "internal/stage/prompts/*.md",
                "internal/scaffold/templates/**/SKILL.md*"):
        for f in ROOT.glob(pat):
            bodies.append(f.read_text(errors="replace"))
    urls = set()
    for b in bodies:
        for u in re.findall(r"https://docs\.asgard-ai\.com/[A-Za-z0-9/_.#-]*[A-Za-z0-9/_-]", b):
            urls.add(u.split("#")[0].rstrip("/"))
    out = set()
    for u in urls:
        s = u.split("docs.asgard-ai.com/", 1)[1]
        slug = s[len("docs/"):] if s.startswith("docs/") else s
        if slug.startswith("img/"):
            continue
        out.add(resolve_slug(slug, index) if index else slug)
    out.discard("")
    return out


def cited_with_commit() -> dict:
    """Each cited docs slug, and the commits the citing documents name beside it.

    **A citation's commit is on the page that cites it, not on the link.** A
    wiki page's Sources section reads "- [thing](url) - asgard-docs `f00e0ee`",
    so the commit governs the block it sits in rather than any one URL, and the
    nearest one above a link is the one that applies. This takes every commit
    named anywhere in the same document, which over-collects rather than
    under-collects: a page citing two commits reports drift against the older,
    and reporting a re-read as drift is the cheap mistake here.
    """
    out = {}
    for pat in ("internal/corpus/**/*.md", "internal/stage/prompts/*.md",
                "internal/scaffold/templates/**/SKILL.md*"):
        for f in ROOT.glob(pat):
            body = f.read_text(errors="replace")
            refs = set(re.findall(r"asgard-docs `([0-9a-f]{7,})`", body))
            if not refs:
                continue
            for u in re.findall(r"https://docs\.asgard-ai\.com/[A-Za-z0-9/_.#-]*[A-Za-z0-9/_-]", body):
                slug = u.split("#")[0].rstrip("/").split("docs.asgard-ai.com/", 1)[1]
                slug = slug[len("docs/"):] if slug.startswith("docs/") else slug
                if slug.startswith("img/"):
                    continue
                out.setdefault(slug, {}).setdefault(f.relative_to(ROOT), set()).update(refs)
    return out


def moved(docs: pathlib.Path, slug: str, since: str) -> str:
    """What happened to one page between a commit and the clone's HEAD.

    The slug is resolved the way `cited_slugs` resolves it - case, and a
    directory's index - so "gone" means gone rather than spelled differently.
    """
    slug = resolve_slug(slug, url_index(docs, "HEAD")) or slug
    for ext in (".mdx", ".md"):
        path = f"docs/{slug}{ext}"
        out = subprocess.run(["git", "-C", str(docs), "log", "--format=%h", f"{since}..HEAD",
                              "--", path], capture_output=True, text=True)
        if out.returncode != 0:
            return "?"
        commits = [c for c in out.stdout.split() if c]
        exists = subprocess.run(["git", "-C", str(docs), "cat-file", "-e", f"HEAD:{path}"],
                                capture_output=True).returncode == 0
        if commits or exists:
            if not exists:
                return "deleted"
            return f"{len(commits)} commit(s)" if commits else ""
    return "not a page at HEAD"


def drift(docs: pathlib.Path) -> int:
    """Which cited pages have changed since the commit the citing document names.

    **This is the checkable half of "the wiki against asgard-docs".** Whether
    somebody read a page is theirs to claim; whether the page has changed under
    the reading is a fact, and this is the list of readings that would have to be
    redone to make the wiki current. It is a report, not a verdict - a page
    changing does not make the prose wrong, it makes it unconfirmed.
    """
    cited = cited_with_commit()
    rows = []
    for slug in sorted(cited):
        for doc, refs in sorted(cited[slug].items()):
            since = sorted(refs)[0] if len(refs) == 1 else min(refs, key=lambda r: age(docs, r))
            what = moved(docs, slug, since)
            if what and what != "?":
                rows.append((str(doc), slug, since, what))
    width = max((len(r[0]) for r in rows), default=0)
    for doc, slug, since, what in rows:
        print(f"  {doc:<{width}}  {slug}  since {since}: {what}")
    print(f"\n{len(rows)} cited page(s) have moved since the commit the citing document names.")
    print("A page changing does not make the prose wrong - it makes it unconfirmed, and\n"
          "this is the size of the re-read.")
    return 0


def age(docs: pathlib.Path, ref: str) -> int:
    out = subprocess.run(["git", "-C", str(docs), "log", "-1", "--format=%ct", ref],
                         capture_output=True, text=True)
    return int(out.stdout.strip()) if out.returncode == 0 and out.stdout.strip() else 0


def measure(docs: pathlib.Path, ref: str):
    P = pages(docs, ref)
    cited = cited_slugs(url_index(docs, ref)) & P
    return len(cited), len(P), len(P) - len(cited), len([p for p in P if EXCLUDED.search(p)])


def main() -> int:
    docs = resolve("docs")
    if "--drift" in sys.argv:
        return drift(docs)
    text = INDEX.read_text()
    m = ROW.search(text)
    if not m:
        sys.exit("the asgard-docs row in internal/corpus/wiki/index.md does not match the shape\n"
                 "this checks: `<cited> / <total> cited at `<commit>`; <n> published and uncited;\n"
                 "<n> deliberately excluded`. Keep the shape or update this script.")
    claim = (int(m.group(1)), int(m.group(2)), int(m.group(4)), int(m.group(5)))
    ref = m.group(3)

    got = measure(docs, ref)
    names = ("cited", "total", "uncited", "excluded")
    print(f"at {ref}, the commit the page names:")
    bad = 0
    for name, c, g in zip(names, claim, got):
        flag = "" if c == g else "   <-- page says " + str(c)
        bad += c != g
        print(f"  {name:<10} {g}{flag}")

    if "--head" in sys.argv:
        head = commit(docs)
        print(f"\nat {head}, the clone's HEAD:")
        for name, g in zip(names, measure(docs, head)):
            print(f"  {name:<10} {g}")
        print("\nA difference here is not a defect: the page records a reading, and the\n"
              "clone has moved since. It is the size of what re-reading would cover.")

    if bad:
        print(f"\n{bad} number(s) in the coverage row do not match a measurement at its own\n"
              "commit. Fix the row, or say what else it counts.")
        return 1
    return 0


if __name__ == "__main__":
    sys.exit(main())
