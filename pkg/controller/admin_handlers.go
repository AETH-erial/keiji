package controller

import (
	"net/http"

	"adeptus-mechanicus.void/git/keiji/pkg/helpers"
	"github.com/gin-gonic/gin"
)

const AUTH_COOKIE_NAME = "X-Server-Auth"

// AddDocument uploads a document to redis
// @Description AddDocument uploads a document to redis
// @Tags admin
// @Success 200
// @Param doc body helpers.DocumentUpload true "Redis Document Upload"
// @Router	/api/v1/admin/add-document [post]
func (c *Controller) AddDocument(ctx *gin.Context) {

	var upload helpers.DocumentUpload

	err := ctx.BindJSON(&upload); if err != nil {
		ctx.JSON(400, map[string]string{
			"Error": err.Error(),
		})
		return
	}
	
	doc := helpers.NewDocument(upload.Name, nil, upload.Text, upload.Category)
	err = helpers.AddDocument(doc, c.RedisConfig); if err != nil {
		ctx.JSON(400, map[string]string{
			"Error": err.Error(),
		})
		return
	}
	
}

// @Name ServeLogin
// @Summary serves the HTML login page
// @Tags admin
// @Router /login [get]
func (c *Controller) ServeLogin(ctx *gin.Context) {
	cookie, _ := ctx.Cookie(AUTH_COOKIE_NAME)
	if c.Cache.Read(cookie) {
		ctx.Redirect(302, "/home")
	}
	ctx.HTML(http.StatusOK, "login", gin.H{
		"heading": "aetherial.dev login",
	})

}

// @Name Auth
// @Summary serves recieves admin user and pass, sets a cookie
// @Tags admin
// @Param cred body helpers.Credentials true "Admin Credentials"
// @Router /login [post]
func (c *Controller) Auth(ctx *gin.Context) {

	var cred helpers.Credentials

	err := ctx.ShouldBind(&cred); if err != nil {
		ctx.JSON(400, map[string]string{
			"Error": err.Error(),
		})
		return
	}
	cookie, err := helpers.Authorize(&cred, c.Cache)
	if err != nil {
		ctx.JSON(400, map[string]string{
			"Error": err.Error(),
		})
		return
	}
	ctx.SetCookie(AUTH_COOKIE_NAME, cookie, 3600, "/", c.Domain, false, false)
	ctx.Header("HX-Redirect", "/admin/panel")

}


// @Name AdminPanel
// @Summary serve the admin panel page
// @Tags admin
// @Router /admin/panel [get]
func (c *Controller) AdminPanel(ctx *gin.Context) {

	ctx.HTML(http.StatusOK, "admin", gin.H{
		"navigation": gin.H{
			"headers": c.Headers().Elements,
			"menu": c.Menu(),
		},
		"Tables": c.AdminTables().Tables,
	})

}
