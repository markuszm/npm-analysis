#!/usr/bin/env python3.7

import sys
import json
from datetime import datetime, timedelta
from pprint import pprint
from pymongo import MongoClient

YEARS = range(2015, 2019)
YEAR = 2018
WINDOW = 10
MAX_PKGS = 150000

client = MongoClient('localhost', 27017, authSource='admin', username='npm', password='npm123')
db = client.npm

results = []

for YEAR in YEARS:
    regex = "^{}.*".format(str(YEAR))
    print("Searching with regex {}".format(regex), file=sys.stderr)
    monthdata = db.packageReach.find({"time": { "$regex": regex}})

    yearPackageSets = {}

    for monthpkg in monthdata:
        pkg = monthpkg["package"]
        if pkg not in yearPackageSets:
            yearPackageSets[pkg] = set()
        for reachedPkg in monthpkg["reachedpackages"]:
            yearPackageSets[pkg].add(reachedPkg)

        topReachers = list(yearPackageSets.keys())
        # Take window size into account when sorting, becomes n^2
        topReachers.sort(key = lambda pkg: blen(yearPackageSets[pkg]), reverse=True)
        i = 0
        vals = []
        print("Top reachers for year {}: {}".format(str(YEAR), repr(topReachers[:10])), file=sys.stderr)
        while (i + WINDOW) <= len(topReachers) and (i + WINDOW) < MAX_PKGS:
            # Compute intersection of yearPackageSets[topReachers[i..i+WINDOW]]
            reached = set()
            for x in range(i, i+WINDOW):
                pkg = topReachers[x]
                reached = reached.union(yearPackageSets[pkg])
            vals.append(len(reached))
            i += 1

    results.append({year: vals})

print(json.dumps(vals, sort_keys=True, indent=4, separators=(',', ': ')))
