#!/usr/bin/env bash
set -euo pipefail

fail() {
  echo "safety boundary failed: $*" >&2
  exit 1
}

ignored_paths=(
  ".env"
  ".env.local"
  ".env.production"
  ".repomind/"
  "dist/"
  "eval/"
  "benchmark/"
  "repomind"
  "repomind.exe"
)

for path in "${ignored_paths[@]}"; do
  echo "checking ignored path: $path"
  git check-ignore --quiet -- "$path" || fail "expected path to be ignored by git: $path"
done

if git check-ignore --quiet -- .env.example; then
  fail ".env.example must remain trackable"
fi

forbidden="$(git ls-files -- .env '.env.*' .repomind dist eval benchmark repomind repomind.exe | grep -v '^\.env\.example$' || true)"
if [ -n "$forbidden" ]; then
  echo "$forbidden" >&2
  fail "forbidden generated or secret files are tracked"
fi

if git grep -nE "(OPENAI_API_KEY|ANTHROPIC_API_KEY|CLAUDE_API_KEY|GEMINI_API_KEY|GOOGLE_API_KEY|GROK_API_KEY|XAI_API_KEY|[A-Z0-9_]*TOKEN|[A-Z0-9_]*SECRET)[[:space:]]*=[[:space:]]*(sk-|xai-|gsk_|AIza|ghp_|glpat-)" -- . ':!*.png' ':!*.ico' ':!*.jpg' ':!*.jpeg' ':!*.webp'; then
  fail "tracked files contain likely real API keys or tokens"
fi

echo "safety boundary passed"
