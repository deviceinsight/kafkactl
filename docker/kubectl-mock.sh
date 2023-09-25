#!/usr/bin/env bash

if [ "$1" = "version" ]; then
  # mock version command
  echo '{"clientVersion": {"major": "9", "minor": "8", "gitVersion": "v9.8.7"}}'
  exit 0
fi

# just print the actual command that was executed
echo "kubectl $@"
