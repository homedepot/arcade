package http_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/gin-gonic/gin"
	arcadehttp "github.com/homedepot/arcade/internal/http"
	"github.com/homedepot/arcade/pkg/provider/providerfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type Tokens struct {
	Token string `json:"token"`
	Error string `json:"error"`
}

var (
	err                 error
	svr                 *httptest.Server
	uri                 string
	req                 *http.Request
	body                *bytes.Buffer
	res                 *http.Response
	tokens              Tokens
	fakeGoogleClient    *providerfakes.FakeClient
	fakeMicrosoftClient *providerfakes.FakeClient
	fakeRancherClient   *providerfakes.FakeClient
	controller          *arcadehttp.Controller
)

var _ = Describe("Token", func() {
	BeforeEach(func() {
		fakeGoogleClient = &providerfakes.FakeClient{}
		fakeGoogleClient.TokenReturns("valid-google-token", nil)
		fakeRancherClient = &providerfakes.FakeClient{}
		fakeRancherClient.TokenReturns("valid-rancher-token", nil)
		fakeMicrosoftClient = &providerfakes.FakeClient{}
		fakeMicrosoftClient.TokenReturns("valid-microsoft-token", nil)
		// Disable debug logging.
		gin.SetMode(gin.ReleaseMode)
		// Setup the controller.
		controller = &arcadehttp.Controller{
			Tokenizers: map[string]arcadehttp.Tokenizer{
				"google":    fakeGoogleClient,
				"microsoft": fakeMicrosoftClient,
				"rancher":   fakeRancherClient,
			},
		}
		// Setup the server.
		r := gin.New()
		r.Use(gin.Recovery())
		r.GET("/tokens", controller.GetToken)

		svr = httptest.NewServer(r)
		body = &bytes.Buffer{}
	})

	AfterEach(func() {
		svr.Close()
		err = res.Body.Close()
		Expect(err).ToNot(HaveOccurred())
	})

	JustBeforeEach(func() {
		req, _ = http.NewRequest(http.MethodGet, uri, nil)
		res, err = http.DefaultClient.Do(req)
	})

	Describe("#GetToken", func() {
		When("provider is not supported", func() {
			BeforeEach(func() {
				uri = svr.URL + "/tokens?provider=fake"
			})

			It("succeeds", func() {
				Expect(res.StatusCode).To(Equal(http.StatusBadRequest))
				b, _ := io.ReadAll(res.Body)
				_ = json.Unmarshal(b, &tokens)
				Expect(tokens.Error).To(Equal("Unsupported token provider: fake"))
			})
		})
	})

	Describe("#GetGoogleToken", func() {
		BeforeEach(func() {
			uri = svr.URL + "/tokens?provider=google"
		})

		When("getting a new token from google fails", func() {
			BeforeEach(func() {
				fakeGoogleClient.TokenReturns("", errors.New("error getting token from google"))
			})

			It("returns an internal server error", func() {
				Expect(res.StatusCode).To(Equal(http.StatusInternalServerError))
				b, _ := io.ReadAll(res.Body)
				_ = json.Unmarshal(b, &tokens)
				Expect(tokens.Error).To(Equal("error getting token from google"))
			})
		})

		When("no provider is specified", func() {
			BeforeEach(func() {
				uri = svr.URL + "/tokens"
			})

			It("defaults to google", func() {
				Expect(res.StatusCode).To(Equal(http.StatusOK))
				b, _ := io.ReadAll(res.Body)
				_ = json.Unmarshal(b, &tokens)
				Expect(tokens.Token).To(Equal("valid-google-token"))
			})
		})

		When("it succeeds", func() {
			It("succeeds", func() {
				Expect(res.StatusCode).To(Equal(http.StatusOK))
				b, _ := io.ReadAll(res.Body)
				_ = json.Unmarshal(b, &tokens)
				Expect(tokens.Token).To(Equal("valid-google-token"))
			})
		})
	})

	Describe("#GetMicrosoftToken", func() {
		BeforeEach(func() {
			uri = svr.URL + "/tokens?provider=microsoft"
		})

		When("microsoft is not a configured provider", func() {
			BeforeEach(func() {
				delete(controller.Tokenizers, "microsoft")
			})

			It("returns a bad request error", func() {
				Expect(res.StatusCode).To(Equal(http.StatusBadRequest))
				b, _ := io.ReadAll(res.Body)
				_ = json.Unmarshal(b, &tokens)
				Expect(tokens.Error).To(Equal("Unsupported token provider: microsoft"))
			})
		})

		When("getting a new token from microsoft fails", func() {
			BeforeEach(func() {
				fakeMicrosoftClient.TokenReturns("", errors.New("error getting token from microsoft"))
			})

			It("returns an internal server error", func() {
				Expect(res.StatusCode).To(Equal(http.StatusInternalServerError))
				b, _ := io.ReadAll(res.Body)
				_ = json.Unmarshal(b, &tokens)
				Expect(tokens.Error).To(Equal("error getting token from microsoft"))
			})
		})

		When("it succeeds", func() {
			It("succeeds", func() {
				Expect(res.StatusCode).To(Equal(http.StatusOK))
				b, _ := io.ReadAll(res.Body)
				_ = json.Unmarshal(b, &tokens)
				Expect(tokens.Token).To(Equal("valid-microsoft-token"))
			})
		})
	})

	Describe("#GetRancherToken", func() {
		BeforeEach(func() {
			uri = svr.URL + "/tokens?provider=rancher"
		})

		When("rancher is not a configured provider", func() {
			BeforeEach(func() {
				delete(controller.Tokenizers, "rancher")
			})

			It("returns a bad request error", func() {
				Expect(res.StatusCode).To(Equal(http.StatusBadRequest))
				b, _ := io.ReadAll(res.Body)
				_ = json.Unmarshal(b, &tokens)
				Expect(tokens.Error).To(Equal("Unsupported token provider: rancher"))
			})
		})

		When("getting a new token from rancher fails", func() {
			BeforeEach(func() {
				fakeRancherClient.TokenReturns("", errors.New("error getting token from rancher"))
			})

			It("returns an internal server error", func() {
				Expect(res.StatusCode).To(Equal(http.StatusInternalServerError))
				b, _ := io.ReadAll(res.Body)
				_ = json.Unmarshal(b, &tokens)
				Expect(tokens.Error).To(Equal("error getting token from rancher"))
			})
		})

		When("it succeeds", func() {
			It("succeeds", func() {
				Expect(res.StatusCode).To(Equal(http.StatusOK))
				b, _ := io.ReadAll(res.Body)
				_ = json.Unmarshal(b, &tokens)
				Expect(tokens.Token).To(Equal("valid-rancher-token"))
			})
		})
	})
})
