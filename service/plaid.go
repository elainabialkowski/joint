package service

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/plaid/plaid-go/plaid"
)

type PlaidService struct {
	*Server
	client   *plaid.APIClient
	userRepo UserRepository
}

func (PlaidService) Initialize(server *Server) PlaidService {
	return PlaidService{
		Server: server,
		client: plaid.NewAPIClient(&plaid.Configuration{
			DefaultHeader: map[string]string{
				"PLAID-CLIENT-ID": os.Getenv("PLAID_CLIENT_ID"),
				"PLAID-SECRET":    os.Getenv("PLAID_CLIENT_SECRET"),
			},
		}),
		userRepo: UserRepository{}.Initialize(server.Db),
	}
}

func (srv *PlaidService) CreateLinkToken(c *gin.Context) {

	uuid := c.GetString("user_id")
	req := plaid.NewLinkTokenCreateRequest(
		os.Getenv("PLAID_CLIENT_NAME"),
		"en",
		[]plaid.CountryCode{plaid.COUNTRYCODE_CA},
		*plaid.NewLinkTokenCreateRequestUser(uuid),
	)
	req.SetWebhook(os.Getenv("WEBHOOK_URI"))
	req.SetRedirectUri(os.Getenv("REDIRECT_URI"))
	req.SetProducts([]plaid.Products{plaid.PRODUCTS_AUTH})

	resp, _, err := srv.client.PlaidApi.LinkTokenCreate(c).Execute()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"err": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"link_token": resp.GetLinkToken(),
	})
}

func (srv *PlaidService) GetAccessToken(c *gin.Context) {
	publicToken := c.PostForm("public_token")

	req := plaid.NewItemPublicTokenExchangeRequest(publicToken)
	resp, _, err := srv.client.PlaidApi.
		ItemPublicTokenExchange(c).
		ItemPublicTokenExchangeRequest(
			*req,
		).Execute()

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"err": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token": resp.GetAccessToken(),
		"item_id":      resp.GetItemId(),
	})
}
