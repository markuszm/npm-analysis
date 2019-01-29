#!/usr/bin/env python3.7

import sys
import json
from datetime import datetime, timedelta

with open('maintainerFirstDays.json', 'r') as f:
    maintainerFirstDays = json.loads(f.read())

dates = list(map(datetime.fromisoformat, list(maintainerFirstDays.values())))
earliest = min(dates)
latest = max(dates)

current = earliest
morethanamonth = timedelta(days=32)

# To make pandas json parsing happy
results = {"maintainers": {}}
while current <= latest:
    key = current.strftime('%Y-%m-%dT%H:%M:%SZ')
    results["maintainers"][key] = 0
    for maintainer, date in maintainerFirstDays.items():
        if datetime.fromisoformat(date) < current:
            results["maintainers"][key] += 1
    current = current + morethanamonth
    current = datetime(
        current.year,
        current.month,
        1)

print(json.dumps(results, sort_keys=True, indent=4, separators=(',', ': ')))
