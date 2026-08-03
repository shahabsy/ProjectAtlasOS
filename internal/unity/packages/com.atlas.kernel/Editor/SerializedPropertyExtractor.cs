using UnityEditor;
using UnityEngine;
using System.Collections.Generic;

namespace Atlas.Kernel.Editor
{
    public static class SerializedPropertyExtractor
    {
        public static List<SerializedFieldInfo> ExtractFilds(MonoBehaviour mono)
        {
            var result = new List<SerializedFieldInfo>();
            var so = new SerializedObject(mono);
            var iterator = so.GetIterator();

            while (iterator.NextVisible(true))
            {
                var info = new SerializedFieldInfo
                {
                    name = iterator.name,
                    type = iterator.type,
                    propertyType = iterator.propertyType
                };

                switch (iterator.propertyType)
                {
                    case SerializedPropertyType.ObjectReference:
                        var objRef = iterator.objectReferenceValue;
                        if (objRef != null)
                        {
                            string refPath = AssetDatabase.GetAssetPath(objRef);
                            info.reference_type = "asset";
                            info.reference_id = AssetDatabase.AssetPathToGUID(refPath);
                            info.reference_path = refPath;
                        }
                        break;
                    case SerializedPropertyType.Float:
                    case SerializedPropertyType.Integer:
                    case SerializedPropertyType.String:
                        info.value = iterator.stringValue;
                        break;
                    // Add other property types as needed (Vector2, Vector3, etc.)
                }
                result.Add(info);
            }
            return result;
        }
    }

    [System.Serializable]
    public class SerializedFieldInfo
    {
        public string name;
        public string type;
        public SerializedPropertyType propertyType; // need to make sure this is working
        public string value;
        public string reference_type;
        public string reference_id;
        public string reference_path;
    }
}