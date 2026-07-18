using UnityEditor;
using UnityEngine;
using System.IO;

/// <summary>
/// Editor window to set the path to the atlas.exe CLI tool.
/// Stores the path in EditorPrefs under the key "Atlas.ExePath".
/// </summary>
namespace Atlas.Kernel.Editor
{
    public class AtlasSettingsWindow : EditorWindow
    {
        [MenuItem("Atlas/Settings", priority = 20)]
        public static void ShowWindow()
        {
            var window = GetWindow<AtlasSettingsWindow>("Atlas Settings");
            window.minSize = new Vector2(500, 150);
            window.Show();
        }

        private string exePath;
        private bool pathValid;

        void OnEnable()
        {
            // Load saved path from EditorPrefs
            exePath = EditorPrefs.GetString("Atlas.ExePath", @"E:\ProjectAtlasOS\atlas.exe");
            ValidatePath();
        }

        void OnGUI()
        {
            EditorGUILayout.LabelField("Atlas CLI Configuration", EditorStyles.boldLabel);
            EditorGUILayout.Space();

            EditorGUILayout.HelpBox(
                "Set the full path to the atlas.exe executable.\n" +
                "This path is used by the Unity Editor to call the Atlas CLI for indexing.",
                MessageType.Info
            );

            EditorGUILayout.Space();

            // Path input field
            EditorGUILayout.BeginHorizontal();
            EditorGUILayout.PrefixLabel("atlas.exe Path");
            string newPath = EditorGUILayout.TextField(exePath);
            if (GUILayout.Button("Browse...", GUILayout.Width(80)))
            {
                string selectedPath = EditorUtility.OpenFilePanel("Select atlas.exe", Path.GetDirectoryName(exePath), "exe");
                if (!string.IsNullOrEmpty(selectedPath))
                {
                    newPath = selectedPath;
                }
            }
            EditorGUILayout.EndHorizontal();

            // Path validation display
            if (exePath != newPath)
            {
                exePath = newPath;
                ValidatePath();
            }

            EditorGUILayout.Space();

            EditorGUILayout.BeginHorizontal();
            GUI.enabled = pathValid;
            if (GUILayout.Button("Save Path"))
            {
                EditorPrefs.SetString("Atlas.ExePath", exePath);
                Debug.Log($"[Atlas] Saved atlas.exe path: {exePath}");
            }
            GUI.enabled = true;

            if (GUILayout.Button("Test Path", GUILayout.Width(100)))
            {
                if (File.Exists(exePath))
                {
                    Debug.Log($"[Atlas] ✔ atlas.exe found at: {exePath}");
                }
                else
                {
                    Debug.LogError($"[Atlas] ✘ atlas.exe NOT found at: {exePath}");
                    ValidatePath();
                }
            }
            EditorGUILayout.EndHorizontal();

            EditorGUILayout.Space();

            // Status message
            if (pathValid)
            {
                EditorGUILayout.HelpBox("✅ Path is valid and atlas.exe exists.", MessageType.Info);
            }
            else
            {
                EditorGUILayout.HelpBox("❌ Path is invalid or atlas.exe not found. Please select a valid executable.", MessageType.Error);
            }

            EditorGUILayout.Space();

            if (GUILayout.Button("Reset to Default", GUILayout.Width(150)))
            {
                exePath = @"E:\ProjectAtlasOS\atlas.exe";
                ValidatePath();
            }

            // Option to open folder containing atlas.exe
            if (pathValid)
            {
                EditorGUILayout.Space();
                if (GUILayout.Button("Open Containing Folder", GUILayout.Width(180)))
                {
                    EditorUtility.RevealInFinder(Path.GetDirectoryName(exePath));
                }
            }
        }

        private void ValidatePath()
        {
            pathValid = File.Exists(exePath);
        }
    }
}