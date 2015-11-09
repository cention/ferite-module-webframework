package auth

/*
 * Implements auth middleware for cention applications.
 * Expects to be running in Gin framework
 */

import (
	wf "c3/osm/webframework"
	"c3/osm/workflow"
	"c3/web/controllers"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/cention-mujibur-rahman/gobcache"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/securecookie"
)

const (
	HTTP_UNAUTHORIZE_ACCESS = 401
	HTTP_FORBIDDEN_ACCESS   = 403
)

var hashKey = securecookie.GenerateRandomKey(64)
var blockKey = securecookie.GenerateRandomKey(32)
var s = securecookie.New(hashKey, blockKey)

func checkingMemcache() bool {
	conn, err := net.Dial("tcp", "localhost:11211")
	if err != nil {
		log.Println("Memcache is not running! ", err)
		return false
	}
	defer conn.Close()
	return true
}
func getCookieHashKey(ctx *gin.Context) (string, error) {
	cookie, err := ctx.Request.Cookie("cention-suiteSSID")
	if err != nil {
		//Keep log for debugging
		//log.Println("Cookie is empty: ", err)
		return "", err
	}
	if cookie != nil && cookie.Value == "" && cookie.Value == "guest" {
		err = errors.New("Cache missed somehow!")
		return "", err
	}
	return cookie.Value, nil
}

