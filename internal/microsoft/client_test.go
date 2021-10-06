package microsoft_test

import (
	"context"
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"

	. "github.com/homedepot/arcade/internal/microsoft"
)

var _ = Describe("Client", func() {
	var (
		server *ghttp.Server
		client *Client
		err    error
		token  string
		ctx    context.Context
	)

	BeforeEach(func() {
		ctx = context.Background()
		server = ghttp.NewServer()
		client = NewClient()
		client.WithLoginEndpoint(server.URL())
		client.WithClientID("fake-client-id")
		client.WithClientSecret("fake-client-secret")
		client.WithResource("fake-resource")
	})

	AfterEach(func() {
		server.Close()
	})

	Describe("#Token", func() {
		JustBeforeEach(func() {
			token, err = client.Token(ctx)
		})

		When("the uri is invalid", func() {
			BeforeEach(func() {
				client.WithLoginEndpoint("::haha")
			})

			It("returns an error", func() {
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal("microsoft: error making request: parse \"::haha\": missing protocol scheme"))
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

		When("the response is not 2XX", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					ghttp.RespondWith(http.StatusInternalServerError, nil),
				)
			})

			It("returns an error", func() {
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal("microsoft: error getting token: 500 Internal Server Error"))
			})
		})

		When("the server returns bad data", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					ghttp.RespondWith(http.StatusOK, ";{["),
				)
			})

			It("returns an error", func() {
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal("microsoft: error unmarshaling body: " +
					"invalid character ';' looking for beginning of value"))
			})
		})

		When("the server returns a descriptive error", func() {
			BeforeEach(func() {
				res := `{
						"error_description": "Error - requested resource not allowed",
						"error": "invalid_grant"
					}`

				server.AppendHandlers(ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodPost, "/"),
					ghttp.RespondWith(http.StatusBadRequest, res),
				))
			})

			It("returns an error", func() {
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal("microsoft: error getting token: Error - requested resource not allowed"))
			})
		})

		When("the token is cached", func() {
			BeforeEach(func() {
				res := `{
						"token_type": "Bearer",
						"expires_in": "3599",
						"ext_expires_in": "3599",
						"expires_on": "1621369811",
						"not_before": "1621365911",
						"resource": "https://graph.microsoft.com",
						"access_token": "fake.bearer.token.cached"
					}`

				server.AppendHandlers(ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodPost, "/"),
					ghttp.RespondWith(http.StatusOK, res),
				))
			})

			JustBeforeEach(func() {
				token, _ = client.Token(ctx)
			})

			It("returns the cached token", func() {
				Expect(err).To(BeNil())
				Expect(token).To(Equal("fake.bearer.token.cached"))
			})
		})

		When("the server returns a token", func() {
			BeforeEach(func() {
				res := `{
						"token_type": "Bearer",
						"expires_in": "3599",
						"ext_expires_in": "3599",
						"expires_on": "1621369811",
						"not_before": "1621365911",
						"resource": "https://graph.microsoft.com",
						"access_token": "fake.bearer.token"
					}`

				server.AppendHandlers(ghttp.CombineHandlers(
					ghttp.VerifyRequest(http.MethodPost, "/"),
					ghttp.RespondWith(http.StatusOK, res),
				))
			})

			It("succeeds", func() {
				Expect(err).To(BeNil())
				Expect(token).To(Equal("fake.bearer.token"))
			})
		})
	})

	When("there is another client", func() {
		var anotherclient *Client

		BeforeEach(func() {
			// Call to server for first client.
			res := `{
				"token_type": "Bearer",
				"expires_in": "3599",
				"ext_expires_in": "3599",
				"expires_on": "1621369811",
				"not_before": "1621365911",
				"resource": "https://graph.microsoft.com",
				"access_token": "fake.bearer.token"
			}`

			server.AppendHandlers(ghttp.CombineHandlers(
				ghttp.VerifyRequest(http.MethodPost, "/"),
				ghttp.RespondWith(http.StatusOK, res),
			))

			// Call to server for second client.
			anotherclient = NewClient()
			anotherclient.WithLoginEndpoint(server.URL())
			anotherclient.WithClientID("another-fake-client-id")
			anotherclient.WithClientSecret("antoher-fake-client-secret")
			anotherclient.WithResource("anotherfake-resource")

			// Call to server for first client.
			res = `{
				"token_type": "Bearer",
				"expires_in": "3599",
				"ext_expires_in": "3599",
				"expires_on": "1621369811",
				"not_before": "1621365911",
				"resource": "https://graph.microsoft.com",
				"access_token": "another.fake.bearer.token"
			}`

			server.AppendHandlers(ghttp.CombineHandlers(
				ghttp.VerifyRequest(http.MethodPost, "/"),
				ghttp.RespondWith(http.StatusOK, res),
			))
		})

		It("succeeds", func() {
			// Validae first client.
			Expect(err).To(BeNil())
			Expect(token).To(Equal("fake.bearer.token"))

			// Validate another client.
			token, err = anotherclient.Token(context.Background())
			Expect(err).To(BeNil())
			Expect(token).To(Equal("fake.bearer.token"))
		})
	})
})
