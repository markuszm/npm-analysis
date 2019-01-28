SELECT name,version FROM packages

SELECT PV.name, PV.version
FROM (SELECT A.package FROM (SELECT
        package
      FROM dependencies
      WHERE name = "fullname") as A INNER JOIN (SELECT package
                                                       FROM popularity
                                                       WHERE package <> ""
                                                       ORDER BY overall DESC
                                                       LIMIT 50000) as P ON A.package = P.package) AS Dependents INNER JOIN
         (SELECT name, version FROM packages) AS PV ON Dependents.package = PV.name;