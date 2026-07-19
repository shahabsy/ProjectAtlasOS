using UnityEditor;
using UnityEngine;
using System.IO;
using System.Collections.Generic;
using UnityEditor.SceneManagement;
using System.Threading.Tasks;
using Unity.VisualScripting.YamlDotNet.Core.Tokens;
//using System.Diagnostics;

namespace Atlas.Kernel.Editor
{
    public class SceneIndexer : AssetPostprocessor
    {
        // ──────────────── PHASE 0: Auto-Index on Import ─────────────────

        static void OnPostprocessAllAssets(
            string[] importedAssets,
            string[] deletedAssets,
            string[] movedAssets,
            string[] movedFromAssetPaths)
        {
            foreach (string path in importedAssets)
            {
                if (!path.EndsWith(".unity")) continue;

                // ✅ Get GUID from .meta file - ALWAYS reliable
                string guid = GetGUIDFromMeta(path);
                if (string.IsNullOrEmpty(guid)) return; // Error already logged

                IndexScene(path, guid);
            }

            foreach (string path in deletedAssets)
            {
                Debug.Log($"[Atlas] Scene removed: {path}");
            }
        }

        static void IndexScene(string path, string guid)
        {
            string projectRoot = Application.dataPath.Replace("/Assets", "");
            string dbPath = Path.Combine(projectRoot, ".atlas", "graph.db");

            if (!File.Exists(dbPath))
            {
                Debug.LogWarning($"[Atlas] No graph.db found. Run `atlas init` first.");
                return;
            }

            string atlasExePath = EditorPrefs.GetString("Atlas.ExePath", @"E:\ProjectAtlasOS\atlas.exe");
            if (string.IsNullOrEmpty(atlasExePath) || !File.Exists(atlasExePath))
            {
                Debug.LogWarning($"[Atlas] atlas.exe not found. Set path in Atlas/Settings.");
                return;
            }

            string sceneName = Path.GetFileNameWithoutExtension(path);

            Task.Run(() =>
            {
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

                    string stdout = process.StandardOutput.ReadToEnd();
                    string stderr = process.StandardError.ReadToEnd();
                    process.WaitForExit();

                    int exitCode = process.ExitCode;
                    UnityEditor.EditorApplication.delayCall += () =>
                    {
                        if (exitCode == 0)
                        {
                            Debug.Log($"[Atlas] Indexed scene: '{sceneName}' (guid: {guid})");
                        }
                        else
                        {
                            Debug.LogError($"[Atlas] Index failed: {stderr}");
                        }
                    };
                }
                catch (System.Exception ex)
                {
                    UnityEditor.EditorApplication.delayCall += () =>
                    {
                        Debug.LogError($"[Atlas] Index error: {ex.Message}");
                    };
                }
            });
        }

        // ──────────────── PHASE 1: Full Scene Hierarchy Indexing ────────

        [MenuItem("Atlas/Index Full Scene")]
        public static void IndexFullScene()
        {
            var scene = EditorSceneManager.GetActiveScene();

            if (!scene.IsValid())
            {
                Debug.LogError("[Atlas] No active scene found.");
                return;
            }

            // Ensure scene is saved
            if (scene.isDirty)
            {
                EditorSceneManager.SaveScene(scene);
                Debug.Log($"[Atlas] Saved unsaved changes in '{scene.name}'.");
            }

            string scenePath = scene.path;
            string guid = GetGUIDFromMeta(scenePath);
            if (string.IsNullOrEmpty(guid))
            {
                Debug.LogError($"[Atlas] Failed to get GUID for '{scene.name}' at path '{scenePath}'.");
                return;
            }

            string json = GenerateSceneJson(scene, guid);
            if (json == null)
            {
                Debug.LogError("[Atlas] Failed to generate scene hierarchy JSON.");
                return;
            }

            RunCLI(scene.name, json); // Pass scene name for logging
        }

