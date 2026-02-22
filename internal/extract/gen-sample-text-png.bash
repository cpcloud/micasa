# Copyright 2026 Phillip Cloud
# Licensed under the Apache License, Version 2.0

set -euo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")"

mkdir -p testdata
[[ -f testdata/sample-text.png ]] && exit 0
magick -size 612x200 xc:white \
  -pointsize 24 \
  -fill black \
  -annotate +72+72 "Sample text for OCR testing\nInvoice #5678\nDate: 2025-03-15\nVendor: Test Plumbing Co.\nTotal: \$2,500.00" \
  testdata/sample-text.png
