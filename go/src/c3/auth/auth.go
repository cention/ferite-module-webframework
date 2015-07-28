package auth

/*
 * Implements auth middleware for cention applications.
 * Expects to be running in Gin framework
 */

import (
	wf "c3/osm/webframework"
	"c3/osm/workflow"
	"c3/web/controllers"
	// "fmt"
	"log"
	"net"

	"github.com/cention-mujibur-rahman/gobcache"
	"github.com/gin-gonic/gin"
)

const (
	HTTP_UNAUTHORIZE_ACCESS = 401
	HTTP_FORBIDDEN_ACCESS   = 403
)

var memcacheIsRunning bool

var checkingMemcache = func() bool {
	conn, err := net.Dial("tcp", "localhost:11211")
	if err != nil {
		log.Println("Memcache is not running! ", err)
		return false
	}
	defer conn.Close()
	return true
}

var getFromMemcache = func(ssid string) (id int) {
	id = -1
	if err := gobcache.GetFromMemcache("Session_"+ssid, &id); err != nil {
		log.Println("[Set(Cookie)] Memcache key `Session` is empty!", err)
	}
	return
}

var saveInMemcache = func(ssid string, id int) {
	if err := gobcache.SaveInMemcache("Session_"+ssid, id); err != nil {
		log.Println("[`SaveInMemcache`] Error on saving:", err)
	}
}

var userIdFromHash = func(ssid string) int {
	wfUser, err := wf.QueryUser_byHashLogin(ssid)
	if err == nil && wfUser != nil {
		if wfUser.Active {
			if memcacheIsRunning {
				saveInMemcache("Session_"+ssid, wfUser.Id)
			}
			return wfUser.Id
		}
	}
	return -1
}

var fetchUser = func(ssid string) *workflow.User {
	var id int = -1

	if memcacheIsRunning {
		id = getFromMemcache(ssid)
	}

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
	memcacheIsRunning = checkingMemcache()

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
