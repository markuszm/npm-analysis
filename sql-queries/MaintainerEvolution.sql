SELECT name, count(package)
FROM maintainers
GROUP BY name
ORDER BY count(package) DESC

SELECT
  A.name,
  A.count - B.count as diff
from maintainerCount A, maintainerCount B
  where A.year = 2017 AND A.month = 1 AND B.year = 2016 AND B.month = 1 AND A.name = B.name
ORDER BY diff ASC;

SELECT year,month, avg(count)
FROM maintainerCount
GROUP BY year, month;

SELECT avg(m.maintainers)
FROM (SELECT
        count(distinct name) as maintainers,
        package
      FROM maintainerChanges
      WHERE changeType = "ADDED"
      GROUP BY package) as m;

SELECT count(lifeTimeMaintainers.package)
FROM (SELECT
        package,
        count(id) as count
      FROM maintainerChanges
      WHERE changeType = "REMOVED"
      GROUP BY package) as lifeTimeMaintainers;

SELECT count(lifeTimeMaintainers.package)
FROM (SELECT
        package,
        count(id) as count
      FROM maintainerChanges
      WHERE changeType = "REMOVED"
      GROUP BY package) as lifeTimeMaintainers INNER JOIN (SELECT package
                                         FROM popularity
                                         WHERE package <> ""
                                         ORDER BY overall DESC
                                         LIMIT 50000) as P ON lifeTimeMaintainers.package = P.package;

SELECT avg(m.maintainers)
FROM (SELECT
        count(distinct name) as maintainers,
        package
      FROM maintainerChanges
      WHERE changeType = "INITIAL" OR changeType = "ADDED"
      GROUP BY package) as m INNER JOIN (SELECT package
                                         FROM popularity
                                         WHERE package <> ""
                                         ORDER BY overall DESC
                                         LIMIT 50000) as P ON m.package = P.package

SELECT *
FROM maintainerCount
WHERE year = 2018 AND month = 4
ORDER BY count DESC LIMIT 20

