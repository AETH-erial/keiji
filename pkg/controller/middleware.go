package controller

import (

	"github.com/gin-gonic/gin"
)

func (c *Controller) IsAuthenticated(ctx *gin.Context) {
	cookie, err := ctx.Cookie(AUTH_COOKIE_NAME)
	if err != nil {
		ctx.Redirect(302, "/login")
		return
	}
	if !c.Cache.Read(cookie) {
		ctx.Redirect(302, "/login")
		return
	}
	ctx.Next()
}