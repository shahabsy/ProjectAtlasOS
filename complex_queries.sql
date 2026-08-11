-- Complex Query Testing for Atlas OS Graph DB
-- All queries use graph.db schema

-- 1. List all scenes with their hierarchy counts
SELECT s.name as scene_name,
       (SELECT COUNT(*) FROM edges e1 WHERE e1.source = s.id AND e1.relationship = 'CONTAINS') as gameobject_count,
       (SELECT COUNT(*) FROM nodes n2 JOIN edges e2 ON e2.target = n2.id WHERE e2.source IN (SELECT target FROM edges WHERE source = s.id AND relationship = 'CONTAINS')) as component_count
FROM nodes s WHERE type = 'scene';

-- 2. Find GameObjects with scripts in a scene
SELECT DISTINCT g.name as gameobject_name, s.name as script_name
FROM edges e_scene JOIN nodes s_node ON s_node.id = e_scene.source
JOIN nodes g ON g.id = e_scene.target AND e_scene.relationship = 'CONTAINS'
JOIN edges e_comp ON e_comp.source = g.id AND e_comp.relationship = 'HAS_COMPONENT'
JOIN nodes c ON c.id = e_comp.target AND e_comp.source = g.id
LEFT JOIN edges e_script ON e_script.source = c.id AND e_script.relationship = 'USES_SCRIPT'
LEFT JOIN nodes s ON s.id = e_script.target AND s.type = 'script'
WHERE s_node.type = 'scene';

-- 3. Prefab instances
SELECT p.name as prefab_name, g.name as instance_name
FROM edges e_inst JOIN nodes p ON p.id = e_inst.target AND p.type = 'prefab'
JOIN nodes g ON g.id = e_inst.source AND g.type = 'gameobject';

-- 4. Component usage statistics
SELECT c.name as component_type, COUNT(*) as usage_count
FROM nodes c WHERE type = 'component' GROUP BY c.name ORDER BY usage_count DESC;

-- 5. Asset references by component
SELECT c.name as component_name, e.relationship, COUNT(*) as reference_count
FROM edges e JOIN nodes c ON c.id = e.source WHERE c.type = 'component' GROUP BY c.id, e.relationship;

-- 6. Orphaned node check
SELECT 'Nodes without type' as issue, COUNT(*) as count FROM nodes WHERE type IS NULL OR type = '';
