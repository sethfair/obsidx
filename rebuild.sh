#!/bin/bash
set -e

./bin/obsidx-rebuild --db .obsidian-index/obsidx.db --dim 768
