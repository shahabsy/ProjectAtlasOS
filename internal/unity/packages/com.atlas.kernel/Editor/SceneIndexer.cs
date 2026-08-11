using UnityEditor;
using UnityEngine;
using UnityEditor.SceneManagement;
using UnityEngine.SceneManagement;
using System.IO;
using System.Collections.Generic;
using System.Threading.Tasks;

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

                Debug.Log($"[Atlas] Detected scene import: {path}");
                string guid = GetGUIDFromMeta(path);
                if (string.IsNullOrEmpty(guid))
                {
                    Debug.LogWarning($"[Atlas] Skipping indexing for {path} — GUID not found.");
                    continue;
                }
                Debug.Log($"[Atlas] Scene GUID: {guid} -> index into graph.db");
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

            string atlasExePath = EditorPrefs.GetString("Atlas.ExePath", @"E:\GameDevAI_Agent\ProjectAtlasOS\atlas.exe");
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
                    EditorApplication.delayCall += () =>
                    {
                        if (exitCode == 0)
                            Debug.Log($"[Atlas] Indexed scene: '{sceneName}' (guid: {guid})");
                        else
                            Debug.LogError($"[Atlas] Index failed: {stderr}");
                    };
                }
                catch (System.Exception ex)
                {
                    EditorApplication.delayCall += () =>
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
            if (string.IsNullOrEmpty(json))
            {
                Debug.LogError("[Atlas] Failed to generate scene hierarchy JSON.");
                return;
            }

            RunCLI(scene.name, json);
        }

        static string GetGUIDFromMeta(string scenePath)
        {
            string relativePath = scenePath;
            if (relativePath.StartsWith("Assets/"))
                relativePath = relativePath.Substring("Assets/".Length);

            relativePath = relativePath.Replace('/', Path.DirectorySeparatorChar);
            string metaPath = Path.Combine(Application.dataPath, relativePath + ".meta");

            if (!File.Exists(metaPath))
            {
                Debug.LogError($"[Atlas] .meta file not found: {metaPath}");
                return null;
            }

            try
            {
                var lines = File.ReadAllLines(metaPath);
                foreach (var raw in lines)
                {
                    var line = raw.Trim();
                    if (line.StartsWith("guid:"))
                        return line.Substring("guid:".Length).Trim();
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
            public string type = "scene";
            public List<GObject> gameObjects;
        }

        [System.Serializable]
        public class GObject
        {
            public string id;
            public string globalId;
            public string type = "gameobject";
            public string name;
            public string tag;
            public int layer;
            public string parent_id;
            public string prefab_guid;
            public List<ComponentInfo> components;
        }

        [System.Serializable]
        public class ComponentInfo
        {
            public string globalId;
            public string type;
            public bool enabled;
            public string parent_id;

            // ─── Script Resolution ──────────────────────────────────────────
            public string script_guid;       // MUST be snake_case to match JSON
            public string class_name;
            public string namespace_name;

            // ─── Serialized Fields ──────────────────────────────────────────
            public List<SerializedFieldInfo> serialized_fields;

            // ─── Asset References ───────────────────────────────────────────
            public List<AssetReferenceInfo> asset_references;
        }

        [System.Serializable]
        public class SerializedFieldInfo
        {
            public string name;
            public string type;
            public string value;
            public string reference_type;
            public string reference_id;
            public string reference_path;
        }

        [System.Serializable]
        public class AssetReferenceInfo
        {
            public string type;
            public string guid;
            public string name;
            public string slot_name;
        }

        static string GenerateSceneJson(Scene scene, string sceneGuid)
        {
            var rootObjects = scene.GetRootGameObjects();
            var gobjs = new List<GObject>();

            foreach (var root in rootObjects)
            {
                TraverseGameObject(root, null, gobjs, sceneGuid);
            }

            var payload = new ScenePayload
            {
                name = scene.name,
                guid = sceneGuid,
                gameObjects = gobjs
            };
            return JsonUtility.ToJson(payload, true);
        }

        static void TraverseGameObject(GameObject go, string parentId, List<GObject> list, string sceneGuid)
        {
            string globalId = GetStableGameObjectId(go);

            var gobj = new GObject
            {
                globalId = globalId,
                name = go.name,
                tag = go.tag,
                layer = go.layer,
                parent_id = parentId,
                components = GetComponents(go, globalId)
            };

            // Detect Prefab Instance
            var prefabAsset = PrefabUtility.GetCorrespondingObjectFromSource(go);
            if (prefabAsset != null)
            {
                string prefabPath = AssetDatabase.GetAssetPath(prefabAsset);
                gobj.prefab_guid = AssetDatabase.AssetPathToGUID(prefabPath);
            }

            list.Add(gobj);

            foreach (Transform child in go.transform)
            {
                TraverseGameObject(child.gameObject, globalId, list, sceneGuid);
            }
        }

        static string GetStableGameObjectId(GameObject go)
        {
            return GlobalObjectId.GetGlobalObjectIdSlow(go).ToString();
        }

        // ─── ENHANCED COMPONENT EXTRACTION ──────────────────────────────

        static List<ComponentInfo> GetComponents(GameObject go, string parentGlobalId)
        {
            var list = new List<ComponentInfo>();

            foreach (var comp in go.GetComponents<Component>())
            {
                if (comp == null) continue;

                string globalId = GlobalObjectId.GetGlobalObjectIdSlow(comp).ToString();
                bool enabled = comp is Behaviour b ? b.enabled : true;

                var compInfo = new ComponentInfo
                {
                    globalId = globalId,
                    type = comp.GetType().Name,
                    enabled = enabled,
                    parent_id = parentGlobalId,
                    serialized_fields = new List<SerializedFieldInfo>(),
                    asset_references = new List<AssetReferenceInfo>()
                };

                // ─── MonoBehaviour: Script + Serialized Fields ──────────────
                if (comp is MonoBehaviour mono)
                {
                    var script = MonoScript.FromMonoBehaviour(mono);
                    if (script != null)
                    {
                        string scriptPath = AssetDatabase.GetAssetPath(script);
                        string scriptGuid = AssetDatabase.AssetPathToGUID(scriptPath);
                        if (!string.IsNullOrEmpty(scriptGuid))
                        {
                            compInfo.script_guid = scriptGuid;
                            compInfo.class_name = script.GetClass().Name;
                            compInfo.namespace_name = script.GetClass().Namespace ?? "";
                            Debug.Log($"[Atlas] Found script: {compInfo.class_name} (GUID: {scriptGuid})");
                        }
                        else
                        {
                            // Fallback: use class name as synthetic GUID
                            compInfo.script_guid = $"script_{script.GetClass().Name}";
                            compInfo.class_name = script.GetClass().Name;
                            compInfo.namespace_name = script.GetClass().Namespace ?? "";
                            Debug.LogWarning($"[Atlas] Script GUID not found for '{compInfo.class_name}', using synthetic ID.");
                        }
                    }
                    else
                    {
                        Debug.LogWarning($"[Atlas] MonoScript is null for component {comp.GetType().Name} on {go.name}");
                    }

                    // ─── Serialized Fields ──────────────────────────────────
                    var so = new SerializedObject(comp);
                    var iterator = so.GetIterator();
                    while (iterator.NextVisible(true))
                    {
                        if (iterator.propertyType == SerializedPropertyType.ObjectReference)
                        {
                            var objRef = iterator.objectReferenceValue;
                            if (objRef != null)
                            {
                                string refPath = AssetDatabase.GetAssetPath(objRef);
                                string refGuid = AssetDatabase.AssetPathToGUID(refPath);
                                compInfo.serialized_fields.Add(new SerializedFieldInfo
                                {
                                    name = iterator.name,
                                    type = objRef.GetType().Name,
                                    reference_type = "asset",
                                    reference_id = refGuid,
                                    reference_path = refPath
                                });
                            }
                        }
                        else
                        {
                            // Handle primitive types
                            var fieldInfo = new SerializedFieldInfo
                            {
                                name = iterator.name,
                                type = iterator.type
                            };
                            switch (iterator.propertyType)
                            {
                                case SerializedPropertyType.Float:
                                    fieldInfo.value = iterator.floatValue.ToString();
                                    break;
                                case SerializedPropertyType.Integer:
                                    fieldInfo.value = iterator.intValue.ToString();
                                    break;
                                case SerializedPropertyType.Boolean:
                                    fieldInfo.value = iterator.boolValue.ToString();
                                    break;
                                case SerializedPropertyType.String:
                                    fieldInfo.value = iterator.stringValue;
                                    break;
                                case SerializedPropertyType.Enum:
                                    fieldInfo.value = iterator.enumValueIndex.ToString();
                                    break;
                                case SerializedPropertyType.Vector2:
                                    fieldInfo.value = iterator.vector2Value.ToString();
                                    break;
                                case SerializedPropertyType.Vector3:
                                    fieldInfo.value = iterator.vector3Value.ToString();
                                    break;
                                case SerializedPropertyType.Vector4:
                                    fieldInfo.value = iterator.vector4Value.ToString();
                                    break;
                                case SerializedPropertyType.Rect:
                                    fieldInfo.value = iterator.rectValue.ToString();
                                    break;
                                case SerializedPropertyType.Bounds:
                                    fieldInfo.value = iterator.boundsValue.ToString();
                                    break;
                                case SerializedPropertyType.Color:
                                    fieldInfo.value = iterator.colorValue.ToString();
                                    break;
                                case SerializedPropertyType.LayerMask:
                                    fieldInfo.value = iterator.intValue.ToString();
                                    break;
                                default:
                                    fieldInfo.value = "Unsupported";
                                    break;
                            }
                            compInfo.serialized_fields.Add(fieldInfo);
                        }
                    }
                }

                // ─── Renderer: Materials ──────────────────────────────────────
                if (comp is Renderer renderer)
                {
                    var materials = renderer.sharedMaterials;
                    foreach (var mat in materials)
                    {
                        if (mat == null) continue;
                        string matPath = AssetDatabase.GetAssetPath(mat);
                        string matGuid = AssetDatabase.AssetPathToGUID(matPath);
                        compInfo.asset_references.Add(new AssetReferenceInfo
                        {
                            type = "material",
                            guid = matGuid,
                            name = mat.name
                        });
                    }
                }

                list.Add(compInfo);
            }

            return list;
        }

        // ─── Helper (unused but kept) ───────────────────────────────────

        static string GetHierarchyPath(GameObject go)
        {
            var parts = new List<string>();
            Transform t = go.transform;
            while (t != null)
            {
                parts.Insert(0, t.name);
                t = t.parent;
            }
            return string.Join("/", parts);
        }

        // ─── CLI Execution ──────────────────────────────────────────────

        static void RunCLI(string sceneName, string jsonContent)
        {
            string projectRoot = Application.dataPath.Replace("/Assets", "");
            string atlasExePath = EditorPrefs.GetString("Atlas.ExePath", @"E:\GameDevAI_Agent\ProjectAtlasOS\atlas.exe");
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
                    EditorApplication.delayCall += () =>
                    {
                        Debug.Log($"[Atlas] stdout:\n{stdout}");
                        if (!string.IsNullOrEmpty(stderr))
                            Debug.LogError($"[Atlas] stderr:\n{stderr}");
                        if (exitCode == 0)
                            Debug.Log($"[Atlas] Indexed scene '{sceneName}' with full hierarchy");
                        else
                            Debug.LogError($"[Atlas] Indexing failed: {stderr}");

                        bool keep = EditorPrefs.GetBool("Atlas.KeepTempJson", true);
                        if (!keep && File.Exists(tempFile)) File.Delete(tempFile);
                        else if (keep)
                            Debug.Log($"[Atlas] Kept temp JSON at: {tempFile} for inspection.");
                    };
                }
                catch (System.Exception ex)
                {
                    EditorApplication.delayCall += () =>
                    {
                        Debug.LogError($"[Atlas] CLI error: {ex.Message}");
                    };
                }
                finally
                {
                    bool keep = EditorPrefs.GetBool("Atlas.KeepTempJson", true);
                    if (!keep && File.Exists(tempFile))
                    {
                        try { File.Delete(tempFile); } catch { }
                    }
                }
            });
        }
    }
}