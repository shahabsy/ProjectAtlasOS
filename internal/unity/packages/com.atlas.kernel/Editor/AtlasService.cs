using UnityEditor;
using UnityEngine;
using System.IO;

namespace Atlas.Kernel.Editor
{
    public class AtlasService : EditorWindow
    {
        [MenuItem("Atlas/Test", priority = 10)]
        public static void OpenTestWindow()
        {
            var window = GetWindow<AtlasService>("Atlas Test");
            window.minSize = new Vector2(400, 300);
        }

        void OnGUI()
        {
            EditorGUILayout.LabelField("Atlas Kernel Loaded!", EditorStyles.boldLabel);
            EditorGUILayout.Space();

            EditorGUILayout.BeginHorizontal();

            if(GUILayout.Button("Open in Explorer"))
            {
                string atlasPath = Path.Combine(Application.dataPath, "..", ".atlas");
                if (Directory.Exists(atlasPath))
                {
                    EditorUtility.RevealInFinder(atlasPath);
                }
                else
                {
                    Debug.LogWarning("[Atlas] .atlas folder not found. Run `atlas init` first.");
                }
            }

            GUILayout.FlexibleSpace();
            
            if(GUILayout.Button("Close"))
            {
                this.Close();
            }
            EditorGUILayout.EndHorizontal();

            EditorGUILayout.Space();

            EditorGUILayout.HelpBox(
                "Your first atlas project is ready. \nNext: Add scene indexing on AssetPostProcess.", MessageType.Info);
        }
        // Ensure Atlas folder exists on startup
        [InitializeOnLoadMethod]
        static void OnEditorStartup()
        {
            string atlasPath = Path.Combine(Application.dataPath, "..", ".atlas");
            if(!Directory.Exists(atlasPath))
            {
                Directory.CreateDirectory(atlasPath);
                Debug.Log("[Atlas] Project initialized - .atlas folder created.");
            }
        }
    }
}
