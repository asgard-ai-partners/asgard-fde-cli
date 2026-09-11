#!/usr/bin/env bash
# Run the gate over every reference deployment, and print the count per chart.
#
# **The gate had never been run over the charts its rules were written from.**
# It was run over charts this tool generates, which pass by construction, and
# over a scratch repository. The first time somebody rendered the reference
# deployments through it, R1b was wrong about nine Agents in a running
# deployment: it counted a semantic layer and a Toolset as capability sources
# and not a SkillSet, so every subagent of a flow-agent supervisor was told it
# had "no source of capability at all" while it had one.
#
# It lives here rather than in the binary for the reason everything here does:
# it needs repositories that are not vendored, and a stale copy of somebody
# else's chart proves nothing. Pull them first.
#
#     hack/verify-references.sh [parent-dir]     default: ..
#
# **A count is not a pass.** Read what the findings say. Three kinds turn up:
# a rule that is wrong (fix the rule), a chart that is wrong (tell whoever owns
# it), and a rule right for one shape applied to another (the expensive kind -
# see R1b).
set -uo pipefail

parent="${1:-..}"
cli="${ASGARD_CLI:-asgard-cli}"
command -v helm >/dev/null || { echo "helm is not on PATH; the gate renders with it" >&2; exit 1; }
command -v "$cli" >/dev/null || { echo "$cli is not on PATH; set ASGARD_CLI to a built binary" >&2; exit 1; }

total=0
charts=0
# Both naming shapes: a customer's repo is <customer>-asgard-kube and one of
# ours is asgard-<name>-kube. Matching only the first missed a deployment with
# 15 findings the first time this ran, which is the kind of miss a glob makes
# silently.
for repo in "$parent"/*-kube; do
    [ -d "$repo" ] || continue
    case "$(basename "$repo")" in
        asgard-kube) continue ;;   # the CRD contract, not a deployment
    esac
    for chart in "$repo"/projects/*/chart "$repo"/tenants/*/chart; do
        [ -d "$chart/app" ] || continue
        values=""
        for v in "$chart"/values-prod.yaml "$chart"/values-dev.yaml; do
            [ -f "$v" ] && { values="$v"; break; }
        done
        [ -n "$values" ] || continue

        rendered="$(helm template "$chart/app" -f "$values" 2>/dev/null)"
        [ -n "$rendered" ] || { printf '%-26s %-16s render failed\n' "$(basename "$repo")" "$(basename "$(dirname "$chart")")"; continue; }

        out="$(printf '%s' "$rendered" | "$cli" verify --rendered - 2>&1)"
        n="$(printf '%s' "$out" | grep -c 'FAIL' || true)"
        crs="$(printf '%s' "$rendered" | grep -c '^kind:' || true)"
        printf '%-26s %-16s CRs=%-4s FAIL=%s\n' "$(basename "$repo")" "$(basename "$(dirname "$chart")")" "$crs" "$n"
        [ "$n" -gt 0 ] && printf '%s\n' "$out" | grep 'FAIL' | sed 's/^/    /'
        total=$((total + n))
        charts=$((charts + 1))
    done
done

echo
echo "$charts chart(s), $total finding(s)."
echo
echo "**--rendered sees a chart with nothing around it** - R7, R10 and R11 ask"
echo "questions whose answer can be in the owning repo rather than in the CR, and"
echo "there is nowhere to record one any more: .asgard-config.json held"
echo "olapOnlyLayers and the sampleQuestions exemption, and it is gone from all"
echo "four reference repos. So a finding here may be answered somewhere this run"
echo "cannot see. That is a reason to read a finding, not to discount one: R1b"
echo "was dismissed as exactly this kind of noise once, and it was a bug."
