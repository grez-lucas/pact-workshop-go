package client

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"testing"

	"github.com/pact-foundation/pact-go/v2/consumer"
	"github.com/pact-foundation/pact-go/v2/log"
	"github.com/pact-foundation/pact-go/v2/matchers"
	"github.com/pact-foundation/pact-workshop-go/model"
	"github.com/stretchr/testify/assert"
)

var (
	Like         = matchers.Like
	EachLike     = matchers.EachLike
	Term         = matchers.Term
	Regex        = matchers.Regex
	HexValue     = matchers.HexValue
	Identifier   = matchers.Identifier
	IPAddress    = matchers.IPAddress
	IPv6Address  = matchers.IPv6Address
	Timestamp    = matchers.Timestamp
	Date         = matchers.Date
	Time         = matchers.Time
	UUID         = matchers.UUID
	ArrayMinLike = matchers.ArrayMinLike
)

type (
	S   = matchers.S
	Map = matchers.MapMatcher
)

var (
	u      *url.URL
	client *Client
)

func TestClientPact_GetUser(t *testing.T) {
	log.SetLogLevel("INFO")
	mockProvider, err := consumer.NewV2Pact(consumer.MockHTTPProviderConfig{
		Consumer: os.Getenv("CONSUMER_NAME"),
		Provider: os.Getenv("PROVIDER_NAME"),
		LogDir:   os.Getenv("LOG_DIR"),
		PactDir:  os.Getenv("PACT_DIR"),
	})
	assert.NoError(t, err)

	t.Run("the user exists", func(t *testing.T) {
		id := 10

		err = mockProvider.
			AddInteraction().
			Given("User sally exists").
			UponReceiving("A request to login with user 'sally'").
			WithRequestPathMatcher("GET", Regex("/user/"+strconv.Itoa(id), "/user/[0-9]+"), func(vrb *consumer.V2RequestBuilder) {
				vrb.Header("Authorization", Like("Bearer 2019-01-01"))
			}).
			WillRespondWith(200, func(b *consumer.V2ResponseBuilder) {
				b.BodyMatch(model.User{}).
					Header("Content-Type", Term("application/json", `application\/json`)).
					Header("X-Api-Correlation-Id", Like("100"))
			}).
			ExecuteTest(t, func(config consumer.MockServerConfig) error {
				// Act: test our API client behaves correctly

				// Get the Pact mock server URL
				u, _ = url.Parse("http://" + config.Host + ":" + strconv.Itoa(config.Port))

				// Initialise the API client and point it at the Pact mock server
				client = &Client{
					BaseURL: u,
				}

				// // Execute the API client
				user, err := client.WithToken("2019-01-01").GetUser(id)

				// // Assert basic fact
				if user.ID != id {
					return fmt.Errorf("wanted user with ID %d but got %d", id, user.ID)
				}

				return err
			})

		assert.NoError(t, err)
	})

	t.Run(
		"the user does not exist",
		func(t *testing.T) {
			id := 10

			err = mockProvider.AddInteraction().
				Given("User sally does not exist").
				UponReceiving("A request to log in with user 'sally'").
				WithRequestPathMatcher("GET", Regex("/user/"+strconv.Itoa(id), "/user/[0-9]+"), func(vrb *consumer.V2RequestBuilder) {
					vrb.Header("Authorization", Like("Bearer 2019-01-01"))
				}).
				WillRespondWith(404, func(vrb *consumer.V2ResponseBuilder) {
					vrb.Header("Content-Type", Term("application/json", `application\/json`)).
						Header("X-Api-Correlation-Id", Like("100"))
				}).
				ExecuteTest(t, func(config consumer.MockServerConfig) error {
					// Act: test our API client behaves correctly

					// Get the pact mock server URL
					u, _ = url.Parse("http://" + config.Host + ":" + strconv.Itoa(config.Port))

					// Initialise the API client and point it at the Pact mock server
					client = &Client{
						BaseURL: u,
					}

					_, err := client.WithToken("2019-01-01").GetUser(id)
					assert.Equal(t, ErrNotFound, err)
					return nil
				})

			assert.NoError(t, err)
		},
	)

	t.Run(
		"the user is not authorized",
		func(t *testing.T) {
			id := 10

			err = mockProvider.AddInteraction().
				Given("User is not authorized").
				UponReceiving("A request to log in with user 'sally' that isn't authorized").
				WithRequestPathMatcher("GET", Regex("/user/"+strconv.Itoa(id), "/user/[0-9]+")).
				WillRespondWith(403,
					func(vrb *consumer.V2ResponseBuilder) {
						vrb.Header("Content-Type", Term("application/json", `application\/json`)).
							Header("X-Api-Correlation-Id", Like("100"))
					}).
				ExecuteTest(t, func(config consumer.MockServerConfig) error {
					// Act: test our API client behaves correctly

					// Get the pact mock server URL
					u, _ = url.Parse("http://" + config.Host + ":" + strconv.Itoa(config.Port))

					// Initialise the API client and point it at the Pact mock server
					client = &Client{
						BaseURL: u,
					}

					_, err := client.WithToken("").GetUser(id)
					assert.Equal(t, ErrUnauthorized, err)
					return nil
				})

			assert.NoError(t, err)
		},
	)
}
