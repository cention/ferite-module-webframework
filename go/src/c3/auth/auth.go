package auth

/*
 * Implements auth middleware for cention applications.
 * Expects to be running in Gin framework
 */

import (
	wf "c3/osm/webframework"
	"c3/web/controllers"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	HTTP_UNAUTHORIZE_ACCESS = 401
	HTTP_FORBIDDEN_ACCESS   = 403
)

func GetWebframeworkUserFromRequest(r *http.Request) *wf.User {
	userIdCookie, err := r.Cookie("cention-suiteWFUserID")
	if err != nil {
		return nil
	}
	userId, err := strconv.Atoi(userIdCookie.Value)
	if err != nil {
		log.Printf("cention-suiteWFUserID: %v\n", err)
		return nil
	}

	passwordCookie, err := r.Cookie("cention-suiteWFPassword")
	if err != nil {
		return nil
	}
	password := passwordCookie.Value

	wfUser, err := wf.LoadUser(userId)
	if err != nil {
		log.Printf("webframework.LoadUser(%d): %v\n", userId, err)
		return nil
	}

	if !wfUser.Active || wfUser.Password != password {
		return nil
	}

	return wfUser
}

func Middleware() func(*gin.Context) {
	return func(ctx *gin.Context) {
		if strings.HasPrefix(ctx.Request.RequestURI, "/debug/pprof/") {
			ctx.Next()
			return
		}

		wfUser := GetWebframeworkUserFromRequest(ctx.Request)
		if wfUser == nil {
			ctx.AbortWithStatus(HTTP_UNAUTHORIZE_ACCESS)
			return
		}

		currUser := controllers.FetchUserObject(wfUser.Id)

		ctx.Keys = make(map[string]interface{})
		ctx.Keys["loggedInUser"] = currUser
		ctx.Keys["wfUser"] = wfUser
		ctx.Next()
	}
}
