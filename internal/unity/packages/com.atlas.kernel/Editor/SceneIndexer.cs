using UnityEditor;
using UnityEngine;
using System.IO;
using System;

namespace Atlas.Kernel.Editor
{
    public class SceneIndexer : AssetPostprocessor
    {
        static void OnPostprocessAllAssets(
            string[] importedAssets,
            string[] deletedAssets,
            string[] movedAssets,
            string[] movedFromAssetPaths)
        {
            foreach(string path in importedAssets)
            {
                if(!path.EndsWith(".unity")) continue;

                Debug.Log($"[Atlas] Detected scene import: {path}");
                
                // Phase 0: Just log GUID for now - full indexing later
                string guid = AssetDatabase.AssetPathToGUID(path);
                Debug.Log($"[Atlas] Scene GUID: {guid} -> index into graph.db");
                IndexScene(path);
            }
            // Also handle deleted/moved assets
            foreach(string path in deletedAssets)
            {
                string guid = AssetDatabase.AssetPathToGUID(path);
                Debug.Log($"[Atlas] Scene removed: {path} (guid: {guid})");
            }
        }
        static void IndexScene(string path)
        {
            // For now: log the GUID to console
            string  guid = AssetDatabase.AssetPathToGUID(path);
            Debug.Log($"[Atlas] Scene GUID: {guid} -> Would insert into graph.db");

            //Fix: use fill path to atlas.exe
            string atlasExePath = @"E:\ProjectAtlasOS\atlas.exe";

            // Get project root and DB path
            string projectRoot = Application.dataPath.Replace("Assets", "");
            string dbPath = Path.Combine(projectRoot, ".atlas", "graph.db");

            if (!File.Exists(dbPath))
            {
                Debug.Log($"[Atlas] No graph.db found at {dbPath} - Run `atlas init` first.");
                return;
            }

            // Step 1: Extract scene name (e.g., "Jungle_01" from "Assets/Scenes/Jungle_01.unity")
            string sceneName = Path.GetFileNameWithoutExtension(path);

            // Step 2: Write node to SQLite (via external command in Phase 0)
            try
            {
                var process = new System.Diagnostics.Process();
                process.StartInfo.FileName = atlasExePath;
                process.StartInfo.Arguments = $"index node --type=scene --guid={guid} --name=\"{sceneName}\"";
                process.StartInfo.WorkingDirectory = projectRoot;
                process.StartInfo.UseShellExecute = false;
                process.StartInfo.RedirectStandardOutput = true;
                process.StartInfo.RedirectStandardError = true;
                process.Start();

                // Optional: Wait for completion and capture output
                string stdout = process.StandardOutput.ReadToEnd();
                string stderr = process.StandardError.ReadToEnd();
                process.WaitForExit();

                if (process.ExitCode == 0)
                {
                    Debug.Log($"[Atlas] Indexed scene as node '{sceneName}' (guid: {guid})");
                }
                else
                {
                    Debug.LogError($"[Atlas] Failed to index scene: (exit {process.ExitCode})" +
                        $".\n stdout: {stdout}\n stderr{stderr}");
                }
            }
            catch (System.Exception ex)
            {
                Debug.LogError($"[Atlas] SQLite integration error: {ex.Message}");
            }
        }

        // Helper for future commands
        static void IndexNode(string type, string guid, string name)
        {
            // TODO: Implement the logic to index a node in the database.
            // Future: Call internal C# functions instead of external process
        }
    }
}