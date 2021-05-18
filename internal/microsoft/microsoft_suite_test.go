package microsoft_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestMicrosoft(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Microsoft Suite")
}
