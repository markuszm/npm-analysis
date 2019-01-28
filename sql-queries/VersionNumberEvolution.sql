SELECT count(distinct package)
FROM versionCount WHERE patch > 0;

SELECT count(distinct package)
FROM versionCount;

SELECT count(distinct package)
FROM versionChanges WHERE versionDiff = "publish";

SELECT *
FROM versionChanges
WHERE package = "@anilanar/workbox-build"

SELECT count(distinct package)
FROM versionCount
WHERE major > 0 AND (minor > 0 OR patch > 0)

SELECT package, version
FROM versionChanges
WHERE versionDiff = "publish" AND version LIKE "%-%"

SELECT package,version
FROM versionChanges
WHERE versionDiff = "prerelease"