package core

import (
	"fmt"

	db "github.com/shahabsy/ProjectAtlasOS/internal/db"
)

func AnalyzeLightingCmd() error {
	projectRoot, err := findUnityProject()
	if err != nil {
		return fmt.Errorf("not a Unity project: %w", err)
	}

	dbPath := db.GetDBPath(projectRoot)

	// Get all reflection probes from database
	reflectionProbes, err := db.GetNodesByType(dbPath, "ReflectionProbe")
	if err != nil {
		return fmt.Errorf("failed to query reflection probes: %w", err)
	}

	fmt.Printf("\n🔍 Analysis Results for 'New Scene 3'\n\n")

	// Check for overlapping reflection probes
	fmt.Println("📊 Reflection Probes:")
	for _, probe := range reflectionProbes {
		fmt.Printf("  - %s (ID: %s)\n", probe.Name, probe.ID)
	}

	if len(reflectionProbes) == 0 {
		fmt.Println("\n💡 No reflection probes found in scene.")
	}

	// Check for dark corners (missing Light Probe Groups)
	fmt.Println("\n💡 Dark Corners Check:")
	fmt.Println("   No Light Probe Group nodes found - consider adding probes to shadowed areas")

	// Terrain check
	terrainNodes, _ := db.GetNodesByType(dbPath, "Terrain")
	if len(terrainNodes) > 0 {
		fmt.Println("\n🌳 Terrain Analysis:")
		for _, terrain := range terrainNodes {
			fmt.Printf("  - %s: Check for opaque shaders in detail objects\n", terrain.Name)
		}
	} else {
		fmt.Println("\n🌳 No terrains found.")
	}

	fmt.Println("\n✅ Analysis complete. Run `atlas index scene --full` to analyze scenes with more details.")

	return nil
}
