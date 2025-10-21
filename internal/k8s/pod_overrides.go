package k8s

type imagePullSecretType struct {
	Name string `json:"name"`
}

// JSONPatchOperation represents a single JSON Patch operation (RFC 6902)
type JSONPatchOperation struct {
	Op    string `json:"op"`
	Path  string `json:"path"`
	Value any    `json:"value,omitempty"`
}

// JSONPatchType is an array of JSON Patch operations
type JSONPatchType []JSONPatchOperation

func (kubectl *executor) createPodOverrides() JSONPatchType {
	var patches JSONPatchType

	// Add imagePullSecrets if specified
	if kubectl.imagePullSecret != "" {
		patches = append(patches, JSONPatchOperation{
			Op:   "add",
			Path: "/spec/imagePullSecrets",
			Value: []imagePullSecretType{
				{Name: kubectl.imagePullSecret},
			},
		})
	}

	// Add serviceAccountName if specified
	if kubectl.serviceAccount != "" {
		patches = append(patches, JSONPatchOperation{
			Op:    "add",
			Path:  "/spec/serviceAccountName",
			Value: kubectl.serviceAccount,
		})
	}

	// Add nodeSelector if specified
	if len(kubectl.nodeSelector) > 0 {
		patches = append(patches, JSONPatchOperation{
			Op:    "add",
			Path:  "/spec/nodeSelector",
			Value: kubectl.nodeSelector,
		})
	}

	// Add affinity if specified
	if len(kubectl.affinity) > 0 {
		patches = append(patches, JSONPatchOperation{
			Op:    "add",
			Path:  "/spec/affinity",
			Value: kubectl.affinity,
		})
	}

	// Add tolerations if specified
	if len(kubectl.tolerations) > 0 {
		patches = append(patches, JSONPatchOperation{
			Op:    "add",
			Path:  "/spec/tolerations",
			Value: kubectl.tolerations,
		})
	}

	// Add labels if specified
	if len(kubectl.labels) > 0 {
		patches = append(patches, JSONPatchOperation{
			Op:    "add",
			Path:  "/metadata/labels",
			Value: kubectl.labels,
		})
	}

	// Add annotations if specified
	if len(kubectl.annotations) > 0 {
		patches = append(patches, JSONPatchOperation{
			Op:    "add",
			Path:  "/metadata/annotations",
			Value: kubectl.annotations,
		})
	}

	return patches
}
