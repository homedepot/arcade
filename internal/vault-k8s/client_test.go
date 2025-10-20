package vaultk8s_test

import (
	"context"
	"net/http"
	"testing"

	. "github.com/homedepot/arcade/internal/vault-k8s"
	. "github.com/onsi/ginkgo"

	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("Client", func() {
	var (
		server   *ghttp.Server
		client   *Client
		password string
		token    string
		err      error
		ctx      context.Context
	)

	BeforeEach(func() {
		server = ghttp.NewServer()
		password = "test-vault-token"
		client = NewClient()
		client.WithURL(server.URL())
		client.WithPassword(password)
		ctx = context.Background()
	})

	AfterEach(func() {
		server.Close()
	})

	Describe("#Token", func() {
		JustBeforeEach(func() {
			t := testing.T{}
			t.Setenv("VAULT_K8S_PATH_PATTERN", "secret/data/[CLUSTER]/vault-k8s-user")
			token, err = client.Token(ctx)
		})

		When("it succeeds", func() {
			BeforeEach(func() {
				ctx = context.WithValue(ctx, "provider", "vault-k8s-my-cluster")
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/v1/secret/data/my-cluster/vault-k8s-user"),
						ghttp.VerifyHeader(http.Header{
							"X-Vault-Token": []string{"test-vault-token"},
						}),
						ghttp.RespondWith(http.StatusOK, `{
						  "data": {
							"data": {
							  "users": [
								{
								  "name": "vault-k8s-user",
								  "user": {
									"token": "test-kubeconfig-token"
								  }
								}
							  ]
							}
						  }
						}`),
					),
				)
			})

			It("returns the kubeconfig token", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(token).To(Equal("test-kubeconfig-token"))
				Expect(server.ReceivedRequests()).To(HaveLen(1))
			})
		})

		When("the context is missing the provider", func() {
			BeforeEach(func() {
				// Don't set the provider in the context
			})

			It("returns an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("cluster name not found in context"))
				Expect(server.ReceivedRequests()).To(HaveLen(0))
			})
		})

		When("the provider in the context has an invalid format", func() {
			BeforeEach(func() {
				ctx = context.WithValue(ctx, "provider", "vault-k8s")
			})

			It("returns an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("invalid cluster name format"))
				Expect(server.ReceivedRequests()).To(HaveLen(0))
			})
		})

		When("the secret is not found in vault", func() {
			BeforeEach(func() {
				ctx = context.WithValue(ctx, "provider", "vault-k8s-my-cluster")
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/v1/secret/data/my-cluster/vault-k8s-user"),
						ghttp.RespondWith(http.StatusNotFound, ""),
					),
				)
			})

			It("returns an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("secret not found at"))
				Expect(server.ReceivedRequests()).To(HaveLen(1))
			})
		})

		When("the kubeconfig is not valid json", func() {
			BeforeEach(func() {
				ctx = context.WithValue(ctx, "provider", "vault-k8s-my-cluster")
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/v1/secret/data/my-cluster/vault-k8s-user"),
						ghttp.RespondWith(http.StatusOK, `{"data":{"data":{"kubeconfig":"not-json"}}}`),
					),
				)
			})

			It("returns an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("no users found in kubeconfig token"))
				Expect(server.ReceivedRequests()).To(HaveLen(1))
			})
		})
	})
})
