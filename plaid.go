package main

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/plaid/plaid-go/v3/plaid"
)

var (
	plaidconfig = &plaid.Configuration{
		DefaultHeader: map[string]string{
			"PLAID-CLIENT-ID": os.Getenv("PLAID_CLIENT_ID"),
			"PLAID-SECRET":    os.Getenv("PLAID_CLIENT_SECRET"),
		},
	}
	plaidclient = plaid.NewAPIClient(plaidconfig)
)

func createLinkToken(c *gin.Context) {
	req := plaid.NewLinkTokenCreateRequest(
		os.Getenv("PLAID_CLIENT_NAME"),
		"en",
		[]plaid.CountryCode{plaid.COUNTRYCODE_CA},
		*plaid.NewLinkTokenCreateRequestUser("user_good"),
	)
	req.SetWebhook(os.Getenv("WEBHOOK_URI"))
	req.SetRedirectUri(os.Getenv("REDIRECT_URI"))
	req.SetProducts([]plaid.Products{plaid.PRODUCTS_AUTH})

	resp, _, err := plaidclient.PlaidApi.LinkTokenCreate(c).Execute()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"err": err.Error(),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"link_token": resp.GetLinkToken(),
	})
}

func getAccessToken(c *gin.Context) {
	publicToken := c.PostForm("public_token")

	req := plaid.NewItemPublicTokenExchangeRequest(publicToken)
	resp, _, err := plaidclient.PlaidApi.
		ItemPublicTokenExchange(c).
		ItemPublicTokenExchangeRequest(
			*req,
		).Execute()

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"err": err.Error(),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token": resp.GetAccessToken(),
		"item_id":      resp.GetItemId(),
	})
}
