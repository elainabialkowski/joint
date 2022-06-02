package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/plaid/plaid-go/v3/plaid"
)

func main() {

	plaidconfig := plaid.NewConfiguration()
	plaidconfig.AddDefaultHeader("PLAID-CLIENT-ID", os.Getenv("PLAID_CLIENT_ID"))
	plaidconfig.AddDefaultHeader("PLAID-SECRET", os.Getenv("PLAID_CLIENT_SECRET"))
	plaidconfig.UseEnvironment(plaid.Sandbox)
	plaidclient := plaid.NewAPIClient(plaidconfig)

	db, err := pgx.Connect(context.Background(), os.Getenv("POSTGRES_URI"))
	if err != nil {
		log.Fatalf("Could not connect to db: %s\n", err.Error())
	}
	defer db.Close(context.Background())

	r := gin.Default()

	r.POST("/plaid/link/token", func(ctx *gin.Context) {
		req := plaid.NewLinkTokenCreateRequest(
			"Plaid Test App",
			"en",
			[]plaid.CountryCode{plaid.COUNTRYCODE_CA},
			*plaid.NewLinkTokenCreateRequestUser("user_good"),
		)
		req.SetWebhook(os.Getenv("WEBHOOK_URI"))
		req.SetRedirectUri(os.Getenv("REDIRECT_URI"))
		req.SetProducts([]plaid.Products{plaid.PRODUCTS_AUTH})

		resp, _, err := plaidclient.PlaidApi.LinkTokenCreate(ctx).Execute()
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"err": err.Error(),
			})
		}

		ctx.JSON(http.StatusOK, gin.H{
			"link_token": resp.GetLinkToken(),
		})
	})

	r.GET("/plaid/access/token", func(ctx *gin.Context) {
		publicToken := ctx.PostForm("public_token")

		req := plaid.NewItemPublicTokenExchangeRequest(publicToken)
		resp, _, err := plaidclient.PlaidApi.
			ItemPublicTokenExchange(ctx).
			ItemPublicTokenExchangeRequest(
				*req,
			).Execute()

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"err": err.Error(),
			})
		}

		ctx.JSON(http.StatusOK, gin.H{
			"access_token": resp.GetAccessToken(),
			"item_id":      resp.GetItemId(),
		})

	})

	r.Run()

}
