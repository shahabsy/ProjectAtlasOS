using UnityEngine;
using UnityEditor;
using UnityEditor.SceneManagement;
using System.IO;
using System.Collections.Generic;
using System.Threading.Tasks;

namespace Atlas.Kernel.Editor
{
    public class SceneIndexer : AssetPostprocessor
    {
        // ─── Phase 0: Auto-index on import ──────────────────────────────
        static void OnPostprocessAllAssets(
            string[] importedAssets,
            string[] deletedAssets,
            string[] movedAssets,
            string[] movedFromAssetPaths)
        {
            foreach (string path in importedAssets)
            {
                if (!path.EndsWith(".unity")) continue;

                Debug.Log($"[Atlas] Detected scene import: {path}");
                string guid = AssetDatabase.AssetPathToGUID(path);
                Debug.Log($"[Atlas] Scene GUID: {guid} -> index into graph.db");
                IndexScene(path);
            }

            foreach (string path in deletedAssets)
            {
                string guid = AssetDatabase.AssetPathToGUID(path);
                Debug.Log($"[Atlas] Scene removed: {path} (guid: {guid})");
            }
        }

        static void IndexScene(string path)
        {
            string guid = AssetDatabase.AssetPathToGUID(path);
            string projectRoot = Application.dataPath.Replace("/Assets", "");
            string dbPath = Path.Combine(projectRoot, ".atlas", "graph.db");

            if (!File.Exists(dbPath))
            {
                Debug.LogWarning($"[Atlas] No graph.db found at {dbPath} - Run `atlas init` first.");
                return;
            }

            // Configurable atlas executable path (EditorPrefs)
            string atlasExePath = EditorPrefs.GetString("Atlas.ExePath", @"E:\ProjectAtlasOS\atlas.exe");
            if (string.IsNullOrEmpty(atlasExePath) || !File.Exists(atlasExePath))
            {
                Debug.LogWarning($"[Atlas] atlas.exe not found at '{atlasExePath}'. Set path with EditorPrefs key 'Atlas.ExePath'.");
                return;
            }

            string sceneName = Path.GetFileNameWithoutExtension(path);

            // Run external process off the main thread to avoid blocking the Editor UI
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
                            Debug.Log($"[Atlas] Indexed scene as node '{sceneName}' (guid: {guid})");
                        }
                        else
                        {
                            Debug.LogError($"[Atlas] Failed to index scene: (exit {exitCode}).\nstdout: {stdout}\nstderr: {stderr}");
                        }
                    };
                }
                catch (System.Exception ex)
                {
                    UnityEditor.EditorApplication.delayCall += () =>
                    {
                        Debug.LogError($"[Atlas] SQLite integration error: {ex.Message}");
                    };
                }
            });
        }

        // ─── Phase 1: Full Scene Hierarchy Indexing ──────────────────────

        [MenuItem("Atlas/Index Full Scene")]
        public static void IndexFullScene()
        {
            var scene = EditorSceneManager.GetActiveScene();
            Debug.Log($"[Atlas] Active scene name: '{scene.name}'");
            Debug.Log($"[Atlas] Active scene path: '{scene.path}'");
            Debug.Log($"[Atlas] Active scene isDirty: {scene.isDirty}");
            Debug.Log($"[Atlas] Active scene is valid: {scene.IsValid()}");

            if (!scene.IsValid())
            {
                Debug.LogError("[Atlas] No active scene found.");
                return;
            }
            // Check if scene is saved
            if (string.IsNullOrEmpty(scene.path))
            {
                Debug.LogError("[Atlas] The scene has not been saved. Please save the scene and try again.");
                return;
            }

            if (scene.isDirty)
            {
                EditorSceneManager.SaveScene(scene);
            }

            string guid = AssetDatabase.AssetPathToGUID(scene.path);
            if (string.IsNullOrEmpty(guid))
            {
                Debug.LogError($"[Atlas] GUID is empty for scene '{scene.name}' at path '{scene.path}'. Ensure the scene is in the Assets folder.");
                return;
            }

            string json = GetSceneHierarchyJson(scene);
            if (string.IsNullOrEmpty(json))
            {
                Debug.LogError("[Atlas] Failed to generate scene hierarchy JSON.");
                return;
            }
            string tempFile = Path.GetTempFileName() + ".json";
            File.WriteAllText(tempFile, json);

            string projectRoot = Application.dataPath.Replace("/Assets", "");
            string atlasExe = EditorPrefs.GetString("Atlas.ExePath", @"E:\ProjectAtlasOS\atlas.exe");
            if (string.IsNullOrEmpty(atlasExe) || !File.Exists(atlasExe))
            {
                Debug.LogWarning($"[Atlas] atlas.exe not found at '{atlasExe}'. Set EditorPrefs 'Atlas.ExePath' correctly.");
                if (File.Exists(tempFile)) File.Delete(tempFile);
                return;
            }

            Task.Run(() =>
            {
                try
                {
                    var process = new System.Diagnostics.Process();
                    process.StartInfo.FileName = atlasExe;
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
                            Debug.Log("[Atlas] Full scene indexing succeeded.\n" + stdout);
                        }
                        else
                        {
                            Debug.LogError("[Atlas] Full scene indexing failed.\nstdout: " + stdout + "\nstderr: " + stderr);
                        }

                        if (File.Exists(tempFile)) File.Delete(tempFile);
                    };
                }
                catch (System.Exception ex)
                {
                    UnityEditor.EditorApplication.delayCall += () =>
                    {
                        Debug.LogError($"[Atlas] Full scene indexing error: {ex.Message}");
                        if (File.Exists(tempFile)) File.Delete(tempFile);
                    };
                }
            });
        }

        // Serializable payload classes so UnityEngine.JsonUtility can serialize correctly
        [System.Serializable]
        class ScenePayload
        {
            public string name;
            public string guid;
            public List<GObject> gameObjects;
        }

        [System.Serializable]
        class GObject
        {
            public string id;
            public string name;
            public string tag;
            public int layer;
            public string parent_id;
            public List<ComponentInfo> components;
        }

        [System.Serializable]
        class ComponentInfo
        {
            public string type;
            public bool enabled;
        }

        static string GetSceneHierarchyJson(UnityEngine.SceneManagement.Scene scene)
        {
            string scenePath = scene.path;
            string guid = AssetDatabase.AssetPathToGUID(scenePath);
            if (string.IsNullOrEmpty(guid))
            {
                Debug.LogError($"[Atlas] failed to get GUID for scene '{scene.name}' at path '{scenePath}'. Make sure the scene is saved.");
                return null;
            }

            var rootObjects = scene.GetRootGameObjects();
            var gobjs = new List<GObject>();

            foreach (var root in rootObjects)
            {
                TraverseGameObject(root, null, gobjs);
            }

            var payload = new ScenePayload
            {
                name = scene.name,
                guid = guid,
                gameObjects = gobjs
            };

            return JsonUtility.ToJson(payload, true);
        }

        static void TraverseGameObject(GameObject go, string parentId, List<GObject> list)
        {
            string id = "gobj_" + go.GetInstanceID().ToString("X");
            list.Add(new GObject
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
            var comps = go.GetComponents<Component>();
            var list = new List<ComponentInfo>();
            foreach (var comp in comps)
            {
                if (comp == null) continue;
                var type = comp.GetType().Name;
                bool enabled = true;
                if (comp is Behaviour b) enabled = b.enabled;
                list.Add(new ComponentInfo
                {
                    type = type,
                    enabled = enabled
                });
            }
            return list;
        }
    }
}