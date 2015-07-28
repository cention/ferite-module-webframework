package auth

/*
 * Implements auth middleware for cention applications.
 * Expects to be running in Gin framework
 */

import (
	wf "c3/osm/webframework"
	"c3/osm/workflow"
	"c3/syncmap"
	"c3/web/controllers"
	"log"

	"github.com/gin-gonic/gin"
)

const (
	HTTP_UNAUTHORIZE_ACCESS = 401
	HTTP_FORBIDDEN_ACCESS   = 403
)

var ssid2idCache = syncmap.New()

var getFromCache = func(ssid string) int {
	if id, exist := ssid2idCache.Get(ssid); exist {
		return id.(int)
	}
	return -1
}

var saveToCache = func(ssid string, id int) {
	ssid2idCache.Put(ssid, id)
}

var userIdFromHash = func(ssid string) int {
	wfUser, err := wf.QueryUser_byHashLogin(ssid)
	if err != nil {
		log.Println(`auth.userIdFromHash():`, err)
		return -1
	}
	if wfUser == nil {
		// No user associated with ssid
		return -1
	}
	if wfUser.Active {
		saveToCache(ssid, wfUser.Id)
		return wfUser.Id
	}
	return -1
}

var fetchUser = func(ssid string) *workflow.User {
	var id int = -1

	id = getFromCache(ssid)

	if id == -1 {
		id = userIdFromHash(ssid)
	}

	if id != -1 {
		return controllers.FetchUserObject(id)
	}
	return nil
}

func User(ctx *gin.Context) *workflow.User {
	return ctx.Keys["loggedInUser"].(*workflow.User)
}

func Middleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var ssid string = ""

		cookie, err := ctx.Request.Cookie("cention-suiteSSID")
		if err == nil {
			ssid = cookie.Value
		}

		currUser := fetchUser(ssid)

		if currUser == nil {
			ctx.AbortWithStatus(HTTP_UNAUTHORIZE_ACCESS)
		}

		ctx.Keys = make(map[string]interface{})
		ctx.Keys["loggedInUser"] = currUser
		ctx.Next()
	}
}
