#!/usr/bin/env python3.7

import sys
import json

with open('cves_reach_growth.json', 'r') as f:
    reach = json.loads(f.read())
with open('pkgCounts.json', 'r') as f:
    counts = json.loads(f.read())

results = {"normalized": {}}
for key, value in reach["cves"].items():
    results["normalized"][key] = float(value) / float(counts["packageCount"][key])

print(json.dumps(results, sort_keys=True, indent=4, separators=(',', ': ')))
