package http_test

import (
	"io"
	"log"
	"os"

	arcadehttp "github.com/homedepot/arcade/internal/http"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Controller", func() {
	var (
		err error
		dir string
	)

	BeforeEach(func() {
		dir = "test"
		log.SetOutput(io.Discard)
	})

	Describe("#NewController", func() {
		JustBeforeEach(func() {
			_, err = arcadehttp.NewController(dir)
		})

		When("the directory does not exist", func() {
			BeforeEach(func() {
				dir = "i-dont-exist"
			})

			It("returns an error", func() {
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal("open i-dont-exist: no such file or directory"))
			})
		})

		When("no files exist", func() {
			BeforeEach(func() {
				dir = "empty-dir"
				_ = os.Mkdir(dir, 0666)
			})

			AfterEach(func() {
				os.Remove(dir)
			})

			It("returns an error", func() {
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal("no token providers found in directory: empty-dir"))
			})
		})

		When("a file exists with bad json", func() {
			var tmpFile *os.File

			BeforeEach(func() {
				tmpFile, err = os.CreateTemp("test", "cred*.json")
			})

			AfterEach(func() {
				os.Remove(tmpFile.Name())
			})

			It("returns an error", func() {
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal("unexpected end of JSON input"))
			})
		})

		When("a file exists without specifying a name", func() {
			var tmpFile *os.File

			BeforeEach(func() {
				tmpFile, err = os.CreateTemp("test", "provider*.json")
				_, err = tmpFile.WriteString("{}")
				Expect(err).To(BeNil())
			})

			AfterEach(func() {
				os.Remove(tmpFile.Name())
			})

			It("returns an error", func() {
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(HavePrefix("no \"name\" found in token provider config file test/provider"))
			})
		})

		When("a duplicate credential exists", func() {
			var tmpFile *os.File

			BeforeEach(func() {
				tmpFile, err = os.CreateTemp("test", "provider*.json")
				_, err = tmpFile.WriteString(`{
					"name": "google-test"
				}`)
				Expect(err).To(BeNil())
			})

			AfterEach(func() {
				os.Remove(tmpFile.Name())
			})

			It("returns an error", func() {
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(HavePrefix("duplicate token provider listed: google-test"))
			})
		})

		When("a microsoft token provider does not set the clientId", func() {
			var tmpFile *os.File

			BeforeEach(func() {
				tmpFile, err = os.CreateTemp("test", "provider*.json")
				_, err = tmpFile.WriteString(`{
          "type": "microsoft",
					"name": "test",
					"clientSecret": "clientSecret",
					"resource": "resource",
					"loginEndpoint": "loginEndpoint"
				}`)
				Expect(err).To(BeNil())
			})

			AfterEach(func() {
				os.Remove(tmpFile.Name())
			})

			It("returns an error", func() {
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(HavePrefix(`microsoft token provider file test missing required "clientId" attribute`))
			})
		})

		When("a microsoft token provider does not set the clientSecret", func() {
			var tmpFile *os.File

			BeforeEach(func() {
				tmpFile, err = os.CreateTemp("test", "provider*.json")
				_, err = tmpFile.WriteString(`{
          "type": "microsoft",
					"name": "test",
					"clientId": "clientId",
					"resource": "resource",
					"loginEndpoint": "loginEndpoint"
				}`)
				Expect(err).To(BeNil())
			})

			AfterEach(func() {
				os.Remove(tmpFile.Name())
			})

			It("returns an error", func() {
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(HavePrefix(`microsoft token provider file test missing required "clientSecret" attribute`))
			})
		})

		When("a microsoft token provider does not set the resource", func() {
			var tmpFile *os.File

			BeforeEach(func() {
				tmpFile, err = os.CreateTemp("test", "provider*.json")
				_, err = tmpFile.WriteString(`{
          "type": "microsoft",
					"name": "test",
					"clientId": "clientId",
					"clientSecret": "clientSecret",
					"loginEndpoint": "loginEndpoint"
				}`)
				Expect(err).To(BeNil())
			})

			AfterEach(func() {
				os.Remove(tmpFile.Name())
			})

			It("returns an error", func() {
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(HavePrefix(`microsoft token provider file test missing required "resource" attribute`))
			})
		})

		When("a microsoft token provider does not set the loginEndpoint", func() {
			var tmpFile *os.File

			BeforeEach(func() {
				tmpFile, err = os.CreateTemp("test", "provider*.json")
				_, err = tmpFile.WriteString(`{
          "type": "microsoft",
					"name": "test",
					"clientId": "clientId",
					"clientSecret": "clientSecret",
					"resource": "resource"
				}`)
				Expect(err).To(BeNil())
			})

			AfterEach(func() {
				os.Remove(tmpFile.Name())
			})

			It("returns an error", func() {
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(HavePrefix(`microsoft token provider file test missing required "loginEndpoint" attribute`))
			})
		})

		When("a rancher token provider does not set the username", func() {
			var tmpFile *os.File

			BeforeEach(func() {
				tmpFile, err = os.CreateTemp("test", "provider*.json")
				_, err = tmpFile.WriteString(`{
          "type": "rancher",
					"name": "test",
					"password": "password",
					"url": "url"
				}`)
				Expect(err).To(BeNil())
			})

			AfterEach(func() {
				os.Remove(tmpFile.Name())
			})

			It("returns an error", func() {
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(HavePrefix(`rancher token provider file test missing required "username" attribute`))
			})
		})

		When("a rancher token provider does not set the password", func() {
			var tmpFile *os.File

			BeforeEach(func() {
				tmpFile, err = os.CreateTemp("test", "provider*.json")
				_, err = tmpFile.WriteString(`{
          "type": "rancher",
					"name": "test",
					"username": "username",
					"url": "url"
				}`)
				Expect(err).To(BeNil())
			})

			AfterEach(func() {
				os.Remove(tmpFile.Name())
			})

			It("returns an error", func() {
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(HavePrefix(`rancher token provider file test missing required "password" attribute`))
			})
		})

		When("a rancher token provider does not set the url", func() {
			var tmpFile *os.File

			BeforeEach(func() {
				tmpFile, err = os.CreateTemp("test", "provider*.json")
				_, err = tmpFile.WriteString(`{
          "type": "rancher",
					"name": "test",
					"username": "username",
					"password": "password"
				}`)
				Expect(err).To(BeNil())
			})

			AfterEach(func() {
				os.Remove(tmpFile.Name())
			})

			It("returns an error", func() {
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(HavePrefix(`rancher token provider file test missing required "url" attribute`))
			})
		})

		When("it succeeds", func() {
			It("succeeds", func() {
				Expect(err).To(BeNil())
			})
		})
	})
})
