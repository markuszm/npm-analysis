import pandas as pd
import numpy as np
import re
from distutils.version import StrictVersion
import math
import numpy as np

def normalize(values):
    length = len(values)
    bucketSize = length / 10
    percentageSums = np.zeros(10)
    for x in range(0,10):
        leftIdx = math.floor(x * bucketSize)        
        rightIdx = math.ceil(leftIdx + bucketSize)
        percentageSums[x] = np.sum(values[leftIdx:rightIdx])
    return percentageSums

def getSortedCommitHashes(repo, engine, population="top_10k"):
    commitHashes = pd.read_sql_query("select commit_hash, tag from repo_tags_2 where repo_url = %s and population = %s order by tag", params=[repo, population], con=engine)
    versionRegex = re.compile("^v?(\d+\.\d+\.\d+)")
    toDrop = []
    for index, row in commitHashes.iterrows():
        result = versionRegex.match(row["tag"])
        if result == None:
            toDrop.append(index)
        else: 
            row["tag"] = result[1]
    for index in toDrop:
        commitHashes = commitHashes.drop(index)
    commitHashes['tag'] = commitHashes['tag'].apply(StrictVersion)
    commitHashes = commitHashes.sort_values(by=['tag'])
    return commitHashes

def getTimestampForCommitId(commitId, engine, population="top_10k"):
    timestamp = pd.read_sql_query("select author_timestamp from repo_tags_2 where commit_hash = %s and population = %s", params=[commitId, population], con=engine)
    val = timestamp["author_timestamp"][0]
    if val == None:
        return None
    return val.year, val.month

def getCommitIdForTag(repo, tag, engine):
    commitId = pd.read_sql_query("select commit_hash from repo_tags_2 where repo_url = %s and tag = %s", params=[repo,tag], con=engine)
    return commitId['commit_hash'][0]

def getLinesOfCodeForCommitId(commitId, engine, filterType=""):
    if filterType == 'production':
        result = pd.read_sql_query("select SUM((extra->>'code')::integer) as loc from results_com_returntocorp_filter_sloc_0_1_1 where commit_hash = %s and check_id = 'JavaScript'", 
                    params=[commitId], con=engine)        
    elif filterType == "test":
         result = pd.read_sql_query("select SUM((extra->>'code')::integer) as loc from results_com_returntocorp_inverse_filter_sloc_0_1_2 where commit_hash = %s and check_id = 'JavaScript'", 
                    params=[commitId], con=engine)           
    else:
        result = pd.read_sql_query("select SUM((extra->>'code')::integer) as loc from results_com_returntocorp_sloc_0_1_1 where commit_hash = %s and check_id = 'JavaScript'", params=
                    [commitId], con=engine)
    loc = result["loc"][0]
    return loc

def getRepoUrlForCommitId(commitId, engine):
    result = pd.read_sql_query("select repo_url from repo_tags_2 where commit_hash = %s", params=[commitId], con=engine)
    return result['repo_url'][0]

def getPackageNameForRepo(repo, engine):
    result = pd.read_sql_query("select package_name from repo_tags_2 where repo_url = %s", params=[repo], con=engine)
    return result['package_name'][0]

def getDependencyChanges(package, dependency, engine):
    result = pd.read_sql_query("select * from npm_dependencies where package = %s and dependency = %s", params=[package, dependency], con=engine)
    return result

import datetime

def findLatestVersionOnDate(year, month, repo, engine):
    date = datetime.datetime(year, month, 1)
    dateStr = date.strftime('%Y-%m-%d')
    result = pd.read_sql_query("select commit_hash from repo_tags_2 where repo_url = %s AND author_timestamp <= %s order by author_timestamp desc limit 1", params=[repo, dateStr], con=engine)
    commitHash = result["commit_hash"]
    if commitHash.empty:
        return None
    return commitHash[0]
    
from concurrent.futures import ThreadPoolExecutor,ProcessPoolExecutor, as_completed
    
# expects a function that takes a repo as first argument
def runInParallel(function, repos, completionFunc):
    pool = ProcessPoolExecutor(8)
    futures = []

    for index, row in repos.iterrows():
        repo = row['repo_url']
        futures.append(pool.submit(function, repo, index))

    for x in as_completed(futures):
        completionFunc(x)
        
# expects a function that takes a repo as first argument
def runInParallelCommitProcessing(function, repos, countFunc, normalizeFunc, completionFunc, filtered=True):
    pool = ProcessPoolExecutor(8)
    futures = []

    for index, row in repos.iterrows():
        repo = row['repo_url']
        progressString = 'Progress: {}/{}'.format(index + 1, len(repos))
        futures.append(pool.submit(function, repo, countFunc, normalizeFunc, progressString, filtered))

    for x in as_completed(futures):
        completionFunc(x)