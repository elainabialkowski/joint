package joint

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/plaid/plaid-go/v3/plaid"
)

func main() {

	plaidconfig := plaid.NewConfiguration()
	plaidconfig.AddDefaultHeader("PLAID-CLIENT-ID", os.Getenv("PLAID_CLIENT_ID"))
	plaidconfig.AddDefaultHeader("PLAID-SECRET", os.Getenv("PLAID_CLIENT_SECRET"))
	plaidconfig.UseEnvironment(plaid.Development)

	plaidclient := plaid.NewAPIClient(plaidconfig)

	r := gin.Default()

	r.POST("/plaid/link/token", func(ctx *gin.Context) {
		req := plaid.NewLinkTokenCreateRequest(
			"Plaid Test App",
			"en",
			[]plaid.CountryCode{plaid.COUNTRYCODE_CA},
			*plaid.NewLinkTokenCreateRequestUser(""),
		)
		req.SetWebhook("https://webhook.sample.com")
		req.SetRedirectUri("https://domainname.com/oauth-page.html")
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

}