func decodeCookie(ctx *gin.Context) (bool, string, error) {
	cookie, err := getCookieHashKey(ctx)
	if err != nil {
		return false, "", err
	}
	if cookie != "" && cookie != "guest" {
		value := make(map[string]string)
		if err = s.Decode("cention-suiteSSID", cookie, &value); err == nil {
			log.Println("Cookie[`cention-suiteSSID`]", value["username"])
			return true, value["username"], nil
		}
	} else {
		return false, "", err
	}
	err = errors.New("Cookie: cention-suiteSSID=? not found")
	return false, "", err
}
func CheckOrCreateAuthCookie(ctx *gin.Context) error {
	memcacheIsRunning := checkingMemcache()
	if ok, username, err := decodeCookie(ctx); err == nil && ok {
		log.Printf("-- User `%s` has login now", username)
		return nil
	} else {
		mu := &sync.Mutex{}
		user := ctx.Request.FormValue("username")
		pass := ctx.Request.FormValue("password")
		expires := time.Now().Add(time.Minute * 135)
		wfUser, err := validateUser(user, pass)
		if err != nil {
			return err
		}
		if wfUser != nil {
			value := map[string]string{
				"username": user,
			}
			mu.Lock()
			if encoded, err := s.Encode("cention-suiteSSID", value); err == nil {
				wfUser.SetSsid(encoded)
				if err := wfUser.Save(); err != nil {
					log.Println("Error on saving webframework User: ", err)
				}
				//Saving the cookie to memcache
				if memcacheIsRunning {
					if err = gobcache.SaveInMemcache("Session_"+encoded, wfUser.Id); err != nil {
						log.Println("[`SaveInMemcache`] Error on saving:", err)
					}
				}
				log.Printf("User `%s` just now Logged In", wfUser.Username)
				cookie := fmt.Sprintf("cention-suiteSSID=%s; Path=/;expires=%s", encoded, expires)
				ctx.Writer.Header().Add("Set-Cookie", cookie)
				mu.Unlock()
				return nil
			} else {
				log.Println("Error", err)
				return err
			}
		}
		err = errors.New("Username or Password doesnt match!")
		return err
	}
}
func validateUser(user, pass string) (*wf.User, error) {
	if user != "" && pass != "" {
		wu, err := wf.QueryUser_byLogin(user)
		if err != nil {
			log.Println("Error on loading: wf.QueryUser_byLogin", err)
			return nil, err
		}
		if wu != nil {
			if wu.Active && encodePassword(pass) == wu.Password {
				return wu, nil
			}
			log.Println("Either Cookie or (Username and Password) doesnt match  or User not active")
			return nil, nil
		}
	}
	msg := "!! Provided wrong Username or Password"
	if user == "" || pass == "" {
		msg = "!! Provided empty Username or Password."
	}
	log.Println(msg)
	err := errors.New("Username or password wrong")
	return nil, err
}
func encodePassword(p string) string {
	ep := sha256.New()
	_, err := ep.Write([]byte(p))
	if err != nil {
		log.Println("Error on Sha256: ", err)
	}
	return base64.StdEncoding.EncodeToString(ep.Sum(nil))
}
func validateUserByCookie(secureCookieUser string, ctx *gin.Context) (*wf.User, error) {
	if cookie, err := getCookieHashKey(ctx); err == nil {
		wu, err := wf.QueryUser_byHashLogin(cookie)
		if err != nil {
			log.Println("Error on loading: wf.QueryUser_byLogin", err)
			return nil, err
		}
		if wu != nil {
			if wu.Username == secureCookieUser { //Need to make sure the cookie hash and secureCookie name is same
				return wu, nil
			}
		}
		err = errors.New("HashKey doesn't match with secureCookie")
		return nil, err
	}
	err := errors.New("webframework user is null")
	return nil, err
}
func destroyAuthCookie(ctx *gin.Context) error {
	memcacheIsRunning := checkingMemcache()
	if ok, username, err := decodeCookie(ctx); err == nil && ok {
		wfUser, err := validateUserByCookie(username, ctx)
		if err != nil {
			log.Println("Error on validateUserByCookie: ", err)
		}
		if wfUser != nil && wfUser.Active {
			wfUser.SetSsid("")
			if err := wfUser.Save(); err != nil {
				log.Println("Error on saving webframework User: ", err)
			}
			ssid, _ := getCookieHashKey(ctx) //Error ignored!
			if memcacheIsRunning {
				gobcache.DeleteFromMemcache("Session_" + ssid)
			}
			log.Printf("User `%s` just now Logged Out", wfUser.Username)
			ctx.Writer.Header().Add("Set-Cookie", "cention-suiteSSID=guest; Path=/;MaxAge=-1")
		}
		return nil
	} else {
		//Nothing to do!!
		return err
	}
}
func Logout(ctx *gin.Context) error {
	return destroyAuthCookie(ctx)
}
func Middleware() func(*gin.Context) {
	memcacheIsRunning := checkingMemcache()

	return func(ctx *gin.Context) {
		if strings.HasPrefix(ctx.Request.RequestURI, "/debug/pprof/") {
			ctx.Next()
			return
		}
		var currUser *workflow.User
		cookie, err := ctx.Request.Cookie("cention-suiteSSID")
		if err != nil {
			ctx.AbortWithStatus(HTTP_FORBIDDEN_ACCESS)
			return
		}
		var id int = -1
		var wfUser *wf.User
		ssid := cookie.Value
		if memcacheIsRunning {
			if err = gobcache.GetFromMemcache("Session_"+ssid, &id); err != nil {
				log.Println("[Set(Cookie)] Memcache key `Session` is empty!", err)
			}
			if id != -1 {
				currUser = controllers.FetchUserObject(id)
				wfUser, err = wf.QueryUser_byHashLogin(ssid)
				if err != nil {
					ctx.AbortWithStatus(HTTP_UNAUTHORIZE_ACCESS)
				}
			} else {
				wfUser, err = wf.QueryUser_byHashLogin(ssid)
				if err != nil {
					ctx.AbortWithStatus(HTTP_UNAUTHORIZE_ACCESS)
				}
				if wfUser != nil {
					if wfUser.Active {
						if err = gobcache.SaveInMemcache("Session_"+ssid, wfUser.Id); err != nil {
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
			wfUser, err = wf.QueryUser_byHashLogin(ssid)
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
		ctx.Keys["wfUser"] = wfUser
		ctx.Next()
	}
}
