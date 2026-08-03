using UnityEditor;
using UnityEngine;

namespace Atlas.Kernel.Editor
{
    public static class PrefabReferenceExtractor
    {
        public static string GetPrefabGuid(GameObject go)
        {
            var prefabAsset = PrefabUtility.GetCorrespondingObjectFromSource(go);
            if (prefabAsset == null) return null;

            string prefabPath = AssetDatabase.GetAssetPath(prefabAsset);
            return AssetDatabase.AssetPathToGUID(prefabPath);
        }
    }
}