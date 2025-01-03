package controller

import (
	"net/http"

	"git.aetherial.dev/aeth/keiji/pkg/auth"
	"git.aetherial.dev/aeth/keiji/pkg/storage"
	"github.com/gin-gonic/gin"
)

const AUTH_COOKIE_NAME = "X-Server-Auth"

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
// @Param cred body storage.Credentials true "Admin Credentials"
// @Router /login [post]
func (c *Controller) Auth(ctx *gin.Context) {

	var cred auth.Credentials

	err := ctx.ShouldBind(&cred)
	if err != nil {
		ctx.JSON(400, map[string]string{
			"Error": err.Error(),
		})
		return
	}
	cookie, err := auth.Authorize(&cred, c.Cache)
	if err != nil {
		ctx.JSON(400, map[string]string{
			"Error": err.Error(),
		})
		return
	}
	ctx.SetCookie(AUTH_COOKIE_NAME, cookie, 3600, "/", c.Domain, false, false)

	ctx.HTML(http.StatusOK, "admin", gin.H{
		"navigation": gin.H{
			"headers": c.database.GetNavBarLinks(),
			"menu":    c.database.GetDropdownElements(),
		},
		"Tables": c.database.GetAdminTables().Tables,
	})

}

/*
@Name AddAdminTableEntry
@Summary add an entry to the admin table
@Tags admin
@Router /admin/panel
*/
func (c *Controller) AddAdminTableEntry(ctx *gin.Context) {
	tables := make(map[string][]storage.TableData)
	adminPage := storage.AdminPage{Tables: tables}
	err := ctx.ShouldBind(&adminPage)
	if err != nil {
		ctx.JSON(400, map[string]string{
			"Error": err.Error(),
		})
		return
	}
	for category := range adminPage.Tables {
		for entry := range adminPage.Tables[category] {
			err := c.database.AddAdminTableEntry(adminPage.Tables[category][entry], category)
			if err != nil {
				ctx.JSON(400, map[string]string{
					"Error": err.Error(),
				})
				return
			}

		}
	}
	ctx.Data(200, "text", []byte("Categories populated."))
}

/*
@Name AddMenuItem
@Summary add an entry to the sidebar menu
@Tags admin
@Router /admin/menu
*/
func (c *Controller) AddMenuItem(ctx *gin.Context) {
	var item storage.LinkPair
	err := ctx.ShouldBind(&item)
	if err != nil {
		ctx.JSON(400, map[string]string{
			"Error": err.Error(),
		})
		return
	}
	err = c.database.AddMenuItem(item)
	if err != nil {
		ctx.JSON(400, map[string]string{
			"Error": err.Error(),
		})
		return
	}
	ctx.Data(200, "text", []byte("menu item added."))
}

/*
@Name AddNavbarItem
@Summary add an entry to the navbar
@Tags admin
@Router /admin/navbar
*/
func (c *Controller) AddNavbarItem(ctx *gin.Context) {

	var item storage.NavBarItem
	err := ctx.ShouldBind(&item)
	if err != nil {
		ctx.JSON(400, map[string]string{
			"Error": err.Error(),
		})
		return
	}
	err = c.database.AddNavbarItem(item)
	if err != nil {
		ctx.JSON(400, map[string]string{
			"Error": err.Error(),
		})
		return
	}

	err = c.database.AddAsset(item.Link, item.Png)
	if err != nil {
		ctx.JSON(400, map[string]string{
			"Error": err.Error(),
		})
		return
	}

	ctx.Data(200, "text", []byte("navbar item added."))
}

/*
@Name AddAsset
@Summary add an asset to the db
@Tags admin
@Router /admin/assets
*/
func (c *Controller) AddAsset(ctx *gin.Context) {
	var item storage.Asset
	err := ctx.ShouldBind(&item)
	if err != nil {
		ctx.JSON(400, map[string]string{
			"Error": err.Error(),
		})
		return
	}
	err = c.database.AddAsset(item.Name, item.Data)
	if err != nil {
		ctx.JSON(400, map[string]string{
			"Error": err.Error(),
		})
		return
	}

}

// @Name AdminPanel
// @Summary serve the admin panel page
// @Tags admin
// @Router /admin/panel [get]
func (c *Controller) AdminPanel(ctx *gin.Context) {

	ctx.HTML(http.StatusOK, "admin", gin.H{
		"navigation": gin.H{
			"headers": c.database.GetNavBarLinks(),
			"menu":    c.database.GetDropdownElements(),
		},
		"Tables": c.database.GetAdminTables().Tables,
	})

}
