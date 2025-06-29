ALTER TABLE dice ADD COLUMN name VARCHAR(20);

INSERT INTO dice (sides, set_id, name)
SELECT 10, d.set_id, NULL
FROM dice d
LEFT JOIN dice d2 
    ON d2.set_id = d.set_id 
    AND d2.sides = 10 
    AND d2.id != d.id
WHERE d.sides = 10
GROUP BY d.set_id
HAVING COUNT(*) = 1;

UPDATE dice
SET name = 'd' || sides
WHERE sides != 10;

UPDATE dice
SET name = 'd10 (1)'
WHERE sides = 10
AND id IN (
    SELECT MIN(id)
    FROM dice
    WHERE sides = 10
    GROUP BY set_id
);

UPDATE dice
SET name = 'd10 (2)'
WHERE sides = 10
AND id IN (
    SELECT MAX(id)
    FROM dice
    WHERE sides = 10
    GROUP BY set_id
);
