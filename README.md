# Atlas Kernel

> The Unity-side intelligence engine for Atlas OS.

Atlas Kernel continuously analyzes your Unity project and builds a structured knowledge graph representing scenes, GameObjects, components, scripts, and relationships. This knowledge is stored locally in a SQLite database and serves as the foundation for Atlas developer tools and future AI capabilities.

---

## Overview

Atlas Kernel is designed around three core principles:

- **Local-first** – Your project data never leaves your machine.
- **Non-intrusive** – No modifications to your game source code.
- **Always up-to-date** – Automatically indexes project changes.

Rather than acting as another editor tool, Atlas Kernel acts as a background intelligence service for Unity projects.

---

## Features

### Safe Installation

- Installs through `atlas init`
- No modification of project source code
- Standard Unity Package Manager package

---

### Automatic Scene Indexing

Atlas automatically indexes:

- Scenes
- GameObjects
- Components
- Scripts
- Parent/Child relationships

Scene indexing runs in the background and keeps the project graph synchronized.

---

### Knowledge Graph Generation

Atlas builds a persistent SQLite database located at:

```
ProjectRoot/
    .atlas/
        graph.db
```

The graph currently contains:

```
Scene
    └── GameObject
            └── Component Metadata
```

Future releases will expand this model to include:

```
Scene
    └── GameObject
            ├── Component
            │      ├── Serialized Fields
            │      ├── References
            │      └── Script Metadata
            ├── Prefab
            ├── Material
            └── Asset Relationships
```

---

### GUID Extraction

Atlas reads Unity `.meta` files directly to obtain stable GUIDs.

Advantages:

- Reliable
- Fast
- Independent of AssetDatabase limitations

---

### Background Processing

Indexing is performed asynchronously.

Unity remains responsive while Atlas updates the knowledge graph.

---

### Idempotent Database Writes

Atlas safely updates existing data using:

- INSERT OR REPLACE
- Relationship synchronization
- Duplicate prevention

---

### Cross Platform

Supported platforms:

- Windows
- macOS
- Linux

---

## Current Architecture

```
Unity Project
      │
      ▼
Atlas Kernel
      │
      ▼
Scene Analysis
      │
      ▼
Knowledge Graph
      │
      ▼
graph.db
```

---

## Current Graph Model

Atlas currently indexes:

- Scene
- GameObject
- Component metadata

Example:

```
Scene
 ├── Main Camera
 │      ├── Transform
 │      └── Camera
 │
 ├── Player
 │      ├── Transform
 │      ├── Rigidbody
 │      └── PlayerController
 │
 └── GameManager
        └── GameManager
```

---

## Roadmap

### Phase 1

- Expanded component model
- Script metadata
- Serialized fields
- Asset references
- Prefab relationships

### Phase 2

- Dependency graph
- Asset graph
- Material graph
- Shader graph

### Phase 3

- Performance metrics
- Lighting analysis
- Memory analysis
- Build diagnostics

---

## Philosophy

Atlas Kernel answers one question:

> **"What exists inside this Unity project?"**

It does **not** perform AI reasoning.

Its responsibility is to build a reliable, structured source of truth for the Atlas ecosystem.

---

## License

MIT License