SELECT
  R.dependency AS Removed,
  A.dependency AS Added,
  count(*)
FROM dependencyChanges R, dependencyChanges A, (SELECT
                                                  package,
                                                  version
                                                FROM (
                                                       SELECT
                                                         package,
                                                         version,
                                                         count(*) as changeCount
                                                       FROM dependencyChanges
                                                       GROUP BY package, version) as A
                                                WHERE A.changeCount = 2) Filter
WHERE
  R.version NOT LIKE "%-%" AND R.changeType = "REMOVED" AND A.changeType = "ADDED" AND Filter.package = R.package AND
  R.package = A.package AND Filter.version = R.version AND R.version = A.version
GROUP BY R.dependency, A.dependency
ORDER BY count(*) DESC;

SELECT DISTINCT R.dependency
FROM dependencyChanges R, dependencyChanges A
WHERE R.package = A.package AND R.version = A.version AND R.version NOT LIKE "%-%" AND R.changeType = "REMOVED" AND
      A.changeType = "ADDED";


# packages that got added and removed at some point
SELECT DISTINCT A.dependency
FROM dependencyChanges A, dependencyChanges R
WHERE R.package = A.package AND R.dependency = A.dependency AND R.releaseTime > A.releaseTime AND R.version NOT LIKE "%-%" AND R.changeType = "REMOVED" AND
      A.changeType = "ADDED";

SELECT *
FROM dependencyChanges
ORDER BY package, version

SELECT
  package,
  version
FROM (
       SELECT
         package,
         version,
         count(*) as changeCount
       FROM dependencyChanges
       GROUP BY package, version) as A
WHERE A.changeCount = 2;

ALTER TABLE dependencyChanges
  ADD INDEX (package, version);
ALTER TABLE dependencyChanges
  ADD INDEX (dependency);


SELECT *
FROM maintainerCount
WHERE count > 0
ORDER BY year, month;

SELECT *
FROM versionChanges
WHERE releaseTime <> "1970-01-01 00:00:01"
ORDER BY releaseTime;

SELECT
  L.changeString,
  count(L.changeString)
FROM licenseChange
  as L INNER JOIN (SELECT package
                   FROM popularity
                   WHERE package <> ""
                   ORDER BY overall DESC
                   ) as P ON L.package = P.package
GROUP BY changeString
ORDER BY count(changeString) DESC;

SELECT count(distinct name)
FROM packages;

SELECT count(distinct package)
FROM maintainers;

SELECT *
FROM (SELECT distinct package
      FROM maintainers) as A RIGHT JOIN (SELECT name
                                        FROM packages) as B ON A.package = B.name
WHERE A.package IS NULL;

SELECT packages.name,packages.description, maintainers.name, packages.version
FROM packages INNER JOIN maintainers ON packages.name = maintainers.package
WHERE version = "0.0.1-security";

SELECT package
FROM maintainerChanges
WHERE name = "ehsalazar" AND changeType = "ADDED"

SELECT name, version
FROM packages
WHERE name like "@0%" LIMIT 20

select name, version from packages where name <> "" order by name;

select count(name) from packages where name <> "" order by name;

SELECT * FROM popularity WHERE package <> "" ORDER BY overall DESC

SELECT DISTINCT name from dependencies