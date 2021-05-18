package rancher_test

import (
	"context"
	"crypto/tls"
	"net/http"

	. "github.com/homedepot/arcade/internal/rancher"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("Client", func() {
	var (
		server   *ghttp.Server
		client   *Client
		username string
		password string
		t        string
		err      error
	)

	BeforeEach(func() {
		server = ghttp.NewServer()
		username = "test-user"
		password = "test-pass"
		client = NewClient()
		client.WithURL(server.URL())
		client.WithUsername(username)
		client.WithPassword(password)
	})

	Describe("#NewToken", func() {
		AfterEach(func() {
			server.Close()
		})

		JustBeforeEach(func() {
			t, err = client.Token(context.Background())
		})

		When("the uri is invalid", func() {
			BeforeEach(func() {
				client.WithURL(":haha")
			})

			It("returns an error", func() {
				Expect(err).ToNot(BeNil())
			})
		})

		When("the server is not reachable", func() {
			BeforeEach(func() {
				server.Close()
			})

			It("returns an error", func() {
				Expect(err).ToNot(BeNil())
			})
		})

		When("the response is not 201", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					ghttp.RespondWith(404, nil),
				)
			})

			It("returns an error", func() {
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal("error getting token: 404 Not Found"))
			})
		})

		When("the response is invalid", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					ghttp.RespondWith(201, `{;'iuiuiu`),
				)
			})

			It("returns an error", func() {
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal("invalid character ';' looking for beginning of object key string"))
			})
		})

		When("the username is set", func() {
			BeforeEach(func() {
				client.WithUsername("new-user")
				json := `{"responseType": "kubeconfig","username": "new-user","password": "test-pass"}`
				server.AppendHandlers(ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/"),
					ghttp.VerifyJSON(json),
					ghttp.VerifyHeaderKV("accept", "application/json"),
					ghttp.RespondWith(http.StatusCreated, payloadKubeconfigToken),
				))
			})

			It("succeeds", func() {
				Expect(err).To(BeNil())
				Expect(t).To(Equal("kubeconfig-u-i76rfanbw5:ltqlpxqz5hh52sxfxfbxxkk6xw7pzkh7d922cww6m9x6fjskskxwl9"))
			})
		})

		When("the password is set", func() {
			BeforeEach(func() {
				client.WithPassword("new-pass")
				json := `{"responseType": "kubeconfig","username": "test-user","password": "new-pass"}`
				server.AppendHandlers(ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/"),
					ghttp.VerifyJSON(json),
					ghttp.VerifyHeaderKV("accept", "application/json"),
					ghttp.RespondWith(http.StatusCreated, payloadKubeconfigToken),
				))
			})

			It("succeeds", func() {
				Expect(err).To(BeNil())
				Expect(t).To(Equal("kubeconfig-u-i76rfanbw5:ltqlpxqz5hh52sxfxfbxxkk6xw7pzkh7d922cww6m9x6fjskskxwl9"))
			})
		})

		When("the transport is set", func() {
			BeforeEach(func() {
				t := &http.Transport{
					TLSClientConfig: &tls.Config{},
				}
				client.WithTransport(t)
				json := `{"responseType": "kubeconfig","username": "test-user","password": "test-pass"}`
				server.AppendHandlers(ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/"),
					ghttp.VerifyJSON(json),
					ghttp.VerifyHeaderKV("accept", "application/json"),
					ghttp.RespondWith(http.StatusCreated, payloadKubeconfigToken),
				))
			})

			It("succeeds", func() {
				Expect(err).To(BeNil())
				Expect(t).To(Equal("kubeconfig-u-i76rfanbw5:ltqlpxqz5hh52sxfxfbxxkk6xw7pzkh7d922cww6m9x6fjskskxwl9"))
			})
		})

		When("it succeeds", func() {
			BeforeEach(func() {
				json := `{"responseType": "kubeconfig","username": "test-user","password": "test-pass"}`
				server.AppendHandlers(ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/"),
					ghttp.VerifyJSON(json),
					ghttp.VerifyHeaderKV("accept", "application/json"),
					ghttp.RespondWith(http.StatusCreated, payloadKubeconfigToken),
				))
			})

			It("succeeds", func() {
				Expect(err).To(BeNil())
				Expect(t).To(Equal("kubeconfig-u-i76rfanbw5:ltqlpxqz5hh52sxfxfbxxkk6xw7pzkh7d922cww6m9x6fjskskxwl9"))
			})
		})
	})
})
