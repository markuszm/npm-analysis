SELECT year, month, name, count FROM maintainerCount GROUP BY year,month,name, count ORDER BY count DESC;

SELECT year, month, name, count FROM maintainerCount ORDER BY name, year, month;

ALTER TABLE maintainerCount ADD INDEX `name_order`(name,year,month);

SELECT count(distinct name) FROM maintainerCount WHERE name <> "" order by name

EXPLAIN SELECT year, month, count FROM maintainerCount WHERE name = "jdalton"


SELECT package FROM dependencies WHERE name = "@matzkoh/slack-outgoing-textlint";