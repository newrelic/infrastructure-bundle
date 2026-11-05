#!/usr/bin/env bash
# Usage: grype-scan.sh <syft-sbom.json> <fixed-report.txt> <unfixed-report.txt> <image-label>
# Scans a syft SBOM with Grype. Writes the fixable findings as "target | package | id | severity | fix" rows and
# the findings that have no fix yet as "label | package version | id | severity" rows.
set -euo pipefail
sbom=$1 fixed=$2 unfixed=$3 label=$4
all="${fixed%.txt}.json"

# Grype matches Java advisories on Maven coordinates. Syft only guesses CPEs such as apache:httpclient5,
# so add the coordinates from each jar's pom.properties.
jq '(.artifacts[] | select(.type == "java-archive" and .metadata.pomProperties.groupId != null)) |=
    (.cpes = ((.cpes // []) + [{source: "maven-coordinates", cpe:
      ("cpe:2.3:a:" + .metadata.pomProperties.groupId + ":" + .metadata.pomProperties.artifactId + ":" + .version + ":*:*:*:*:maven:*:*")}]))' \
  "$sbom" > "$sbom.maven"

GRYPE_MATCH_JAVA_USING_CPES=true "${GRYPE_CMD:-grype}" "sbom:$sbom.maven" -o json -q > "$all"

jq -r '.matches[] | select(.vulnerability.fix.state == "fixed")
  | "\(.artifact.locations[0].path) | \(.artifact.name) | \(.vulnerability.id) | \(.vulnerability.severity | ascii_upcase) | \(.vulnerability.fix.versions[0] // "n/a")"' "$all" > "$fixed"

jq -r --arg label "$label" '.matches[]
  | select(.vulnerability.fix.state != "fixed" and (.vulnerability.severity | IN("Critical", "High", "Medium", "Low")))
  | "\($label) | \(.artifact.name) \(.artifact.version) | \(.vulnerability.id) | \(.vulnerability.severity | ascii_upcase)"' "$all" | sort -u > "$unfixed"
