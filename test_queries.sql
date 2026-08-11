-- ============================================================================
-- Atlas OS - SQLite Test Database Queries
-- Validates graph.db structure, data, and all query methods
-- ============================================================================

-- -----------------------------------------------------------------------------
-- SECTION 1: SCHEMA VALIDATION
-- -----------------------------------------------------------------------------

-- Verify tables exist
SELECT name FROM sqlite_master WHERE type='table' ORDER BY name;

-- Verify indexes exist
SELECT name FROM sqlite_master WHERE type='index' AND name LIKE 'idx_%' ORDER BY name;

-- Check table structures
SELECT sql FROM sqlite_master WHERE type='table' AND name IN ('nodes', 'edges');

-- -----------------------------------------------------------------------------
-- SECTION 2: BASIC NODE COUNTS & OVERVIEW
-- -----------------------------------------------------------------------------

-- Count total nodes by type
SELECT 
    type,
    COUNT(*) as count
FROM nodes 
GROUP BY type
ORDER BY count DESC;

-- List all unique relationship types in edges
SELECT DISTINCT relationship FROM edges ORDER BY relationship;

-- Show node types distribution with percentages
SELECT 
    type,
    COUNT(*) as count,
    ROUND(COUNT(*) * 100.0 / (SELECT COUNT(*) FROM nodes), 2) as percentage
FROM nodes 
GROUP BY type
ORDER BY count DESC;

-- -----------------------------------------------------------------------------
-- SECTION 3: SCENE QUERIES
-- -----------------------------------------------------------------------------

-- List all scenes with metadata
SELECT 
    id,
    name,
    guid,
    datetime(created_at, 'unixepoch') as created_at,
    datetime(updated_at, 'unixepoch') as updated_at,
    datetime(last_indexed_at, 'unixepoch') as last_indexed_at
FROM nodes 
WHERE type = 'scene'
ORDER BY last_indexed_at DESC;

-- Find scene by GUID (exact match)
SELECT * FROM nodes WHERE type = 'scene' AND guid = 'YOUR_SCENE_GUID_HERE';

-- Find scene by name (fuzzy)
SELECT id, name, guid FROM nodes WHERE type = 'scene' AND name LIKE '%Main%';

-- -----------------------------------------------------------------------------
-- SECTION 4: SCENE STATISTICS
-- -----------------------------------------------------------------------------

-- Count scenes in database
SELECT COUNT(*) as total_scenes FROM nodes WHERE type = 'scene';

-- List all GameObjects with scene association
SELECT 
    g.name as gameobject_name,
    s.name as scene_name,
    s.guid as scene_guid
FROM edges e_scene
JOIN nodes s ON s.id = e_scene.source AND s.type = 'scene'
JOIN nodes g ON g.id = e_scene.target AND g.type = 'gameobject'
WHERE e_scene.relationship = 'CONTAINS';

-- -----------------------------------------------------------------------------
-- SECTION 5: GAMEOBJECT QUERIES
-- -----------------------------------------------------------------------------

-- List all GameObjects with parent relationships
SELECT 
    n.id,
    n.global_id,
    n.name,
    COALESCE(e2.source, '') as parent_global_id
FROM nodes n
JOIN edges e1 ON e1.target = n.id AND e1.relationship = 'CONTAINS'
LEFT JOIN edges e2 ON e2.target = n.id AND e2.relationship = 'CHILD_OF'
WHERE n.type = 'gameobject'
ORDER BY n.name;

-- Count GameObjects per scene
SELECT 
    s.name as scene_name,
    COUNT(*) as gameobject_count
FROM nodes s
JOIN edges e ON e.source = s.id AND e.relationship = 'CONTAINS'
JOIN nodes g ON g.id = e.target AND g.type = 'gameobject'
WHERE s.type = 'scene'
GROUP BY s.id, s.name;

-- -----------------------------------------------------------------------------
-- SECTION 6: COMPONENT QUERIES
-- -----------------------------------------------------------------------------

-- List all components with their associated GameObjects
SELECT 
    c.global_id,
    c.name as component_type,
    c.enabled,
    g.name as gameobject_name,
    COALESCE(e_parent.source, '') as parent_gameobject_id
FROM nodes c
JOIN edges e_comp ON e_comp.target = c.id AND e_comp.relationship = 'HAS_COMPONENT'
JOIN nodes g ON g.id = e_comp.source AND g.type = 'gameobject'
LEFT JOIN edges e_parent ON e_parent.target = g.id AND e_parent.relationship = 'CHILD_OF'
WHERE c.type = 'component';

-- Count components per GameObject
SELECT 
    g.name as gameobject_name,
    COUNT(c.id) as component_count
FROM nodes g
JOIN edges e_comp ON e_comp.source = g.id AND e_comp.relationship = 'HAS_COMPONENT'
JOIN nodes c ON c.id = e_comp.target AND c.type = 'component'
GROUP BY g.id, g.name
ORDER BY component_count DESC;

-- List scripts used in components (via USES_SCRIPT edge)
SELECT 
    s.name as script_name,
    s.guid as script_guid,
    c.global_id as component_global_id,
    c.name as component_type
FROM nodes s
JOIN edges e_script ON e_script.target = s.id AND e_script.relationship = 'USES_SCRIPT'
JOIN nodes c ON c.id = e_script.source AND c.type = 'component';

-- -----------------------------------------------------------------------------
-- SECTION 7: SCRIPT QUERIES
-- -----------------------------------------------------------------------------

-- List all scripts in project
SELECT 
    id,
    guid,
    name as script_name,
    path
FROM nodes WHERE type = 'script'
ORDER BY name;

