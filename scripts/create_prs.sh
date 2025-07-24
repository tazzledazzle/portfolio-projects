#!/usr/bin/env bash
set -euo pipefail

PRS=(
  "01 chore/repo-bootstrap-hygiene .github/pr-descriptions/01-chore-repo-bootstrap-hygiene.md"
  "02 docs/ia-rename-and-manifest .github/pr-descriptions/02-docs-ia-rename-and-manifest.md"
  "03 docs/readme-project-table .github/pr-descriptions/03-readme-project-table.md"
  "04 ci/matrix-lint-test-build-docs .github/pr-descriptions/04-ci-matrix-lint-test-docs.yml.md"
  "05 dev/devcontainer-and-bootstrap-script .github/pr-descriptions/05-devcontainer-bootstrap.md"
  "06 docs/gh-pages-site .github/pr-descriptions/06-docs-gh-pages-site.md"
  "07 tool/portfolio-runner .github/pr-descriptions/07-tool-portfolio-runner.md"
  "08 feat/observability-stack .github/pr-descriptions/08-feat-observability-stack.md"
  "09 ci/release-please-dependabot-badges .github/pr-descriptions/09-ci-release-please-dependabot-badges.md"
  "10 meta/competency-mapping .github/pr-descriptions/10-meta-competency-mapping.md"
)

ROOT_BRANCH="main"

# sanity check
if ! command -v gh >/dev/null; then
  echo "gh CLI required" >&2
  exit 1
fi

git fetch origin "$ROOT_BRANCH"
git checkout "$ROOT_BRANCH"
git pull origin "$ROOT_BRANCH"

for ENTRY in "${PRS[@]}"; do
  read -r num branch pr_body <<<"$ENTRY"
  branch_name="${branch// /-}"  # spaces → hyphen

  echo "=== Working on $num: $branch_name ==="

  git checkout -b "$branch_name" "$ROOT_BRANCH"

  # Create placeholder commit if nothing staged.
  if [ -z "$(git status --porcelain)" ]; then
    echo "# Placeholder for $branch_name" > "PLACEHOLDER-$branch_name.txt"
    git add "PLACEHOLDER-$branch_name.txt"
  fi

  git commit -m "${branch%% *}: scaffold for ${branch_name}"
  git push -u origin "$branch_name"

  gh pr create \
    --draft \
    --title "${branch^}" \
    --body-file "$pr_body" \
    -B "$ROOT_BRANCH"

  git checkout "$ROOT_BRANCH"
done