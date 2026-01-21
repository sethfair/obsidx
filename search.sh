#!/bin/bash
# Search wrapper - uses the daemon architecture

# Just pass through to obsidx-recall
# It will auto-start the server if needed
./bin/obsidx-recall "$@"
