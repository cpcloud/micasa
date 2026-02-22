# Copyright 2026 Phillip Cloud
# Licensed under the Apache License, Version 2.0

set -euo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")"

mkdir -p testdata
[[ -f testdata/scanned-invoice.pdf ]] && exit 0
convert testdata/invoice.png \
  -page Letter -gravity North -extent 612x792 \
  testdata/scanned-invoice.pdf
