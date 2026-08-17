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
	// 1. QUERY ENGINE TESTS (unchanged, these work with (data, error) returns)
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
	if len(scenes) > 0 {
		sceneID := scenes[0].ID
		gobjs, err := q.GetGameObjectsByScene(sceneID)
		if err == nil && len(gobjs) > 0 {
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
		fmt.Printf("1.17 WalkChildren(%q) (scene)\n", sceneID)
		children, err := q.WalkChildren(sceneID)
		if err != nil {
			log.Fatalf("WalkChildren error: %v", err)
		}
		fmt.Printf("  Found %d children (GameObjects) under scene\n", len(children))
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
	// 2. STATISTICS ENGINE TESTS (unchanged)
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

	// ------------------------------------------------------------
	// 3. TOOL RUNTIME TESTS (updated to use the current tools.Result API)
	// ------------------------------------------------------------

	fmt.Println("\n=== TOOL RUNTIME TESTS ===")

	// Reuse the existing engines
	toolCtx := tools.NewContext(q, stats) // q is *query.Engine, stats is *statistics.Engine
	registry := tools.NewToolRegistry(toolCtx)

	// 3.1 ListScenes
	sceneResult := registry.Scene.ListScenes()
	if !sceneResult.Success {
		log.Fatalf("ListScenes failed: %s", sceneResult.Error.Message)
	}
	scenesResp, ok := sceneResult.Data.(tools.ListScenesResponse)
	if !ok {
		log.Fatalf("ListScenes: unexpected response type")
	}
	fmt.Printf("3.1 ListScenes: %d scenes\n", len(scenesResp.Scenes))

	// 3.2 DescribeScene
	if len(scenesResp.Scenes) > 0 {
		sceneID := scenesResp.Scenes[0].ID
		descResult := registry.Scene.DescribeScene(sceneID)
		if !descResult.Success {
			log.Fatalf("DescribeScene failed: %s", descResult.Error.Message)
		}
		describeResp, ok := descResult.Data.(tools.DescribeSceneResponse)
		if !ok {
			log.Fatalf("DescribeScene: unexpected response type")
		}
		fmt.Printf("3.2 DescribeScene: %s | GO=%d | Scripts=%d | Warnings=%v\n",
			describeResp.Scene.Name,
			describeResp.Statistics.GameObjectCount,
			describeResp.Statistics.ScriptCount,
			describeResp.Warnings)
	}

	// 3.3 ListGameObjects (first scene)
	if len(scenesResp.Scenes) > 0 {
		sceneID := scenesResp.Scenes[0].ID
		goResult := registry.GameObject.ListGameObjects(sceneID)
		if !goResult.Success {
			log.Fatalf("ListGameObjects failed: %s", goResult.Error.Message)
		}
		goList, ok := goResult.Data.(tools.ListGameObjectsResponse)
		if !ok {
			log.Fatalf("ListGameObjects: unexpected response type")
		}
		fmt.Printf("3.3 ListGameObjects: %d GameObjects in scene\n", len(goList.GameObjects))
	}

	// 3.4 ListComponents (use first GameObject)
	if len(scenesResp.Scenes) > 0 {
		sceneID := scenesResp.Scenes[0].ID
		goResult := registry.GameObject.ListGameObjects(sceneID)
		if goResult.Success {
			goList, ok := goResult.Data.(tools.ListGameObjectsResponse)
			if ok && len(goList.GameObjects) > 0 {
				gobjID := goList.GameObjects[0].ID
				compResult := registry.Component.ListComponents(gobjID)
				if !compResult.Success {
					log.Fatalf("ListComponents failed: %s", compResult.Error.Message)
				}
				compList, ok := compResult.Data.(tools.ListComponentsResponse)
				if !ok {
					log.Fatalf("ListComponents: unexpected response type")
				}
				fmt.Printf("3.4 ListComponents: %d components on GameObject\n", len(compList.Components))
			}
		}
	}

	// 3.5 ListScripts (first scene)
	if len(scenesResp.Scenes) > 0 {
		sceneID := scenesResp.Scenes[0].ID
		scriptResult := registry.Script.ListScripts(sceneID)
		if !scriptResult.Success {
			log.Fatalf("ListScripts failed: %s", scriptResult.Error.Message)
		}
		scriptList, ok := scriptResult.Data.(tools.ListScriptsResponse)
		if !ok {
			log.Fatalf("ListScripts: unexpected response type")
		}
		fmt.Printf("3.5 ListScripts: %d scripts in scene\n", len(scriptList.Scripts))
	}

	// 3.6 ListAssets
	assetResult := registry.Asset.ListAssets()
	if !assetResult.Success {
		log.Fatalf("ListAssets failed: %s", assetResult.Error.Message)
	}
	assetList, ok := assetResult.Data.(tools.ListAssetsResponse)
	if !ok {
		log.Fatalf("ListAssets: unexpected response type")
	}
	fmt.Printf("3.6 ListAssets: %d assets\n", len(assetList.Assets))

	// 3.7 ProjectStats
	projResult := registry.Stats.ProjectStats()
	if !projResult.Success {
		log.Fatalf("ProjectStats failed: %s", projResult.Error.Message)
	}
	projStats, ok := projResult.Data.(tools.ProjectStatsResponse)
	if !ok {
		log.Fatalf("ProjectStats: unexpected response type")
	}
	fmt.Printf("3.7 ProjectStats: scenes=%d gobjs=%d scripts=%d assets=%d\n",
		projStats.Summary.TotalScenes,
		projStats.Summary.TotalGameObjects,
		projStats.Summary.UniqueScripts,
		projStats.Summary.TotalAssets)

	// 3.8 SceneStats
	if len(scenesResp.Scenes) > 0 {
		sceneID := scenesResp.Scenes[0].ID
		sceneStatResult := registry.Stats.SceneStats(sceneID)
		if !sceneStatResult.Success {
			log.Fatalf("SceneStats failed: %s", sceneStatResult.Error.Message)
		}
		sceneStats, ok := sceneStatResult.Data.(tools.SceneStatsResponse)
		if !ok {
			log.Fatalf("SceneStats: unexpected response type")
		}
		fmt.Printf("3.8 SceneStats: GO=%d Scripts=%d\n",
			sceneStats.Statistics.GameObjectCount,
			sceneStats.Statistics.ScriptCount)
	}
}
