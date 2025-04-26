#!/bin/sh

# ANSI color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

# Run commitlint and capture the exit code
npx --no -- commitlint --edit "$1"
EXIT_CODE=$?

# If commitlint failed, print a helpful guide
if [ $EXIT_CODE -ne 0 ]; then
  printf "\n${RED}⛔️ Commit message format error!${NC}\n\n"
  printf "${GREEN}✅ Expected format:${NC}\n"
  printf "  <type>(<scope>): <Capitalized summary>\n\n"
  printf "${YELLOW}Example:${NC}\n"
  printf "  feat(stream-poller): Add Twitch polling support (#10)\n\n"
  printf "${GREEN}📚 Types:${NC} feat, fix, docs, chore, refactor, test, style, perf, build, ci\n"
  printf "${GREEN}📦 Scopes:${NC} stream-poller, webhook-sender, user-service, frontend, infra, auth\n"
fi

exit $EXIT_CODE
