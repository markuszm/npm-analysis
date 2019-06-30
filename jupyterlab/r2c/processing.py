import utils
import datetime
from sqlalchemy import create_engine


DB_URL = 'dbUrl'

def processCommitsLatestVersion(repo, countFunc, normalizeFunc, progressString, filterType=""):
    engine = create_engine(DB_URL)
    timeToUsageMapOneRepo = {}
    commitCallCache = {}
    locCache = {}
    kLoC = 0
    for y in range(2010, 2019):
        startMonth = 1
        endMonth = 12
        if y == 2010:
            startMonth = 11
        
        if y == 2018:
            endMonth = 11
                
        for m in range(startMonth, endMonth+1):
            commitHash = utils.findLatestVersionOnDate(y, m, repo, engine)
            if commitHash is None:
                continue
            if commitHash in commitCallCache:
                count = commitCallCache[commitHash]
            else:
                count = countFunc(commitHash, engine)
                commitCallCache[commitHash] = count

            date = datetime.datetime(y, m, 1)
            
            kLoC = getkLoCForCommit(commitHash, locCache, engine, filterType)

            normalized = normalizeFunc(count, kLoC)
            
            for m, c in normalized.items():
                if date not in timeToUsageMapOneRepo:
                    timeToUsageMapOneRepo[date] = {}
                if m in timeToUsageMapOneRepo[date]:
                    timeToUsageMapOneRepo[date][m] = timeToUsageMapOneRepo[date][m] + c
                else:
                    timeToUsageMapOneRepo[date][m] = c
                    

    print(progressString)
    return timeToUsageMapOneRepo, repo, kLoC

def getkLoCForCommit(commitHash, cache, engine, filtered=True):
    loc = 0
    if commitHash in cache:
        loc = cache[commitHash]
    else:
        loc = utils.getLinesOfCodeForCommitId(commitHash, engine, filtered)
        if loc == None:
            loc = 0
        cache[commitHash] = loc  
    kloc = loc / 1000.0
    return kloc

def normalizekLoC(counts, kloc):
    normalized = {}
    for m, c in counts.items():
        locNormalized = 0
        if kloc != 0:
            locNormalized = c / kloc
        normalized[m] = locNormalized
    return normalized

def normalizeNoop(counts, kloc):
    return counts

# immutable function
def averageOverallPackages(timeToUsageMap, repos):
    result = {}
    for key, value in timeToUsageMap.items():
        result[key] = {}
        for m, c in value.items():
            averaged = c[0] / c[1]
            result[key][m] = averaged
    return result

def normalizeOverallPackages(timeToUsageMap, repos):
    result = {}
    for key, value in timeToUsageMap.items():
        result[key] = {}
        for m, c in value.items():
            averaged = c[0] / c[2]
            result[key][m] = averaged
    return result

def processCommitsReleaseTime(repo, countFunc, normalizeFunc, progressString):
    engine = create_engine(DB_URL)
    timeToUsageMapOneRepo = {}
    commitHashes = utils.getSortedCommitHashes(repo, engine)
    for index, row in commitHashes.iterrows():
        commitHash = row['commit_hash']
        count = countFunc(commitHash, engine)
        date = utils.getTimestampForCommitId(commitHash, engine)
        if date == None:
            continue
            
        normalized = normalizeFunc(commitHash, count, {}, engine)
        
        for m, c in normalized.items():
            if date not in timeToUsageMapOneRepo:
                timeToUsageMapOneRepo[date] = {}
            if m in timeToUsageMapOneRepo[date]:
                timeToUsageMapOneRepo[date][m] = timeToUsageMapOneRepo[date][m] + c
            else:
                timeToUsageMapOneRepo[date][m] = c
            
    print(progressString)
            
    return timeToUsageMapOneRepo