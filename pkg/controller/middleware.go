package controller

import (
	"github.com/gin-gonic/gin"
)

func (c *Controller) IsAuthenticated(ctx *gin.Context) {
	cookie, err := ctx.Cookie(AUTH_COOKIE_NAME)
	if err != nil {
		ctx.Redirect(302, "/login")
		ctx.AbortWithStatus(401)
		return
	}
	if !c.Cache.Read(cookie) {
		ctx.Redirect(302, "/login")
		ctx.AbortWithStatus(401)
		return
	}
	ctx.Next()
}
