package unity

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/shahabsy/ProjectAtlasOS/internal/utils"
)

func DetectUnityPath() (string, error) {
	switch runtime.GOOS {
	case "windows":
		return detectWindows()
	case "darwin":
		return detectMacOS()
	default:
		return "", fmt.Errorf("Linux not supported in Phase 0")
	}
}

// Windows: Check running Unity process first
func detectWindows() (string, error) {
	cmd := exec.Command("powershell", "-Command",
		`Get-Process -Name 'Unity' -ErrorAction SilentlyContinue | Select-Object -ExpandProperty Path`)
	out, err := cmd.Output()
	if err == nil && len(out) > 0 {
		return strings.TrimSpace(string(out)), nil
	}

	// Common install paths
	paths := []string{
		filepath.Join(os.Getenv("ProgramFiles"), "Unity", "Editor", "Unity.exe"),
		filepath.Join(os.Getenv("ProgramFiles(x86)"), "Unity", "Editor", "Unity.exe"),
	}

	for _, p := range paths {
		if utils.Exists(p) {
			return p, nil
		}
	}
	return "", fmt.Errorf("Unity Editor not found. Install  >= 2022.3 LTS")
}

func detectMacOS() (string, error) {
	cmd := exec.Command("mdfind", "kMDItemCFBundleIdentifier == 'com.unity.UnityEditor'")
	out, err := cmd.Output()
	if err != nil || len(out) == 0 {
		return "", fmt.Errorf("mdfind failed or no Unity found.")
	}

	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		appPath := filepath.Join(line, "Contents", "MacOS", "Unity")
		if utils.Exists(appPath) {
			return appPath, nil
		}
	}
	// Fallback: /Applications/Unity-*.app
	matches, _ := filepath.Glob("/Applications/Unity-*.app")
	for _, m := range matches {
		unityPath := filepath.Join(m, "Contents", "MacOS", "Unity")
		if utils.Exists(unityPath) {
			return unityPath, nil
		}
	}
	return "", fmt.Errorf("Unity Editor not found. Install  >= 2022.3 LTS")
}
