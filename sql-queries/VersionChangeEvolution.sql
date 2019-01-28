# average release frequency over all version changes
SELECT avg(releaseFrequency)
FROM (SELECT
        package,
        avg(timeDiff) AS releaseFrequency
      FROM versionChanges
      WHERE timeDiff > 0
      GROUP BY package) as perPackage INNER JOIN (SELECT package
                                                  FROM popularity
                                                  WHERE package <> ""
                                                  ORDER BY overall DESC
                                                  LIMIT 50000) as P ON perPackage.package = P.package;

# Distribution of version change types
SELECT count(*)
FROM (SELECT
        package,
        versionDiff
      FROM versionChanges
      WHERE versionDiff = "prerelease") as A INNER JOIN (SELECT package
                                                       FROM popularity
                                                       WHERE package <> ""
                                                       ORDER BY overall DESC
                                                       LIMIT 50000) as P ON A.package = P.package;

# Average version releases per package
SELECT avg(A.c)
FROM (SELECT package
      FROM popularity
      WHERE package <> ""
      ORDER BY overall DESC
      LIMIT 500) as P
  INNER JOIN
  (SELECT
     package,
     count(*) as c
   FROM versionChanges
   WHERE versionDiff = "minor"
   GROUP BY package) as A ON A.package = P.package;

# Average minor between majors per package
SELECT avg(perPackage.val)
FROM (SELECT
        package,
        avgPatchBetweenMinor as val
      FROM versionCount) as perPackage INNER JOIN (SELECT package
                                                                FROM popularity
                                                                WHERE package <> ""
                                                                ORDER BY overall DESC
                                                                LIMIT 5000) as P ON perPackage.package = P.package

select * from versionChanges where package = 'atlassian-connect-express'