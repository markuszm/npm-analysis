SELECT main
FROM packages
WHERE name = "aws-sdk";

SELECT l.package, l.type, d.package, dl.type
FROM (SELECT package, type FROM license WHERE type LIKE "GPL%") as l
             JOIN (SELECT package, name
      FROM dependencies) as d JOIN (SELECT package, type FROM license) as dl on d.name = l.package AND d.package = dl.package
WHERE l.type <> dl.type AND dl.type NOT LIKE "%GPL%"
