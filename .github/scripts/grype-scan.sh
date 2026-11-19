#!/usr/bin/env bash
# Usage: grype-scan.sh <syft-sbom.json> <fixed-report.txt> <unfixed-report.txt> <image-label>
# Scans a syft SBOM with Grype. Writes the fixable findings as "target | package | id | severity | fix" rows and
# the findings that have no fix yet as "label | package version | id | severity" rows. The target is the binary or jar
# that carries the package, or "<image-label> (OS)" for apk/deb/rpm packages, which belong to no integration.
# The raw Grype JSON is kept next to <fixed-report.txt> as <name>.json for repo-advisories.py.
# GRYPE_CMD overrides the grype binary.
set -euo pipefail
sbom=$1 fixed=$2 unfixed=$3 label=$4
raw="${fixed%.txt}.json"

# Grype matches Java advisories on Maven coordinates. Syft only guesses CPEs such as apache:httpclient5,
# so add the coordinates from each jar's pom.properties.
jq '(.artifacts[] | select(.type == "java-archive" and .metadata.pomProperties.groupId != null)) |=
    (.metadata.pomProperties as $p | .cpes += [{source: "maven-coordinates",
      cpe: "cpe:2.3:a:\($p.groupId):\($p.artifactId):\(.version):*:*:*:*:maven:*:*"}])' \
  "$sbom" > "$sbom.maven"

GRYPE_MATCH_JAVA_USING_CPES=true "${GRYPE_CMD:-grype}" "sbom:$sbom.maven" -o json -q > "$raw"

jq -r --arg label "$label" '.matches[] | select(.vulnerability.fix.state == "fixed")
  | "\(if .artifact.type | IN("apk", "deb", "rpm") then "\($label) (OS)" else .artifact.locations[0].path end) | \(.artifact.name) | \(.vulnerability.id) | \(.vulnerability.severity | ascii_upcase) | \(.vulnerability.fix.versions[0] // "n/a")"' "$raw" > "$fixed"

jq -r --arg label "$label" '.matches[]
  | select(.vulnerability.fix.state != "fixed" and (.vulnerability.severity | IN("Critical", "High", "Medium", "Low")))
  | "\($label) | \(.artifact.name) \(.artifact.version) | \(.vulnerability.id) | \(.vulnerability.severity | ascii_upcase)"' "$raw" | sort -u > "$unfixed"
