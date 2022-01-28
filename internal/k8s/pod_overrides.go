package k8s

type imagePullSecretType struct {
	Name string `json:"name"`
}

type specType struct {
	ImagePullSecrets []imagePullSecretType `json:"imagePullSecrets"`
}

type PodOverrideType struct {
	APIVersion string   `json:"apiVersion"`
	Spec       specType `json:"spec"`
}

func createPodOverrideForImagePullSecret(imagePullSecret string) PodOverrideType {
	var override PodOverrideType
	override.APIVersion = "v1"
	override.Spec.ImagePullSecrets = make([]imagePullSecretType, 1)
	override.Spec.ImagePullSecrets[0].Name = imagePullSecret
	return override
}
