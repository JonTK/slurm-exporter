#!/bin/bash
# Script to add SPDX license headers to Go files

set -e

HEADER="// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors
"

# Find all Go files that don't have SPDX headers
while IFS= read -r file; do
    # Skip generated files
    if grep -q "Code generated" "$file" 2>/dev/null; then
        echo "⏭️  Skipping generated file: $file"
        continue
    fi
    
    # Skip test fixtures and mock data
    if [[ "$file" =~ /testdata/ ]] || [[ "$file" =~ /fixtures/ ]]; then
        echo "⏭️  Skipping test fixture: $file"
        continue
    fi
    
    # Check if file already has SPDX header
    if grep -q "SPDX-License-Identifier: Apache-2.0" "$file"; then
        echo "✅ Already has header: $file"
        continue
    fi
    
    # Add header to file
    echo "$HEADER" | cat - "$file" > "$file.tmp" && mv "$file.tmp" "$file"
    echo "➕ Added header: $file"
    
done < <(find . -name "*.go" -type f -not -path "./vendor/*" -not -path "./.git/*")

echo ""
echo "✅ SPDX headers added to all Go files!"