-- Find script by GUID
SELECT * FROM nodes WHERE type = 'script' AND guid IS NOT NULL ORDER BY name;

-- Count components using each script
SELECT 
    s.name as script_name,
    COUNT(c.id) as component_count
FROM nodes s
JOIN edges e_script ON e_script.target = s.id AND e_script.relationship = 'USES_SCRIPT'
JOIN nodes c ON c.id = e_script.source AND c.type = 'component'
GROUP BY s.id, s.name
ORDER BY component_count DESC;

-- -----------------------------------------------------------------------------
-- SECTION 8: PREFAB QUERIES
-- -----------------------------------------------------------------------------

-- List all prefabs in project
SELECT 
    id,
    guid,
    name,
    path
FROM nodes WHERE type = 'prefab'
ORDER BY name;

-- Find prefab by GUID (if any exist)
SELECT * FROM nodes WHERE type = 'prefab' AND guid IS NOT NULL;

-- List GameObject instances linked to prefabs
SELECT 
    g.name as instance_name,
    g.global_id,
    p.name as prefab_name,
    e_inst.relationship
FROM edges e_inst
JOIN nodes g ON g.id = e_inst.source AND g.type = 'gameobject'
JOIN nodes p ON p.id = e_inst.target AND p.type = 'prefab';

-- -----------------------------------------------------------------------------
-- SECTION 9: ASSET QUERIES
-- -----------------------------------------------------------------------------

-- List all assets by type
SELECT 
    id,
    guid,
    type,
    name,
    path
FROM nodes WHERE type IN ('material', 'shader', 'texture')
ORDER BY type, name;

-- Find asset by GUID pattern (if any exist)
SELECT * FROM nodes WHERE type IN ('material', 'shader', 'texture') AND guid IS NOT NULL LIMIT 10;

-- -----------------------------------------------------------------------------
-- SECTION 10: EDGE/RELATIONSHIP QUERIES
-- -----------------------------------------------------------------------------

-- List all edges with full context
SELECT 
    source_node.name as source_name,
    source_node.type as source_type,
    e.relationship,
    target_node.name as target_name,
    target_node.type as target_type
FROM edges e
JOIN nodes source_node ON source_node.id = e.source
JOIN nodes target_node ON target_node.id = e.target;

-- Count edges by relationship type
SELECT 
    relationship,
    COUNT(*) as count
FROM edges
GROUP BY relationship
ORDER BY count DESC;

-- Find all scenes (source nodes with CONTAINS relationships)
SELECT DISTINCT
    source_node.name as scene_name,
    source_node.global_id,
    COUNT(target_node.id) as child_count
FROM edges e
JOIN nodes source_node ON source_node.id = e.source AND source_node.type = 'scene'
JOIN nodes target_node ON target_node.id = e.target
WHERE e.relationship = 'CONTAINS'
GROUP BY source_node.id, source_node.name, source_node.global_id;

-- -----------------------------------------------------------------------------
-- SECTION 11: DATA INTEGRITY CHECKS
-- -----------------------------------------------------------------------------

-- Check for orphaned edges (source node missing)
SELECT 
    e.source,
    e.relationship,
    COUNT(*) as edge_count
FROM edges e
LEFT JOIN nodes src ON src.id = e.source
WHERE src.id IS NULL
GROUP BY e.source, e.relationship;

-- Check for orphaned edges (target node missing)
SELECT 
    e.target,
    e.relationship,
    COUNT(*) as edge_count
FROM edges e
LEFT JOIN nodes tgt ON tgt.id = e.target
WHERE tgt.id IS NULL
GROUP BY e.target, e.relationship;

-- Check for duplicate global_ids
SELECT global_id, COUNT(*) as count
FROM nodes 
WHERE global_id IS NOT NULL AND global_id != ''
GROUP BY global_id 
HAVING COUNT(*) > 1;

-- Verify all nodes have type assigned
SELECT id, name FROM nodes WHERE type IS NULL OR type = '';

-- Check for empty GUIDs where they should exist
SELECT id, type, name FROM nodes WHERE guid IS NULL ORDER BY type, name LIMIT 20;

-- -----------------------------------------------------------------------------
-- SECTION 12: SAMPLE INSPECTION QUERIES
-- -----------------------------------------------------------------------------

-- Random node sample (for initial inspection)
SELECT id, type, name, global_id, guid FROM nodes LIMIT 50;

-- Random edge sample (for initial inspection)
SELECT source, target, relationship FROM edges LIMIT 50;

-- Show all unique script names (MonoBehaviours)
SELECT DISTINCT s.name as script_name, s.guid
FROM nodes s
WHERE s.type = 'script';

-- -----------------------------------------------------------------------------
-- SECTION 13: PROJECT STATISTICS SUMMARY
-- -----------------------------------------------------------------------------

-- Overall project statistics
SELECT 
    (SELECT COUNT(*) FROM nodes WHERE type = 'scene') as scenes,
    (SELECT COUNT(*) FROM nodes WHERE type = 'gameobject') as gameobjects,
    (SELECT COUNT(*) FROM nodes WHERE type = 'component') as components,
    (SELECT COUNT(*) FROM nodes WHERE type = 'script') as scripts,
    (SELECT COUNT(*) FROM nodes WHERE type = 'prefab') as prefabs,
    (SELECT COUNT(*) FROM nodes WHERE type IN ('material', 'shader', 'texture')) as assets,
    (SELECT COUNT(*) FROM edges) as total_edges;

-- -----------------------------------------------------------------------------
-- END OF TEST QUERIES
-- =============================================================================
