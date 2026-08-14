package main

import (
	"fmt"
	"log"

	"github.com/shahabsy/ProjectAtlasOS/internal/query"
	"github.com/shahabsy/ProjectAtlasOS/internal/statistics"
	"github.com/shahabsy/ProjectAtlasOS/internal/tools"
)

func main() {
	dbPath := "D:/PortfolioProject/XtreamShooter/.atlas/graph.db"

	// Initialize Query Engine
	q := query.NewEngine(dbPath)

	// Initialize Statistics Engine
	stats := statistics.NewEngine(q)

	fmt.Println("=== ATLAS STACK TEST ====")

	// ------------------------------------------------------------
	// 1. QUERY ENGINE TESTS
	// ------------------------------------------------------------

	// 1.1 List all scenes
	fmt.Println("1.1 ListScenes()")
	scenes, err := q.ListScenes()
	if err != nil {
		log.Fatalf("ListScenes error: %v", err)
	}
	fmt.Printf("Found %d scenes:\n", len(scenes))
	for _, s := range scenes {
		fmt.Printf("  - ID: %s, Name: %s, GUID: %s\n", s.ID, s.Name, s.GUID)
	}
	fmt.Println()

	// 1.2 Get scene by ID (use the first scene from the list)
	if len(scenes) > 0 {
		sceneID := scenes[0].ID
		fmt.Printf("1.2 GetSceneByID(%q)\n", sceneID)
		scene, err := q.GetSceneByID(sceneID)
		if err != nil {
			log.Fatalf("GetSceneByID error: %v", err)
		}
		if scene != nil {
			fmt.Printf("  Scene: %+v\n", scene)
		} else {
			fmt.Println("  Scene not found")
		}
		fmt.Println()
	}

	// 1.3 Get scene by GUID (hardcode the known GUID)
	sceneGUID := "f4bdba331571635429465ab726720755"
	fmt.Printf("1.3 GetSceneByGUID(%q)\n", sceneGUID)
	sceneByGUID, err := q.GetSceneByGUID(sceneGUID)
	if err != nil {
		log.Fatalf("GetSceneByGUID error: %v", err)
	}
	if sceneByGUID != nil {
		fmt.Printf("  Scene: %+v\n", sceneByGUID)
	} else {
		fmt.Println("  Scene not found")
	}
	fmt.Println()

	// 1.4 Get GameObjects by scene (use first scene ID)
	if len(scenes) > 0 {
		sceneID := scenes[0].ID
		fmt.Printf("1.4 GetGameObjectsByScene(%q)\n", sceneID)
		gobjs, err := q.GetGameObjectsByScene(sceneID)
		if err != nil {
			log.Fatalf("GetGameObjectsByScene error: %v", err)
		}
		fmt.Printf("  Found %d GameObjects\n", len(gobjs))
		// Print first 5
		for i, gobj := range gobjs {
			if i >= 5 {
				fmt.Printf("  ... and %d more\n", len(gobjs)-5)
				break
			}
			fmt.Printf("  - %s (%s)\n", gobj.Name, gobj.ID)
		}
		fmt.Println()
	}

	// 1.5 Find GameObjects by name
	if len(scenes) > 0 {
		sceneID := scenes[0].ID
		searchName := "Camera"
		fmt.Printf("1.5 FindGameObjectsByName(%q, %q)\n", sceneID, searchName)
		gobjs, err := q.FindGameObjectsByName(sceneID, searchName)
		if err != nil {
			log.Fatalf("FindGameObjectsByName error: %v", err)
		}
		fmt.Printf("  Found %d GameObjects containing '%s'\n", len(gobjs), searchName)
		for _, gobj := range gobjs {
			fmt.Printf("  - %s (%s)\n", gobj.Name, gobj.ID)
		}
		fmt.Println()
	}

	// 1.6 Get GameObject by ID (pick a GameObject from the list)
	// We'll get the first GameObject from the scene.
	if len(scenes) > 0 {
		sceneID := scenes[0].ID
		gobjs, err := q.GetGameObjectsByScene(sceneID)
		if err == nil && len(gobjs) > 0 {
			gobjID := gobjs[0].ID
			fmt.Printf("1.6 GetGameObjectByID(%q)\n", gobjID)
			gobj, err := q.GetGameObjectByID(gobjID)
			if err != nil {
				log.Fatalf("GetGameObjectByID error: %v", err)
			}
			if gobj != nil {
				fmt.Printf("  GameObject: %+v\n", gobj)
			} else {
				fmt.Println("  GameObject not found")
			}
			fmt.Println()
		}
	}

	// 1.7 Get Components by GameObject
	if len(scenes) > 0 {
		sceneID := scenes[0].ID
		gobjs, err := q.GetGameObjectsByScene(sceneID)
		if err == nil && len(gobjs) > 0 {
			gobjID := gobjs[0].ID
			fmt.Printf("1.7 GetComponentsByGameObject(%q)\n", gobjID)
			comps, err := q.GetComponentsByGameObject(gobjID)
			if err != nil {
				log.Fatalf("GetComponentsByGameObject error: %v", err)
			}
			fmt.Printf("  Found %d components\n", len(comps))
			for _, comp := range comps {
				fmt.Printf("  - %s (%s) [%s]\n", comp.Type, comp.ID, comp.GlobalID)
			}
			fmt.Println()
		}
	}

	// 1.8 Get Component by ID (pick a component from previous step)
	if len(scenes) > 0 {
		sceneID := scenes[0].ID
		gobjs, err := q.GetGameObjectsByScene(sceneID)
		if err == nil && len(gobjs) > 0 {
			gobjID := gobjs[0].ID
			comps, err := q.GetComponentsByGameObject(gobjID)
			if err == nil && len(comps) > 0 {
				compID := comps[0].ID
				fmt.Printf("1.8 GetComponentByID(%q)\n", compID)
				comp, err := q.GetComponentByID(compID)
				if err != nil {
					log.Fatalf("GetComponentByID error: %v", err)
				}
				if comp != nil {
					fmt.Printf("  Component: %+v\n", comp)
				} else {
					fmt.Println("  Component not found")
				}
				fmt.Println()
			}
		}
	}

	// 1.9 Get Script by ID (we need a script ID; we can get from the component if it has one, or list scripts)
	fmt.Println("1.9 GetScriptByID (using first script from ListScripts)")
	scripts, err := q.ListScripts()
	if err != nil {
		log.Fatalf("ListScripts error: %v", err)
	}
	if len(scripts) > 0 {
		scriptID := scripts[0].ID
		script, err := q.GetScriptByID(scriptID)
		if err != nil {
			log.Fatalf("GetScriptByID error: %v", err)
		}
		if script != nil {
			fmt.Printf("  Script: %+v\n", script)
		} else {
			fmt.Println("  Script not found")
		}
	} else {
		fmt.Println("  No scripts found")
	}
	fmt.Println()

	// 1.10 GetScriptsByName
	searchScript := "UIManager"
	fmt.Printf("1.10 GetScriptsByName(%q)\n", searchScript)
	scriptsByName, err := q.GetScriptsByName(searchScript)
	if err != nil {
		log.Fatalf("GetScriptsByName error: %v", err)
	}
	fmt.Printf("  Found %d scripts containing '%s'\n", len(scriptsByName), searchScript)
	for _, s := range scriptsByName {
		fmt.Printf("  - %s (%s)\n", s.Name, s.ID)
	}
	fmt.Println()

	// 1.11 ListScripts
	fmt.Println("1.11 ListScripts()")
	allScripts, err := q.ListScripts()
	if err != nil {
		log.Fatalf("ListScripts error: %v", err)
	}
	fmt.Printf("  Found %d scripts total\n", len(allScripts))
	// print first 5
	for i, s := range allScripts {
		if i >= 5 {
			fmt.Printf("  ... and %d more\n", len(allScripts)-5)
			break
		}
		fmt.Printf("  - %s (%s)\n", s.Name, s.ID)
	}
	fmt.Println()

	// 1.12 GetGameObjectsUsingScript (use a script ID from above)
	if len(allScripts) > 0 {
		scriptID := allScripts[0].ID
		fmt.Printf("1.12 GetGameObjectsUsingScript(%q)\n", scriptID)
		gobjsUsingScript, err := q.GetGameObjectsUsingScript(scriptID)
		if err != nil {
			log.Fatalf("GetGameObjectsUsingScript error: %v", err)
		}
		fmt.Printf("  Found %d GameObjects using script '%s'\n", len(gobjsUsingScript), allScripts[0].Name)
		for _, gobj := range gobjsUsingScript {
			fmt.Printf("  - %s (%s)\n", gobj.Name, gobj.ID)
		}
		fmt.Println()
	}

	// 1.13 GetAssetByID (pick an asset from ListAssets)
	assets, err := q.ListAssets()
	if err != nil {
		log.Fatalf("ListAssets error: %v", err)
	}
	if len(assets) > 0 {
		assetID := assets[0].ID
		fmt.Printf("1.13 GetAssetByID(%q)\n", assetID)
		asset, err := q.GetAssetByID(assetID)
		if err != nil {
			log.Fatalf("GetAssetByID error: %v", err)
		}
		if asset != nil {
			fmt.Printf("  Asset: %+v\n", asset)
		} else {
			fmt.Println("  Asset not found")
		}
		fmt.Println()
	}

	// 1.14 GetAssetsByType (try "Texture")
	assetType := "Texture"
	fmt.Printf("1.14 GetAssetsByType(%q)\n", assetType)
	assetsByType, err := q.GetAssetsByType(assetType)
	if err != nil {
		log.Fatalf("GetAssetsByType error: %v", err)
	}
	fmt.Printf("  Found %d assets of type '%s'\n", len(assetsByType), assetType)
	for _, a := range assetsByType {
		fmt.Printf("  - %s (%s)\n", a.Name, a.ID)
	}
	fmt.Println()

	// 1.15 ListAssets
	fmt.Println("1.15 ListAssets()")
	allAssets, err := q.ListAssets()
	if err != nil {
		log.Fatalf("ListAssets error: %v", err)
	}
	fmt.Printf("  Found %d assets total\n", len(allAssets))
	for i, a := range allAssets {
		if i >= 5 {
			fmt.Printf("  ... and %d more\n", len(allAssets)-5)
			break
		}
		fmt.Printf("  - %s (%s)\n", a.Name, a.ID)
	}
	fmt.Println()

	// 1.16 Traversal: WalkParents
	// Use a GameObject from the scene
	if len(scenes) > 0 {
		sceneID := scenes[0].ID
		gobjs, err := q.GetGameObjectsByScene(sceneID)
		if err == nil && len(gobjs) > 0 {
			// pick a child if possible, but we'll just use the first one and see parents
			gobjID := gobjs[0].ID
			fmt.Printf("1.16 WalkParents(%q)\n", gobjID)
			parents, err := q.WalkParents(gobjID)
			if err != nil {
				log.Fatalf("WalkParents error: %v", err)
			}
			if len(parents) == 0 {
				fmt.Println("  No parents found (root GameObject)")
			} else {
				fmt.Printf("  Parents chain: %v\n", parents)
			}
			fmt.Println()
		}
	}

	// 1.17 Traversal: WalkChildren
	if len(scenes) > 0 {
		sceneID := scenes[0].ID
		// Use scene ID itself to get children (GameObjects)
		fmt.Printf("1.17 WalkChildren(%q) (scene)\n", sceneID)
		children, err := q.WalkChildren(sceneID)
		if err != nil {
			log.Fatalf("WalkChildren error: %v", err)
		}
		fmt.Printf("  Found %d children (GameObjects) under scene\n", len(children))
		// Show first 5 child IDs
		for i, child := range children {
			if i >= 5 {
				fmt.Printf("  ... and %d more\n", len(children)-5)
				break
			}
			fmt.Printf("  - %s\n", child)
		}
		fmt.Println()
	}

	// 1.18 Traversal: GetNeighbors
	// Pick a node (e.g., scene) and get neighbors via CONTAINS relationship
	if len(scenes) > 0 {
		sceneID := scenes[0].ID
		relationship := "CONTAINS"
		fmt.Printf("1.18 GetNeighbors(%q, %q)\n", sceneID, relationship)
		neighbors, err := q.GetNeighbors(sceneID, relationship)
		if err != nil {
			log.Fatalf("GetNeighbors error: %v", err)
		}
		fmt.Printf("  Found %d neighbors via '%s'\n", len(neighbors), relationship)
		for i, n := range neighbors {
			if i >= 5 {
				fmt.Printf("  ... and %d more\n", len(neighbors)-5)
				break
			}
			fmt.Printf("  - %s\n", n)
		}
		fmt.Println()
	}

	// 1.19 WalkDependencies (use a script or asset)
	if len(allScripts) > 0 {
		scriptID := allScripts[0].ID
		fmt.Printf("1.19 WalkDependencies(%q) (script)\n", scriptID)
		deps, err := q.WalkDependencies(scriptID)
		if err != nil {
			log.Fatalf("WalkDependencies error: %v", err)
		}
		fmt.Printf("  Found %d dependencies (nodes that depend on this script)\n", len(deps))
		// Show some IDs
		for i, dep := range deps {
			if i >= 5 {
				fmt.Printf("  ... and %d more\n", len(deps)-5)
				break
			}
			fmt.Printf("  - %s\n", dep)
		}
		fmt.Println()
	}

	// 1.20 WalkReverseDependencies (use a GameObject)
	if len(scenes) > 0 {
		sceneID := scenes[0].ID
		gobjs, err := q.GetGameObjectsByScene(sceneID)
		if err == nil && len(gobjs) > 0 {
			gobjID := gobjs[0].ID
			fmt.Printf("1.20 WalkReverseDependencies(%q) (GameObject)\n", gobjID)
			revDeps, err := q.WalkReverseDependencies(gobjID)
			if err != nil {
				log.Fatalf("WalkReverseDependencies error: %v", err)
			}
			fmt.Printf("  Found %d reverse dependencies (nodes this GameObject depends on)\n", len(revDeps))
			for i, dep := range revDeps {
				if i >= 5 {
					fmt.Printf("  ... and %d more\n", len(revDeps)-5)
					break
				}
				fmt.Printf("  - %s\n", dep)
			}
			fmt.Println()
		}
	}

	// ------------------------------------------------------------
	// 2. STATISTICS ENGINE TESTS
	// ------------------------------------------------------------

	fmt.Println("=== STATISTICS ENGINE TESTS ===")

	// 2.1 ProjectSummary
	fmt.Println("2.1 ProjectSummary()")
	projectSummary, err := stats.ProjectSummary()
	if err != nil {
		log.Fatalf("ProjectSummary error: %v", err)
	}
	fmt.Printf("Project Summary: %+v\n", projectSummary)
	fmt.Println()

	// 2.2 SceneSummary (for the first scene)
	if len(scenes) > 0 {
		sceneID := scenes[0].ID
		fmt.Printf("2.2 SceneSummary(%q)\n", sceneID)
		sceneStats, err := stats.SceneSummary(sceneID)
		if err != nil {
			log.Fatalf("SceneSummary error: %v", err)
		}
		fmt.Printf("Scene Statistics: %+v\n", sceneStats)
		fmt.Println()
	}

	fmt.Println("=== ALL TESTS COMPLETED ===")

	// 3. TOOL RUNTIME TESTS
	fmt.Println("\n=== TOOL RUNTIME TESTS ===")

	// Reuse the existing engines
	toolCtx := tools.NewContext(q, stats) // q is *query.Engine, stats is *statistics.Engine
	registry := tools.NewToolRegistry(toolCtx)

	// 3.1 ListScenes
	scenesResp, err := registry.Scene.ListScenes()
	if err != nil {
		log.Fatalf("ListScenes: %v", err)
	}
	fmt.Printf("3.1 ListScenes: %d scenes\n", len(scenesResp.Scenes))

	// 3.2 DescribeScene
	if len(scenesResp.Scenes) > 0 {
		sceneID := scenesResp.Scenes[0].ID
		describeResp, err := registry.Scene.DescribeScene(sceneID)
		if err != nil {
			log.Fatalf("DescribeScene: %v", err)
		}
		fmt.Printf("3.2 DescribeScene: %s | GO=%d | Scripts=%d | Warnings=%v\n",
			describeResp.Scene.Name,
			describeResp.Statistics.GameObjectCount,
			describeResp.Statistics.ScriptCount,
			describeResp.Warnings)
	}

	// 3.3 FindGameObjects
	if len(scenesResp.Scenes) > 0 {
		findResp, err := registry.GameObject.FindGameObjects(tools.FindGameObjectsRequest{
			SceneID: scenesResp.Scenes[0].ID,
			Name:    "Camera",
		})
		if err != nil {
			log.Fatalf("FindGameObjects: %v", err)
		}
		fmt.Printf("3.3 FindGameObjects('Camera'): %d found\n", len(findResp.GameObjects))

		// 3.4 GetHierarchy
		if len(findResp.GameObjects) > 0 {
			hierResp, err := registry.GameObject.GetGameObjectHierarchy(findResp.GameObjects[0].ID)
			if err != nil {
				log.Fatalf("GetHierarchy: %v", err)
			}
			fmt.Printf("3.4 Hierarchy(%s): children=%d parents=%d | Warnings=%v\n",
				hierResp.GameObject.Name,
				len(hierResp.Children),
				len(hierResp.Parents),
				hierResp.Warnings)
		}
	}

	// 3.5 FindScripts
	scriptResp, err := registry.Script.FindScripts(tools.FindScriptsRequest{Name: "UIManager"})
	if err != nil {
		log.Fatalf("FindScripts: %v", err)
	}
	fmt.Printf("3.5 FindScripts('UIManager'): %d found\n", len(scriptResp.Scripts))

	// 3.6 GetGameObjectsUsingScript
	if len(scriptResp.Scripts) > 0 {
		usageResp, err := registry.Script.GetGameObjectsUsingScript(tools.GetGameObjectsUsingScriptRequest{
			ScriptID: scriptResp.Scripts[0].ID,
		})
		if err != nil {
			log.Fatalf("GetGameObjectsUsingScript: %v", err)
		}
		fmt.Printf("3.6 GetGameObjectsUsingScript: %d GameObjects\n", len(usageResp.GameObjects))
	}

	// 3.7 ListAssets
	assetsResp, err := registry.Asset.ListAssets(tools.ListAssetsRequest{})
	if err != nil {
		log.Fatalf("ListAssets: %v", err)
	}
	fmt.Printf("3.7 ListAssets: %d assets\n", len(assetsResp.Assets))

	// 3.8 ProjectStats
	projStatsResp, err := registry.Stats.ProjectStats(tools.ProjectStatsRequest{})
	if err != nil {
		log.Fatalf("ProjectStats: %v", err)
	}
	fmt.Printf("3.8 ProjectStats: scenes=%d gobjs=%d scripts=%d assets=%d\n",
		projStatsResp.Summary.TotalScenes,
		projStatsResp.Summary.TotalGameObjects,
		projStatsResp.Summary.UniqueScripts,
		projStatsResp.Summary.TotalAssets)

	// 3.9 SceneStats
	if len(scenesResp.Scenes) > 0 {
		sceneStatsResp, err := registry.Stats.SceneStats(tools.SceneStatsRequest{
			SceneID: scenesResp.Scenes[0].ID,
		})
		if err != nil {
			log.Fatalf("SceneStats: %v", err)
		}
		fmt.Printf("3.9 SceneStats: GO=%d Scripts=%d\n",
			sceneStatsResp.Statistics.GameObjectCount,
			sceneStatsResp.Statistics.ScriptCount)
	}
}
