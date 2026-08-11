package models

// Edge relationship constants.
const (
	// EdgeContains represents a parent–container relationship.
	// Used for: Scene → GameObject, GameObject → child GameObject.
	EdgeContains = "CONTAINS"

	// EdgeChildOf represents a strict parent–child hierarchy.
	// Used for: GameObject → child GameObject (when parent_id is set).
	EdgeChildOf = "CHILD_OF"

	// EdgeHasComponent links a GameObject to its Component.
	EdgeHasComponent = "HAS_COMPONENT"

	// EdgeUsesScript links a Component to its Script.
	EdgeUsesScript = "USES_SCRIPT"

	// EdgeHasField links a Component to its SerializedField.
	EdgeHasField = "HAS_FIELD"

	// EdgeReferences links a SerializedField to an Asset.
	EdgeReferences = "REFERENCES"

	// EdgeInstanceOf links a GameObject to its Prefab.
	EdgeInstanceOf = "INSTANCE_OF"

	// EdgeDependsOn is a generic dependency edge (future use).
	EdgeDependsOn = "DEPENDS_ON"
)

// AllEdges returns a list of all valid edge types.
func AllEdges() []string {
	return []string{
		EdgeContains,
		EdgeChildOf,
		EdgeHasComponent,
		EdgeUsesScript,
		EdgeHasField,
		EdgeReferences,
		EdgeInstanceOf,
		EdgeDependsOn,
	}
}
