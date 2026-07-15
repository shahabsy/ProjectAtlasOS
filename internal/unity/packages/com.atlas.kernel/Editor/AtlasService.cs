using UnityEditor;
using UnityEngine;

namespace Atlas.Kernel.Editor
{
    public static class AtlasService 
    {
        [MenuItem("Atlas/Test")]
        public static void Test()
        {
            Debug.Log("Atlas Kernel loaded.");
        }
    }
}