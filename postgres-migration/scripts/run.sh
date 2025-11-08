#!/bin/bash
set -e

echo "Starting migrations..."

for dir in /app/migrations/*/; do
    if [ -d "$dir" ]; then
        echo "Found migrations from version $(basename $dir)"
        tern migrate -c /app/tern.conf -m "$dir"
        echo "Migrations from $(basename $dir) completed"
    fi
done
        
echo "All migrations completed successfully"
