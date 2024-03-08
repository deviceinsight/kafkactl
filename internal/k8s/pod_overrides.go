package k8s

type imagePullSecretType struct {
	Name string `json:"name"`
}

type specType struct {
	ImagePullSecrets   []imagePullSecretType `json:"imagePullSecrets"`
	ServiceAccountName *string               `json:"serviceAccountName"`
}

type PodOverrideType struct {
	APIVersion string   `json:"apiVersion"`
	Spec       specType `json:"spec"`
}

func createPodOverride(imagePullSecret string, serviceAccount string) PodOverrideType {
	var override PodOverrideType
	override.APIVersion = "v1"

	if imagePullSecret != "" {
		override.Spec.ImagePullSecrets = make([]imagePullSecretType, 1)
		override.Spec.ImagePullSecrets[0].Name = imagePullSecret
	}

	if serviceAccount != "" {
		override.Spec.ServiceAccountName = &serviceAccount
	}

	return override
}
