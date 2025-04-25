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
  echo ""
  echo -e "${RED}⛔️ Commit message format error!${NC}"
  echo ""
  echo -e "${GREEN}✅ Expected format:${NC}"
  echo "  <type>(<scope>): <Capitalized summary>"
  echo ""
  echo -e "${YELLOW}Example:${NC}"
  echo "  feat(stream-poller): Add Twitch polling support (#10)"
  echo ""
  echo -e "${GREEN}📚 Types:${NC} feat, fix, docs, chore, refactor, test, style, perf, build, ci"
  echo -e "${GREEN}📦 Scopes:${NC} stream-poller, webhook-sender, user-service, frontend, infra, auth"
fi

exit $EXIT_CODE
