#!/usr/bin/env bash
# SPDX-FileCopyrightText: 2025 Jon Thor Kristinsson
# SPDX-License-Identifier: Apache-2.0

set -euo pipefail

# Check that all Go files have SPDX headers (except generated files)
missing_headers=0
checked_files=0

while IFS= read -r file; do
  # Skip generated files
  if grep -q "Code generated" "$file" 2>/dev/null; then
    continue
  fi

  # Skip test fixtures and mock data
  if [[ "$file" =~ /testdata/ ]] || [[ "$file" =~ /fixtures/ ]]; then
    continue
  fi

  checked_files=$((checked_files + 1))

  # Check for SPDX license identifier
  if ! grep -q "SPDX-License-Identifier: Apache-2.0" "$file"; then
    echo "❌ Missing SPDX header: $file"
    missing_headers=$((missing_headers + 1))
  fi
done < <(find . -name "*.go" -type f \
  -not -path "./vendor/*" \
  -not -path "./.git/*" \
  -not -path "*/testdata/*")

echo "Checked $checked_files Go files"

if [ $missing_headers -gt 0 ]; then
  echo ""
  echo "Found $missing_headers files missing SPDX headers"
  echo ""
  echo "Add the following headers to each file:"
  echo "  // SPDX-FileCopyrightText: 2025 Jon Thor Kristinsson"
  echo "  // SPDX-License-Identifier: Apache-2.0"
  exit 1
fi

echo "✅ All files have proper SPDX headers"
