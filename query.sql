SELECT *
FROM (
    (
        SELECT row, value FROM cells WHERE column = 0
    ) AS c0
    LEFT JOIN(
        SELECT row, value FROM cells WHERE column = 1
    ) AS c1 USING (row)
)
WHERE rowid >= 23
LIMIT 10
UNION
SELECT *
FROM (
    (
        SELECT row, value FROM cells WHERE column = 0
    ) AS c0
    LEFT JOIN(
        SELECT row, value FROM cells WHERE column = 1
    ) AS c1 USING (row)
)
WHERE rowid >= 23
LIMIT 10;
