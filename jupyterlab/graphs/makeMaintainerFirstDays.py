#!/usr/bin/env python3.7

import sys
import json
from datetime import datetime
from pprint import pprint
from pymongo import MongoClient

client = MongoClient('localhost', 27017, authSource='admin', username='npm', password='npm123')
db = client.npm

maintainerFirstDays = {}

for module in db.timelineNew.find():
    for event in module['timeline']:
        parsedTime = datetime.strptime(event['time'], '%Y-%m-%d %H:%M:%S %z %Z') # "2012-11-01 00:00:00 +0000 UTC"
        parsedDay = datetime(
            parsedTime.year,
            parsedTime.month,
            parsedTime.day)
        maintainers = event['packageData']['maintainers']
        if maintainers is None:
            continue
        for maintainer in maintainers:
            if maintainer not in maintainerFirstDays:
                maintainerFirstDays[maintainer] = parsedDay
            elif maintainerFirstDays[maintainer] > parsedDay:
                maintainerFirstDays[maintainer] = parsedDay

jsonable = {}
for key, value in maintainerFirstDays.items():
    jsonable[key] = value.isoformat()


print(json.dumps(jsonable, sort_keys=True, indent=4, separators=(',', ': ')))
