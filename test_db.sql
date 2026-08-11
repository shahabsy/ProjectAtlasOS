-- Atlas OS - SQLite Testing Script
.headers on
.mode column

SELECT '=== SCHEMA VALIDATION ===';
SELECT name FROM sqlite_master WHERE type='table';

SELECT '=== NODE COUNTS BY TYPE ===';
SELECT type, COUNT(*) as count FROM nodes GROUP BY type;

SELECT '=== SCENES ===';
SELECT id, name, global_id FROM nodes WHERE type='scene';

SELECT '=== GAMEOBJECTS ===';
SELECT id, name, global_id FROM nodes WHERE type='gameobject' LIMIT 10;

SELECT '=== COMPONENTS ===';
SELECT id, global_id, name FROM nodes WHERE type='component' LIMIT 10;

SELECT '=== SCRIPTS ===';
SELECT id, name, guid FROM nodes WHERE type='script';

SELECT '=== EDGES BY RELATIONSHIP ===';
SELECT relationship, COUNT(*) as count FROM edges GROUP BY relationship ORDER BY count DESC;

SELECT '=== PROJECT STATS ===';
 SELECT (SELECT COUNT(*) FROM nodes WHERE type='scene') as scenes, (SELECT COUNT(*) FROM nodes WHERE type='gameobject') as gameobjects, (SELECT COUNT(*) FROM nodes WHERE type='component') as components;

SELECT '=== SAMPLE EDGE DATA ===';
SELECT source, target, relationship FROM edges LIMIT 10;
