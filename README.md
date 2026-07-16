# Atlas OS — Phase 0 MVP

**AI-native Game Development Operating System**

*Your AI Technical Director. Not a coding assistant.*

---

## 🎯 What is Atlas?

Atlas builds a **persistent understanding of your Unity project** — not just files, but gameplay systems, performance budgets, and engineering decisions.

It lives inside your project as a safe, read-only kernel.

---

## ✅ Phase 0 Features (Shipping Now)

| Feature | Status |
|---------|--------|
| Safe installation (no source changes) | ✅ |
| Unity Editor integration | ✅ |
| SQLite graph database (`*.atlas/graph.db`) | ✅ |
| CLI for install & indexing | ✅ |
| Scene import detection | ✅ |

---

## 🚀 Quick Start

### 1. Install Atlas Kernel
```bash
cd your-unity-project/
atlas init --dry-run     # Preview changes first!
atlas init               # Actually install (creates Packages/ and .atlas/)
