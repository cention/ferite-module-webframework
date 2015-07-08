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
	"github.com/cention-mujibur-rahman/gobcache"
	"github.com/gin-gonic/gin"
	"net"
	"log"
)

const (
	HTTP_UNAUTHORIZE_ACCESS = 401
	HTTP_FORBIDDEN_ACCESS   = 403
)

func checkingMemcache() bool {
	conn, err := net.Dial("tcp", "localhost:11211")
	if err != nil {
		log.Println("Memcache is not running! ", err)
		return false
	}
	defer conn.Close()
	return true
}

func Middleware () func (*gin.Context) {
	memcacheIsRunning := checkingMemcache()

	return func (ctx *gin.Context) {
		var currUser *workflow.User
		cookie, err := ctx.Request.Cookie("cention-suiteSSID")
		if err != nil {
			ctx.AbortWithStatus(HTTP_FORBIDDEN_ACCESS)
			return
		}
		var id int = -1
		ssid := cookie.Value
		if memcacheIsRunning {
			if err := gobcache.GetFromMemcache("Session_"+ssid, &id); err != nil {
				log.Println("[Set(Cookie)] Memcache key `Session` is empty!", err)
			}
			if id != -1 {
				currUser = controllers.FetchUserObject(id)
			} else {
				wfUser, err := wf.QueryUser_byHashLogin(ssid)
				if err != nil {
					ctx.AbortWithStatus(HTTP_UNAUTHORIZE_ACCESS)
				}
				if wfUser != nil {
					if wfUser.Active {
						if err := gobcache.SaveInMemcache("Session_"+ssid, wfUser.Id); err != nil {
							log.Println("[`SaveInMemcache`] Error on saving:", err)
						}
						currUser = controllers.FetchUserObject(wfUser.Id)
					} else {
						ctx.AbortWithStatus(HTTP_UNAUTHORIZE_ACCESS)
					}
				} else {
					ctx.AbortWithStatus(HTTP_UNAUTHORIZE_ACCESS)
				}
			}
	
		} else {
			wfUser, err := wf.QueryUser_byHashLogin(ssid)
			if err != nil {
				ctx.AbortWithStatus(HTTP_UNAUTHORIZE_ACCESS)
			}
			if wfUser != nil {
				if wfUser.Active {
					currUser = controllers.FetchUserObject(wfUser.Id)
				} else {
					ctx.AbortWithStatus(HTTP_UNAUTHORIZE_ACCESS)
				}
			} else {
				ctx.AbortWithStatus(HTTP_UNAUTHORIZE_ACCESS)
			}
		}
		ctx.Keys = make(map[string]interface{})
		ctx.Keys["loggedInUser"] = currUser
		ctx.Next()
	}
}
