#!/usr/bin/env python3
"""Usage: repo-advisories.py <out.txt> --sbom <syft.json>... --grype <grype.json>...

Some advisories are published on a project's own GitHub repo but never reach the global advisory feeds that
Grype and Trivy read. For every github.com Go module in the SBOMs, list the repo's published advisories and write
the ones that affect the shipped version as "target | module | id | severity | fix" rows. Vanity import paths
such as go.opentelemetry.io/otel are resolved to their GitHub repo. Advisories Grype already
reported for the same module (by any of its ids) are skipped. Needs GH_TOKEN. Never fails the workflow."""
import argparse, json, re, subprocess, sys, urllib.request
from concurrent.futures import ThreadPoolExecutor

CMP = re.compile(r"(>=|<=|>|<|=)\s*v?(\d[^\s,]*)")

def vkey(v):
    m = re.match(r"(\d+)\.(\d+)(?:\.(\d+))?(?:-([0-9A-Za-z.-]+))?", v.lstrip("v"))
    if not m: return None
    # a pre-release sorts before its release: (x, y, z, 0, pre) < (x, y, z, 1, "")
    return (int(m[1]), int(m[2]), int(m[3] or 0), 0 if m[4] else 1, m[4] or "")

def affected(version, rng):
    """GitHub ranges join comparators with commas, and sometimes several ranges the same way
    (">= 2.2.0, < 2.2.10, >= 2.3.0, < 2.3.7"), so a lower bound after an upper bound starts a new range."""
    v = vkey(version)
    if v is None: return False
    ranges, cur = [], []
    for op, ver in CMP.findall(rng or ""):
        if op in (">", ">=") and any(o in ("<", "<=", "=") for o, _ in cur):
            ranges.append(cur); cur = []
        cur.append((op, ver))
    if cur: ranges.append(cur)
    for r in ranges:
        ok = True
        for op, ver in r:
            b = vkey(ver)
            if b is None: ok = False; break
            ok &= {">": v > b, ">=": v >= b, "<": v < b, "<=": v <= b, "=": v == b}[op]
        if ok: return True
    return False

def fix_for(version, patched):
    """Smallest patched version above the shipped one; patched_versions lists one fix per release line."""
    cands = [p.strip().lstrip(">= ").strip() for p in (patched or "").split(",") if p.strip()]
    above = sorted((vkey(c), c) for c in cands if vkey(c) and vkey(c) > vkey(version))
    return above[0][1] if above else (cands[0] if cands else "n/a")

def github_repo(module):
    """owner/repo for a Go module path, following the go-import meta tag for vanity paths; None if not on GitHub."""
    parts = module.split("/")
    if parts[0] == "github.com" and len(parts) >= 3: return "/".join(parts[1:3])
    if parts[0] == "golang.org" or len(parts) < 2: return None
    try:
        html = urllib.request.urlopen(f"https://{parts[0]}/{parts[1]}?go-get=1", timeout=10).read().decode("utf-8", "replace")
    except Exception:
        return None
    m = re.search(r'name="go-import"\s+content="\S+\s+git\s+https://github\.com/([^/"\s]+/[^/".\s]+)', html)
    return m[1] if m else None

def advisories(repo):
    r = subprocess.run(["gh", "api", "--paginate", f"repos/{repo}/security-advisories?per_page=100&state=published", "--jq", ".[]"],
                       capture_output=True, text=True)
    if r.returncode:
        print(f"::warning::advisories for {repo}: {r.stderr.strip()[:120]}", file=sys.stderr); return None
    return [json.loads(l) for l in r.stdout.splitlines() if l.strip()]

def main():
    ap = argparse.ArgumentParser(); ap.add_argument("out"); ap.add_argument("--sbom", nargs="+", default=[]); ap.add_argument("--grype", nargs="*", default=[])
    a = ap.parse_args()
    mods = {}  # (module, version) -> target path
    for p in a.sbom:
        for art in json.load(open(p))["artifacts"]:
            if art["type"] == "go-module" and art["version"].startswith("v"):
                mods.setdefault((art["name"], art["version"]), art["locations"][0]["path"])
    known = set()  # (module, id) already reported by Grype
    for p in a.grype:
        for m in json.load(open(p))["matches"]:
            for i in [m["vulnerability"]["id"]] + [x["id"] for x in m.get("relatedVulnerabilities", [])]:
                known.add((m["artifact"]["name"], i))
    with ThreadPoolExecutor(8) as ex:
        prefixes = sorted({"/".join(n.split("/")[:3] if n.startswith("github.com/") else n.split("/")[:2]) for n, _ in mods})
        repo_of = dict(zip(prefixes, ex.map(github_repo, prefixes)))
        repos = sorted({r for r in repo_of.values() if r})
        by_repo = dict(zip(repos, ex.map(advisories, repos)))
    rows = set()
    for (name, ver), target in mods.items():
        key = "/".join(name.split("/")[:3] if name.startswith("github.com/") else name.split("/")[:2])
        for adv in by_repo.get(repo_of.get(key)) or []:
            if adv.get("state") != "published": continue
            ids = {adv["ghsa_id"], adv.get("cve_id")} - {None}
            if any((name, i) in known for i in ids): continue
            for v in adv.get("vulnerabilities") or []:
                if (v.get("package") or {}).get("name") == name and affected(ver, v.get("vulnerable_version_range")):
                    fix = fix_for(ver, v.get("patched_versions"))
                    rows.add(f"{target} | {name} | {adv['ghsa_id']} | {adv['severity'].upper()} | {fix}")
    open(a.out, "w").write("".join(r + "\n" for r in sorted(rows)))
    failed = sum(v is None for v in by_repo.values())
    print(f"{len(repos)} repos checked ({failed} failed), {len(rows)} additional advisories")

if __name__ == "__main__":
    try: main()
    except Exception as e:  # report-only step: never break the daily scan
        print(f"::warning::repo advisory check skipped: {e}", file=sys.stderr); open(sys.argv[1], "a").close()
