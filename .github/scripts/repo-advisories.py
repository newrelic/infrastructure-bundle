#!/usr/bin/env python3
"""Usage: repo-advisories.py <fixed.txt> <unfixed.txt> --sbom <syft.json>... --grype <grype.json>...

Some advisories are published on a project's own GitHub repo but never reach the global advisory feeds that
Grype and Trivy read. For every Go module in the SBOMs that lives on GitHub, list the repo's published advisories
and keep the ones that affect the shipped version: with a patch as "target | module | id | severity | fix" rows in
<fixed.txt>, without one as "binary | module version | id | severity" rows in <unfixed.txt>. Vanity import paths
such as go.opentelemetry.io/otel are resolved to their GitHub repo. Advisories Grype already reported for the same
module (by any of its ids) are skipped. Needs the gh CLI and GH_TOKEN."""
import argparse, json, operator, os, re, subprocess, sys, urllib.request
from concurrent.futures import ThreadPoolExecutor

COMPARE = {">": operator.gt, ">=": operator.ge, "<": operator.lt, "<=": operator.le, "=": operator.eq}
COMPARATOR = re.compile(r"(>=|<=|>|<|=)\s*v?(\d[^\s,]*)")

def vkey(v):
    m = re.match(r"(\d+)\.(\d+)(?:\.(\d+))?(?:-([0-9A-Za-z.-]+))?", v.lstrip("v"))
    if not m:
        return None
    # a pre-release sorts before its release: (x, y, z, 0, pre) < (x, y, z, 1, "")
    # ponytail: pre-release ids compare as strings (rc.10 < rc.2); fine for the few advisories on pre-releases
    return (int(m[1]), int(m[2]), int(m[3] or 0), 0 if m[4] else 1, m[4] or "")

def holds(v, op, bound):
    b = vkey(bound)
    return b is not None and COMPARE[op](v, b)

def affected(version, rng):
    """GitHub ranges join comparators with commas, and sometimes several ranges the same way
    (">= 2.2.0, < 2.2.10, >= 2.3.0, < 2.3.7"), so a lower bound after an upper bound starts a new range."""
    v = vkey(version)
    if v is None:
        return False
    ranges, cur = [], []
    for op, bound in COMPARATOR.findall(rng or ""):
        if op in (">", ">=") and any(o in ("<", "<=", "=") for o, _ in cur):
            ranges.append(cur)
            cur = []
        cur.append((op, bound))
    if cur:
        ranges.append(cur)
    return any(all(holds(v, op, bound) for op, bound in r) for r in ranges)

def fix_for(version, patched):
    """Smallest patched version above the shipped one (patched_versions lists one fix per release line); None if unpatched."""
    cands = [p.strip().lstrip(">= ").strip() for p in (patched or "").split(",") if p.strip()]
    shipped = vkey(version)
    above = sorted((k, c) for c in cands if (k := vkey(c)) and shipped and k > shipped)
    return above[0][1] if above else (cands[0] if cands else None)

def module_root(name):
    """The leading part of a Go module path that names a repo: host/owner/repo on github.com, host/first-segment elsewhere."""
    parts = name.split("/")
    return "/".join(parts[:3] if parts[0] == "github.com" else parts[:2])

def github_repo(root):
    """owner/repo for a module root, following the go-import meta tag for vanity paths; None if not on GitHub."""
    host, _, path = root.partition("/")
    if host == "github.com":
        return path
    if host == "golang.org" or not path:
        return None
    try:
        html = urllib.request.urlopen(f"https://{root}?go-get=1", timeout=10).read().decode("utf-8", "replace")
    except Exception:
        return None
    m = re.search(r'name="go-import"\s+content="\S+\s+git\s+https://github\.com/([\w.-]+/[\w.-]+?)(?:\.git)?(?=["/\s])', html)
    return m[1] if m else None

def advisories(repo):
    r = subprocess.run(["gh", "api", "--paginate", f"repos/{repo}/security-advisories?per_page=100&state=published", "--jq", ".[]"],
                       capture_output=True, text=True)
    if r.returncode:
        print(f"::warning::advisories for {repo}: {r.stderr.strip()[:120]}", file=sys.stderr)
        return None
    return [json.loads(line) for line in r.stdout.splitlines() if line.strip()]

def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("fixed_out")
    ap.add_argument("unfixed_out")
    ap.add_argument("--sbom", nargs="+", required=True)
    ap.add_argument("--grype", nargs="*", default=[])
    a = ap.parse_args()

    shipped_in = {}  # (module, version) -> paths of the binaries that carry it
    for path in a.sbom:
        with open(path) as f:
            for art in json.load(f)["artifacts"]:
                if art["type"] == "go-module" and art["version"].startswith("v"):
                    shipped_in.setdefault((art["name"], art["version"]), set()).update(loc["path"] for loc in art["locations"])
    known = set()  # (module, id) already reported by Grype
    for path in a.grype:
        with open(path) as f:
            for m in json.load(f)["matches"]:
                for i in [m["vulnerability"]["id"]] + [x["id"] for x in m.get("relatedVulnerabilities", [])]:
                    known.add((m["artifact"]["name"], i))

    roots = sorted({module_root(name) for name, _ in shipped_in})
    with ThreadPoolExecutor(8) as ex:
        repo_of = dict(zip(roots, ex.map(github_repo, roots)))
        repos = sorted({r for r in repo_of.values() if r})
        by_repo = dict(zip(repos, ex.map(advisories, repos)))

    fixed, unfixed = set(), set()
    for (name, ver), paths in shipped_in.items():
        for adv in by_repo.get(repo_of[module_root(name)]) or []:
            if any((name, i) in known for i in {adv["ghsa_id"], adv.get("cve_id")} - {None}):
                continue
            adv_id = adv.get("cve_id") or adv["ghsa_id"]
            sev = (adv.get("severity") or "unknown").upper()
            for v in adv.get("vulnerabilities") or []:
                if (v.get("package") or {}).get("name") != name or not affected(ver, v.get("vulnerable_version_range")):
                    continue
                fix = fix_for(ver, v.get("patched_versions"))
                for path in paths:
                    if fix:
                        fixed.add(f"{path} | {name} | {adv_id} | {sev} | {fix}")
                    else:
                        unfixed.add(f"{os.path.basename(path)} | {name} {ver} | {adv_id} | {sev}")

    for path, rows in ((a.fixed_out, fixed), (a.unfixed_out, unfixed)):
        with open(path, "w") as f:
            f.writelines(row + "\n" for row in sorted(rows))
    failed = sum(v is None for v in by_repo.values())
    print(f"{len(repos)} repos checked ({failed} failed), {len(fixed)} fixable and {len(unfixed)} unfixed additional advisories")

if __name__ == "__main__":
    main()
