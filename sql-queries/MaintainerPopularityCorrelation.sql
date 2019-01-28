SELECT
        -- For Population
        (avg(x * y) - avg(x) * avg(y)) /
        (sqrt(avg(x * x) - avg(x) * avg(x)) * sqrt(avg(y * y) - avg(y) * avg(y)))
        AS correlation_coefficient_population
  FROM
(SELECT m.package, P.overall as x, m.maintainers as y
FROM (SELECT
        count(distinct name) as maintainers,
        package
      FROM maintainers
      GROUP BY package) as m
INNER JOIN (SELECT package, overall FROM popularity WHERE package <> "" ORDER BY overall DESC) as P
ON m.package = P.package
GROUP BY m.package, P.overall) as t;