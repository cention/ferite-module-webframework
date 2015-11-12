package auth

/*
 * Implements auth middleware for cention applications.
 * Expects to be running in Gin framework
 */

import (
	wf "c3/osm/webframework"
	"c3/osm/workflow"
	"c3/web/controllers"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	HTTP_UNAUTHORIZE_ACCESS = 401
	HTTP_FORBIDDEN_ACCESS   = 403
)

func Middleware() func(*gin.Context) {
	return func(ctx *gin.Context) {
		if strings.HasPrefix(ctx.Request.RequestURI, "/debug/pprof/") {
			ctx.Next()
			return
		}

		var userId int = -1
		var password string
		var wfUser *wf.User
		var currUser *workflow.User

		userIdCookie, err := ctx.Request.Cookie("cention-suiteWFUserID")
		if err != nil {
			ctx.AbortWithStatus(HTTP_FORBIDDEN_ACCESS)
			return
		}
		userId, err = strconv.Atoi(userIdCookie.Value)
		if err != nil {
			ctx.AbortWithStatus(HTTP_FORBIDDEN_ACCESS)
			return
		}

		passwordCookie, err := ctx.Request.Cookie("cention-suiteWFPassword")
		if err != nil {
			ctx.AbortWithStatus(HTTP_FORBIDDEN_ACCESS)
			return
		}
		password = passwordCookie.Value

		wfUser, err = wf.LoadUser(userId)
		if err != nil {
			ctx.AbortWithStatus(HTTP_UNAUTHORIZE_ACCESS)
			return
		}

		if wfUser != nil {
			if wfUser.Active && wfUser.Password == password {
				currUser = controllers.FetchUserObject(wfUser.Id)
			} else {
				ctx.AbortWithStatus(HTTP_UNAUTHORIZE_ACCESS)
				return
			}
		} else {
			ctx.AbortWithStatus(HTTP_UNAUTHORIZE_ACCESS)
			return
		}

		ctx.Keys = make(map[string]interface{})
		ctx.Keys["loggedInUser"] = currUser
		ctx.Keys["wfUser"] = wfUser
		ctx.Next()
	}
}
