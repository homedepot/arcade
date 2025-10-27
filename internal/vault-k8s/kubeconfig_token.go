package vaultk8s

// KubeconfigToken represents the structure of the kubeconfig token response from Vault
// for the vault-k8s provider.
type KubeconfigToken struct {
	Data struct {
		Users []struct {
			Name string `json:"name"`
			User struct {
				Token string `json:"token"`
			} `json:"user"`
		} `json:"users"`
	} `json:"data"`
}
