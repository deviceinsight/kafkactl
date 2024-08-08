package k8s

type imagePullSecretType struct {
	Name string `json:"name"`
}

type metadataType struct {
	Labels      *map[string]string `json:"labels,omitempty"`
	Annotations *map[string]string `json:"annotations,omitempty"`
}

type specType struct {
	ImagePullSecrets   []imagePullSecretType `json:"imagePullSecrets,omitempty"`
	ServiceAccountName *string               `json:"serviceAccountName,omitempty"`
	NodeSelector       *map[string]string    `json:"nodeSelector,omitempty"`
	Affinity           *map[string]any       `json:"affinity,omitempty"`
	Tolerations        *[]map[string]any     `json:"tolerations,omitempty"`
}

type PodOverrideType struct {
	APIVersion string        `json:"apiVersion"`
	Metadata   *metadataType `json:"metadata,omitempty"`
	Spec       *specType     `json:"spec,omitempty"`
}

func (o *PodOverrideType) IsEmpty() bool {
	return o.Metadata == nil && o.Spec == nil
}

func (kubectl *executor) createPodOverride() PodOverrideType {
	var override PodOverrideType
	override.APIVersion = "v1"

	if kubectl.serviceAccount != "" || kubectl.imagePullSecret != "" || len(kubectl.nodeSelector) > 0 || len(kubectl.affinity) > 0 || len(kubectl.tolerations) > 0 {
		override.Spec = &specType{}

		if kubectl.serviceAccount != "" {
			override.Spec.ServiceAccountName = &kubectl.serviceAccount
		}

		if kubectl.imagePullSecret != "" {
			override.Spec.ImagePullSecrets = make([]imagePullSecretType, 1)
			override.Spec.ImagePullSecrets[0].Name = kubectl.imagePullSecret
		}

		if len(kubectl.nodeSelector) > 0 {
			override.Spec.NodeSelector = &kubectl.nodeSelector
		}

		if len(kubectl.affinity) > 0 {
			override.Spec.Affinity = &kubectl.affinity
		}

		if len(kubectl.tolerations) > 0 {
			override.Spec.Tolerations = &kubectl.tolerations
		}
	}

	if len(kubectl.labels) > 0 || len(kubectl.annotations) > 0 {
		override.Metadata = &metadataType{}

		if len(kubectl.labels) > 0 {
			override.Metadata.Labels = &kubectl.labels
		}

		if len(kubectl.annotations) > 0 {
			override.Metadata.Annotations = &kubectl.annotations
		}
	}

	return override
}
