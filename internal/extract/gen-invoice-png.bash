# Copyright 2026 Phillip Cloud
# Licensed under the Apache License, Version 2.0

set -euo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")"

mkdir -p testdata
[[ -f testdata/invoice.png ]] && exit 0
convert -size 800x400 xc:white \
  -pointsize 24 \
  -annotate +50+50  "PACIFIC NORTHWEST PLUMBING LLC" \
  -annotate +50+90  "4521 SE Hawthorne Blvd, Portland, OR 97215" \
  -annotate +50+140 "Invoice #INV-2025-0042" \
  -annotate +50+180 "Date: January 15, 2025" \
  -annotate +50+220 "Bill To: Jordan Kim" \
  -annotate +50+280 "Repair kitchen sink drain           \$450.00" \
  -annotate +50+320 "Replace garbage disposal            \$285.00" \
  -annotate +50+370 "Total:                              \$735.00" \
  testdata/invoice.png
