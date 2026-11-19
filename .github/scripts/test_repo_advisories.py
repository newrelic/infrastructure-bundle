#!/usr/bin/env python3
"""Self-check for the version logic in repo-advisories.py. Run: python3 .github/scripts/test_repo_advisories.py"""
import importlib.util, pathlib

spec = importlib.util.spec_from_file_location("ra", pathlib.Path(__file__).with_name("repo-advisories.py"))
ra = importlib.util.module_from_spec(spec)
spec.loader.exec_module(ra)

MULTI = ">= 2.2.0, < 2.2.10, >= 2.3.0, < 2.3.7"  # several ranges joined by commas, as GitHub writes them
assert [ra.affected(v, MULTI) for v in ("v2.1.9", "v2.2.4", "v2.2.10", "v2.3.0", "v2.3.7")] == [False, True, False, True, False]
assert ra.affected("v2.2.4", "<= 2.4.2") and not ra.affected("v2.4.3", "<= 2.4.2")
assert ra.affected("v1.0.0", "= 1.0.0") and not ra.affected("v1.0.1", "= 1.0.0")
assert ra.affected("v1.2.3-rc.1", "< 1.2.3") and not ra.affected("v1.2.3", "< 1.2.3")  # pre-release sorts before its release
assert ra.affected("v0.0.0-20210101000000-abcdef123456", "< 0.1.0")  # Go pseudo-version
assert not ra.affected("v1.0.0", "") and not ra.affected("dev", "< 1.0.0")

assert ra.fix_for("v2.2.4", "2.2.10, 2.3.7") == "2.2.10" and ra.fix_for("v2.3.0", "2.2.10, 2.3.7") == "2.3.7"
assert ra.fix_for("v2.2.4", ">= 2.4.3") == "2.4.3"
assert ra.fix_for("v1.0.0", None) is None and ra.fix_for("v1.0.0", "") is None

assert ra.module_root("github.com/o/r/v2/sub") == "github.com/o/r"
assert ra.module_root("go.opentelemetry.io/otel/sdk/metric") == "go.opentelemetry.io/otel"
print("ok")
