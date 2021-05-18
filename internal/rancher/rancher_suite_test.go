package rancher_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestRancher(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Rancher Suite")
}
