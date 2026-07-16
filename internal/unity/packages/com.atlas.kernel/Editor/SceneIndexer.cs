using UnityEditor;
using UnityEngine;
using System.IO;

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
            }
        }
        static void IndexScene(string path)
        {
            // For now: log the GUID to console
            string  guid = AssetDatabase.AssetPathToGUID(path);
            Debug.Log($"[Atlas] Scene GUID: {guid} -> Would insert into graph.db");
        }
    }
}