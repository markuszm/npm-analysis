#!/usr/bin/env python3.7

import sys
import json
from datetime import datetime, timedelta
from pprint import pprint
from pymongo import MongoClient

def field_to_date(field):
    def f(row):
        try:
            row[field] = datetime.strptime(
                row[field].split('+')[0],
                '%Y-%m-%dT%H:%M:%S.%f')
        except ValueError as ve:
            row[field] = datetime.strptime(
                row[field].split('+')[0],
                '%Y-%m-%dT%H:%M:%S')
        return row
    return f

with open('cves.json', 'r') as f:
    cves = json.loads(f.read())

converted = list(map(field_to_date('publish'), map(field_to_date('created'), cves)))

latest_publish = max(converted, key=lambda f: f['publish'])
earliest_publish = min(converted, key=lambda f: f['publish'])
latest_create = max(converted, key=lambda f: f['created'])
earliest_create = min(converted, key=lambda f: f['created'])

# print(earliest_publish['publish'], earliest_create['created'])
# print(latest_publish['publish'], latest_create['created'])

earliest= min(earliest_publish['publish'], earliest_create['created'])
latest = max(latest_publish['publish'], latest_create['created'])

current = datetime(
    earliest.year,
    earliest.month,
    1)
morethanamonth = timedelta(
    days=32)

reachesByDate = {}
while current <= latest:
    reachesByDate[current.strftime('%Y-%m-%d')] = set()
    current = current + morethanamonth
    current = datetime(
        current.year,
        current.month,
        1)

# Do the CVE reach calcs
client = MongoClient('localhost', 27017, authSource='admin', username='npm', password='npm123')
db = client.npm
modules = list(map(lambda m: m["module"], cves))
current = datetime(
    earliest.year,
    earliest.month,
    1)

earlystr = earliest.strftime('%Y-%m-%d')
laterstr = latest.strftime('%Y-%m-%d')

while current <= latest:
    key = current.strftime('%Y-%m-%d')
    print(
        "Processing from {} at {} to {}".format(
            earlystr, key, laterstr),
        file=sys.stderr)
    reachItems = db.packageReach.find({"package": {"$in": modules}})
    for reachItem in reachItems:
        reachTime = datetime.strptime(
            reachItem['time'],
            '%Y-%m-%d %H:%M:%S %z %Z') # 2010-11-01 00:00:00 +0000 UTC
        reachDay = datetime(
            reachTime.year,
            reachTime.month,
            1)
        if key not in reachesByDate:
            continue
        if reachDay < current:
            for package in reachItem["reachedpackages"]:
                reachesByDate[key].add(package)
    current = current + morethanamonth
    current = datetime(
        current.year,
        current.month,
        1)

results = {}
for key, value in reachesByDate.items():
    results[key] = len(value)

print(json.dumps(results, sort_keys=True, indent=4, separators=(',', ': ')))
