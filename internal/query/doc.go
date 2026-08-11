// Package query provides a typed, read‑only API for accessing the Unity
// knowledge graph. It is the foundation for the Statistics Engine and
// Tool Runtime layers.
//
// All methods return types from the `models` package and hide all SQL
// implementation details. This package is used by tools, but tools never
// call it directly – they go through the Tool Runtime.
package query
