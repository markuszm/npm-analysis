SELECT count(*) FROM dependencies;
SELECT count(*) FROM packages

SELECT count(*)
FROM dependencyChanges
WHERE changeType = "INITIAL";

SELECT count(distinct package)
from dependencyChanges
WHERE changeType = "REMOVED";

SELECT count(distinct package)
FROM versionChanges;

SELECT min(releaseTime)
FROM dependencyChanges;

SELECT
  A.dependency,
  A.addedCount,
  R.removedCount
FROM
  (SELECT
     dependency,
     count(id) as addedCount
   FROM dependencyChanges
   WHERE changeType = "ADDED"
   GROUP BY dependency) as A
  INNER JOIN
  ((SELECT
      dependency,
      count(id) as removedCount
    FROM dependencyChanges
    WHERE changeType = "REMOVED"
    GROUP BY dependency)) as R
    ON A.dependency = R.dependency
ORDER BY addedCount DESC, removedCount DESC
LIMIT 25;

SELECT
      dependency,
      count(id) as removedCount
    FROM dependencyChanges
    WHERE changeType = "REMOVED"
    GROUP BY dependency
    ORDER BY removedCount DESC;

SELECT D.*
FROM dependencyChanges D
  INNER JOIN (SELECT package
              FROM popularity
              WHERE package <> ""
              ORDER BY overall DESC
              LIMIT 5000) as P ON D.package = P.package
WHERE D.dependency = "tough-cookie"