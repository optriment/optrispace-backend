#!/bin/bash
set -e

# NOTE: From https://github.com/mskri/check-uncommitted-changes-action
status=$(git status --porcelain)
if [ -z "$status" ]; then
  exit 0
fi

echo "There are uncommitted changes:"
echo "$status"

exit 1
