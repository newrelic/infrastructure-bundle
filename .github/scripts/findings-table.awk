# Usage: awk -F'|' -v title=INTEGRATION -v col=1 [-v basename=1] [-v heading=1] [-v max=15] [-v maxids=2] -f findings-table.awk rows.txt
# Turns "x | x | id | SEVERITY | ..." rows into a monospace table for Slack: one line per group with its severity
# counts and its worst finding ids. POSIX awk only (mawk, BSD awk).
#   col      column that names the group (default 1)
#   basename keep only the last path segment of the name, so binary paths become integration names
#   heading  column printed as a sub-heading above its groups when it has more than one value, 0 for none (default)
#   max      show only the worst <max> groups and count the rest, 0 for all (default)
#   maxids   ids listed per group, worst first, then "+N more" (default 2)

function trim(s) { gsub(/^ +| +$/, "", s); return s }
function rank(sev) { return sev == "CRITICAL" ? 0 : sev == "HIGH" ? 1 : sev == "MEDIUM" ? 2 : sev == "LOW" ? 3 : 4 }
function dashes(n,    s) { s = sprintf("%" n "s", ""); gsub(/ /, "-", s); return s }

BEGIN {
    if (!col) col = 1
    if (!maxids) maxids = 2
    if (title == "") title = "NAME"
}

{
    name = trim($col)
    if (basename) sub(/.*\//, "", name)
    head = heading ? trim($heading) : ""
    g = head SUBSEP name
    if (!(head in head_seen)) { head_seen[head] = 1; heads++ }
    if (!(g in count_of_group)) { n++; order[n] = g; head_of[g] = head; name_of[g] = name; count_of_group[g] = 0 }
    r = rank(trim($4))
    count_of_group[g]++
    sev_count[g, r]++
    id = trim($3)
    if (!((g, id) in listed)) {
        listed[g, id] = 1
        ids[g, r] = ids[g, r] (ids[g, r] == "" ? "" : " ") id
    }
    if (length(name) > namew) namew = length(name)
}

END {
    if (n == 0) exit
    if (namew < length(title)) namew = length(title)

    # worst first (most critical, then high, ...), then by name; insertion sort, there are only a handful of groups
    for (i = 1; i <= n; i++) {
        g = order[i]
        sortkey[g] = head_of[g] sprintf("|%05d%05d%05d%05d%05d|", 99999 - sev_count[g, 0], 99999 - sev_count[g, 1], 99999 - sev_count[g, 2], 99999 - sev_count[g, 3], 99999 - sev_count[g, 4]) name_of[g]
        j = i
        while (j > 1 && sortkey[order[j - 1]] > sortkey[g]) { order[j] = order[j - 1]; j-- }
        order[j] = g
    }
    shown = (max > 0 && n > max) ? max : n

    # the worst ids of each group, worst severity first, and how many were left out
    for (i = 1; i <= shown; i++) {
        g = order[i]
        all = ""
        for (r = 0; r <= 4; r++) all = all (all == "" || ids[g, r] == "" ? "" : " ") ids[g, r]
        m = split(all, tok, " ")
        cell[g] = ""
        for (k = 1; k <= m && k <= maxids; k++) cell[g] = cell[g] (k > 1 ? ", " : "") tok[k]
        if (m > maxids) cell[g] = cell[g] " +" (m - maxids) " more"
        if (length(cell[g]) > idw) idw = length(cell[g])
    }

    printf "ST  %-" namew "s | CR | HI | ME | LO | OT | TOT | CVEs\n", title
    print "--- " dashes(namew) " | -- | -- | -- | -- | -- | --- | " dashes(idw)
    for (i = 1; i <= shown; i++) {
        g = order[i]
        if (heads > 1 && head_of[g] != last_head) { print "[" head_of[g] "]"; last_head = head_of[g] }
        c0 = sev_count[g, 0] + 0; c1 = sev_count[g, 1] + 0; c2 = sev_count[g, 2] + 0; c3 = sev_count[g, 3] + 0; c4 = sev_count[g, 4] + 0
        icon = c0 ? "🔴" : c1 ? "🟠" : c2 ? "🟡" : c3 ? "🔵" : "⚪"
        printf "%s  %-" namew "s | %2d | %2d | %2d | %2d | %2d | %3d | %s\n", icon, name_of[g], c0, c1, c2, c3, c4, count_of_group[g], cell[g]
    }
    if (shown < n) printf "... and %d more, see the workflow run\n", n - shown
}
