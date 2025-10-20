package vaultk8s_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestVaultK8s(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "VaultK8s Suite")
}