        static string GetGUIDFromMeta(string scenePath)
        {
            // Normalize path to Assets/ format (Unity uses forward slashes)
            string relativePath = scenePath.Replace("Assets/", "");

            // Construct meta file path (handles Windows/macOS/Linux paths)
            string metaPath = Path.Combine(Application.dataPath, relativePath + ".meta");

            if (!File.Exists(metaPath))
            {
                Debug.LogError($"[Atlas] .meta file not found: {metaPath}");
                return null;
            }

            try
            {
                var lines = File.ReadAllLines(metaPath);
                foreach (var line in lines)
                {
                    if (line.StartsWith("guid:"))
                        return line.Substring(5).Trim(); // "guid: d20ebab..." → "d20ebab..."
                }
                Debug.LogError($"[Atlas] No GUID found in .meta file: {scenePath}");
            }
            catch (System.Exception ex)
            {
                Debug.LogError($"[Atlas] Failed to read .meta file: {ex.Message}");
            }

            return null;
        }
        [System.Serializable]
        public class ScenePayload
        {
            public string name;
            public string guid;
            public List<GObject> gameObjects;
        }

        [System.Serializable]
        public class GObject
        {
            public string id;
            public string name;
            public string tag;
            public int layer;
            public string parent_id;
            public List<ComponentInfo> components;
        }

        [System.Serializable]
        public class ComponentInfo
        {
            public string type;
            public bool enabled;
        }

        static string GenerateSceneJson(UnityEngine.SceneManagement.Scene scene, string guid)
        {
            var rootObjects = scene.GetRootGameObjects();
            var gobjs = new List<GObject>();

            foreach (var root in rootObjects)
            {
                TraverseGameObject(root, null, gobjs);
            }

            var playload = new ScenePayload
            {
                name = scene.name,
                guid = guid,
                gameObjects = gobjs
            };
            return JsonUtility.ToJson(playload, true);
        }

        static void TraverseGameObject(GameObject go, string parentId, List<GObject> list)
        {
            string id = $"gobj_{go.GetInstanceID():X}";

            list.Add( new GObject
            {
                id = id,
                name = go.name,
                tag = go.tag,
                layer = go.layer,
                parent_id = parentId,
                components = GetComponents(go)
            });

            foreach (Transform child in go.transform)
            {
                TraverseGameObject(child.gameObject, id, list);
            }
        }

        static List<ComponentInfo> GetComponents(GameObject go)
        {
            var list = new List<ComponentInfo>();
            foreach (var comp in go.GetComponents<Component>())
            {
                if (comp == null) continue;
                bool enabled = comp is Behaviour b ? b.enabled : true;
                list.Add(new ComponentInfo
                {
                    type = comp.GetType().Name,
                    enabled = enabled
                });
            }
            return list;
        }

        static void RunCLI(string sceneName, string jsonContent)
        {
            string projectRoot = Application.dataPath.Replace("/Assets", "");
            string atlasExePath = EditorPrefs.GetString("Atlas.ExePath", @"E:\ProjectAtlasOS\atlas.exe");
            if (string.IsNullOrEmpty(atlasExePath) || !File.Exists(atlasExePath))
            {
                Debug.LogError("[Atlas] atlas.exe not found. Set path in Atlas/Settings.");
                return;
            }
            Task.Run(() =>
            {
                string tempFile = Path.GetTempFileName() + ".json";
                try
                {
                    File.WriteAllText(tempFile, jsonContent);
                    Debug.Log($"[Atlas] Writing JSON to: {tempFile}");

                    var process = new System.Diagnostics.Process();
                    process.StartInfo.FileName = atlasExePath;
                    process.StartInfo.Arguments = $"index scene --full --file \"{tempFile}\"";

                    process.StartInfo.WorkingDirectory = projectRoot;
                    process.StartInfo.UseShellExecute = false;
                    process.StartInfo.RedirectStandardOutput = true;
                    process.StartInfo.RedirectStandardError = true;
                    process.Start();

                    string stdout = process.StandardOutput.ReadToEnd();
                    string stderr = process.StandardError.ReadToEnd();
                    process.WaitForExit();

                    int exitCode = process.ExitCode;
                    UnityEditor.EditorApplication.delayCall += () =>
                    {
                        if (exitCode == 0)
                        {
                            Debug.Log($"[Atlas] ✅ Indexed scene '{sceneName}' with full hierarchy");
                        }
                        else
                        {
                            Debug.LogError($"[Atlas] ❌ Indexing failed: {stderr}");
                        }
                    };
                }
                catch (System.Exception ex)
                {
                    UnityEditor.EditorApplication.delayCall += () =>
                    {
                        Debug.LogError($"[Atlas] CLI error: {ex.Message}");
                    };
                }
                finally
                {
                    if (File.Exists(tempFile))
                    {
                        try { File.Delete(tempFile); } catch { }
                    }
                }
            });
        }
    }
}
