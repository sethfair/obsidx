#!/bin/bash
set -e

DB_PATH="${1:-.obsidian-index/obsidx.db}"

echo "🔄 Migrating database schema..."
echo "Database: $DB_PATH"
echo ""

if [ ! -f "$DB_PATH" ]; then
    echo "✓ No existing database found - schema will be created fresh"
    exit 0
fi

echo "Backing up database..."
cp "$DB_PATH" "$DB_PATH.backup.$(date +%s)"
echo "✓ Backup created"

echo ""
echo "Adding category columns..."

sqlite3 "$DB_PATH" <<EOF
-- Add new columns if they don't exist
ALTER TABLE chunks ADD COLUMN category TEXT DEFAULT 'project';
ALTER TABLE chunks ADD COLUMN status TEXT DEFAULT 'active';
ALTER TABLE chunks ADD COLUMN scope TEXT;
ALTER TABLE chunks ADD COLUMN note_type TEXT;
ALTER TABLE chunks ADD COLUMN category_weight REAL DEFAULT 1.0;
ALTER TABLE chunks ADD COLUMN canon INTEGER DEFAULT 0;

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_chunks_category ON chunks(category);
CREATE INDEX IF NOT EXISTS idx_chunks_status ON chunks(status);
CREATE INDEX IF NOT EXISTS idx_chunks_canon ON chunks(canon);
CREATE INDEX IF NOT EXISTS idx_chunks_category_active ON chunks(category, active);

-- Update existing records with default values
UPDATE chunks SET category = 'project' WHERE category IS NULL;
UPDATE chunks SET status = 'active' WHERE status IS NULL;
UPDATE chunks SET category_weight = 1.0 WHERE category_weight IS NULL;
UPDATE chunks SET canon = 0 WHERE canon IS NULL;
EOF

echo "✓ Migration complete"
echo ""
echo "Database is ready with category support!"
