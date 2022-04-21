package rancher_test

import (
	"context"
	"crypto/tls"
	"net/http"
	"time"

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
		client.WithTimeout(time.Second)
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

		When("the response times out", func() {
			BeforeEach(func() {
				client.WithTimeout(0)
			})

			It("returns an error", func() {
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(HaveSuffix("context deadline exceeded"))
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

		When("the token is cached", func() {
			BeforeEach(func() {
				json := `{"responseType": "kubeconfig","username": "test-user","password": "test-pass"}`
				server.AppendHandlers(ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/"),
					ghttp.VerifyJSON(json),
					ghttp.VerifyHeaderKV("accept", "application/json"),
					ghttp.RespondWith(http.StatusCreated, payloadKubeconfigTokenCached),
				))
			})

			It("succeeds", func() {
				Expect(err).To(BeNil())
				Expect(t).To(Equal("fake.token.cached"))
				// Second call returns cached token
				t2, _ := client.Token(context.Background())
				Expect(t2).To(Equal("fake.token.cached"))
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

		When("there is another client", func() {
			var anotherclient *Client

			BeforeEach(func() {
				// Call to server for first client.
				json := `{"responseType": "kubeconfig","username": "test-user","password": "test-pass"}`
				server.AppendHandlers(ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/"),
					ghttp.VerifyJSON(json),
					ghttp.VerifyHeaderKV("accept", "application/json"),
					ghttp.RespondWith(http.StatusCreated, payloadKubeconfigToken),
				))
				// Call to server for second client.
				anotherclient = NewClient()
				anotherclient.WithURL(server.URL())
				anotherclient.WithUsername("another-test-user")
				anotherclient.WithPassword("another-test-pass")
				anotherclient.WithTimeout(time.Second)
				json = `{"responseType": "kubeconfig","username": "another-test-user","password": "another-test-pass"}`
				server.AppendHandlers(ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/"),
					ghttp.VerifyJSON(json),
					ghttp.VerifyHeaderKV("accept", "application/json"),
					ghttp.RespondWith(http.StatusCreated, payloadKubeconfigTokenAnother),
				))
			})

			It("succeeds", func() {
				// Validate first client.
				Expect(err).To(BeNil())
				Expect(t).To(Equal("kubeconfig-u-i76rfanbw5:ltqlpxqz5hh52sxfxfbxxkk6xw7pzkh7d922cww6m9x6fjskskxwl9"))

				// Validate another client.
				t, err = anotherclient.Token(context.Background())
				Expect(err).To(BeNil())
				Expect(t).To(Equal("another.token"))
			})
		})

		When("there is a cached token, shortExpiration set and it has passed", func() {
			BeforeEach(func() {
				client.WithShortExpiration(1)
				json := `{"responseType": "kubeconfig","username": "test-user","password": "test-pass"}`
				// create the "cache" in the client
				server.AppendHandlers(ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/"),
					ghttp.VerifyJSON(json),
					ghttp.VerifyHeaderKV("accept", "application/json"),
					ghttp.RespondWith(http.StatusCreated, payloadKubeconfigToken),
				))
				// call to test the short expiration
				server.AppendHandlers(ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/"),
					ghttp.VerifyJSON(json),
					ghttp.VerifyHeaderKV("accept", "application/json"),
					ghttp.RespondWith(http.StatusCreated, payloadKubeconfigTokenAnother),
				))
			})
			It("fetches a new token", func() {
				// make the initial call to create the cached token
				Expect(err).To(BeNil())
				Expect(t).To(Equal("kubeconfig-u-i76rfanbw5:ltqlpxqz5hh52sxfxfbxxkk6xw7pzkh7d922cww6m9x6fjskskxwl9"))
				time.Sleep(2 * time.Second)

				// Second call returns the newly fetched token
				t2, err2 := client.Token(context.Background())
				Expect(err2).To(BeNil())
				Expect(t2).To(Equal("another.token"))
			})
		})

		When("there is a shortExpiration set and it has not passed and there is a cached token", func() {
			BeforeEach(func() {
				client.WithShortExpiration(60)
				json := `{"responseType": "kubeconfig","username": "test-user","password": "test-pass"}`
				server.AppendHandlers(ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/"),
					ghttp.VerifyJSON(json),
					ghttp.VerifyHeaderKV("accept", "application/json"),
					ghttp.RespondWith(http.StatusCreated, payloadKubeconfigTokenCached),
				))
			})

			It("returns the cached token", func() {
				// First call with cached response
				Expect(err).To(BeNil())
				Expect(t).To(Equal("fake.token.cached"))

				time.Sleep(3 * time.Second)

				// Second call returns the same response
				t2, err2 := client.Token(context.Background())
				Expect(err2).To(BeNil())
				Expect(t2).To(Equal("fake.token.cached"))

			})
		})

		When("there is a shortExpiration set and it has not passed and there is no cached token", func() {
			BeforeEach(func() {
				client.WithShortExpiration(1)
				json := `{"responseType": "kubeconfig","username": "test-user","password": "test-pass"}`
				server.AppendHandlers(ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/"),
					ghttp.VerifyJSON(json),
					ghttp.VerifyHeaderKV("accept", "application/json"),
					ghttp.RespondWith(http.StatusCreated, payloadKubeconfigToken),
				))
			})

			It("fetches a new token", func() {
				Expect(err).To(BeNil())
				Expect(t).To(Equal("kubeconfig-u-i76rfanbw5:ltqlpxqz5hh52sxfxfbxxkk6xw7pzkh7d922cww6m9x6fjskskxwl9"))
			})
		})

	})
})
