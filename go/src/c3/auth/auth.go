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
	HTTP_MOVED_PERMANENTLY  = 301
)

var (
	hashKey                        = securecookie.GenerateRandomKey(64)
	blockKey                       = securecookie.GenerateRandomKey(32)
	sc                             = securecookie.New(hashKey, blockKey)
	ERROR_USER_PASS_MISMATCH       = errors.New("Username or Password doesnt match!")
	ERROR_CACHE_MISSED             = errors.New("Cache missed somehow!")
	ERROR_ON_SECURECOOKIE_HASHKEY  = errors.New("HashKey doesn't match with secureCookie")
	ERROR_COOKIE_NOT_FOUND         = errors.New("Cookie: cention-suiteSSID=? not found")
	ERROR_WF_USER_NULL             = errors.New("webframework user is null")
	ERROR_ON_MULTIPLE_LOGIN_COOKIE = errors.New("User logged in from multiple place")
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
func getCookieHashKey(ctx *gin.Context) (string, error) {
	cookie, err := ctx.Request.Cookie("cention-suiteSSID")
	if err != nil {
		log.Println("getCookieHashKey(): Cookie is empty - ", err)
		return "", err
	}
	if cookie != nil && cookie.Value == "" && cookie.Value == "guest" {
		return "", ERROR_CACHE_MISSED
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
		if err = sc.Decode("cention-suiteSSID", cookie, &value); err == nil {
			log.Println("Cookie[`cention-suiteSSID`]", value["username"])
			return true, value["username"], nil
		} else {
			//Somehow cookie got malformed so need to invalidate
			log.Println("decodeCookie(): Error to decode ", err)
			InvalidateCookie(ctx)
			return false, "", err
		}
	} else {
		return false, "", err
	}
	return false, "", ERROR_COOKIE_NOT_FOUND
}
func validateByOnlyCookie(ssid string) (*wf.User, error) {
	wu, err := wf.QueryUser_byHashLogin(ssid)
	if err != nil {
		log.Println("validateByOnlyCookie(): Error on loading - wf.QueryUser_byLogin", err)
		return nil, err
	}
	if wu != nil {
		if wu.Active {
			return wu, nil
		}
	}
	return nil, ERROR_ON_SECURECOOKIE_HASHKEY

}
func InvalidateCookie(ctx *gin.Context) {
	ctx.Writer.Header().Add("Set-Cookie", "cention-suiteSSID=guest; Path=/;MaxAge=-1")
}
func CheckOrCreateAuthCookie(ctx *gin.Context) error {
	user := ctx.Request.FormValue("username")
	pass := ctx.Request.FormValue("password")
	mu := &sync.Mutex{}
	if ok, cookieUser, err := decodeCookie(ctx); err == nil && ok {
		ssid, _ := getCookieHashKey(ctx)
		cu, err := validateByOnlyCookie(ssid)
		if err != nil {
			log.Println("Error on CheckOrCreateAuthCookie() - validateByOnlyCookie(): ", err)
			return err
		}
		if cu != nil {
			log.Printf("CentionAuth: User `%s` has logedin now", cookieUser)
			return nil
		} else {
			log.Println("Error on CheckOrCreateAuthCookie() - validateByOnlyCookie(): ", err)
			InvalidateCookie(ctx)
			return ERROR_ON_MULTIPLE_LOGIN_COOKIE
		}
	} else if user == "" && pass == "" {
		//log.Println("Error on empty user/pass")
		return ERROR_USER_PASS_MISMATCH
	} else {
		expires := time.Now().Add(time.Minute * 135)
		wfUser, err := validateUser(user, pass)
		if err != nil {
			log.Println("Error on CheckOrCreateAuthCookie() - validateUser: ", err)
			return err
		}
		if wfUser != nil {
			value := map[string]string{
				"username": user,
			}
			mu.Lock()
			if encoded, err := sc.Encode("cention-suiteSSID", value); err == nil {
				wfUser.SetSsid(encoded)
				if err := wfUser.Save(); err != nil {
					log.Println("CheckOrCreateAuthCookie(): Error on saving webframework User: ", err)
				}
				//Saving the cookie to memcache
				if checkingMemcache() {
					if err = gobcache.SaveInMemcache("Session_"+encoded, wfUser.Id); err != nil {
						log.Println("[`SaveInMemcache`] Error on saving:", err)
					}
				}
				log.Printf("CentionAuth: User `%s` just now Logged In", wfUser.Username)
				cookie := fmt.Sprintf("cention-suiteSSID=%s; Path=/;expires=%s", encoded, expires)
				ctx.Writer.Header().Add("Set-Cookie", cookie)
				defer mu.Unlock()
				return nil
			} else {
				log.Println("CheckOrCreateAuthCookie(): Error on wrong cookie - ", err)
				return err
			}
		}
		return ERROR_USER_PASS_MISMATCH
	}
}
func CheckAuthCookie(ctx *gin.Context) error {
	ssid, err := getCookieHashKey(ctx)
	if err != nil {
		log.Println("Error on CheckAuthCookie() - getCookieHashKey(): ", err)
		return err
	}
	cu, err := validateByOnlyCookie(ssid)
	if err != nil {
		return err
	}
	if cu == nil {
		log.Println("Error on CheckAuthCookie() - Webframework user nil: ")
		return ERROR_ON_MULTIPLE_LOGIN_COOKIE
	}
	return nil
}
func validateUser(user, pass string) (*wf.User, error) {
	if user != "" && pass != "" {
		wu, err := wf.QueryUser_byLogin(user)
		if err != nil {
			log.Println("validateUser() - Error on loading: wf.QueryUser_byLogin", err)
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
	return nil, ERROR_USER_PASS_MISMATCH
}
func encodePassword(p string) string {
	ep := sha256.New()
	_, err := ep.Write([]byte(p))
	if err != nil {
		log.Println("encodePassword() - Error on Sha256: ", err)
	}
	return base64.StdEncoding.EncodeToString(ep.Sum(nil))
}
func validateUserByCookie(secureCookieUser string, ctx *gin.Context) (*wf.User, error) {
	if cookie, err := getCookieHashKey(ctx); err == nil {
		wu, err := wf.QueryUser_byHashLogin(cookie)
		if err != nil {
			log.Println("validateUserByCookie() - Error on loading: wf.QueryUser_byLogin", err)
			return nil, err
		}
		if wu != nil {
			if wu.Username == secureCookieUser { //Need to make sure the cookie hash and secureCookie name is same
				return wu, nil
			}
		}
		return nil, ERROR_ON_SECURECOOKIE_HASHKEY
	}
	return nil, ERROR_WF_USER_NULL
}
func destroyAuthCookie(ctx *gin.Context) error {
	if ok, username, err := decodeCookie(ctx); err == nil && ok {
		wfUser, err := validateUserByCookie(username, ctx)
		if err != nil {
			log.Println("destroyAuthCookie() - Error on validateUserByCookie: ", err)
		}
		if wfUser != nil && wfUser.Active {
			wfUser.SetSsid("")
			if err := wfUser.Save(); err != nil {
				log.Println("destroyAuthCookie() - Error on saving webframework User: ", err)
			}
			ssid, _ := getCookieHashKey(ctx) //Error ignored!
			if checkingMemcache() {
				gobcache.DeleteFromMemcache("Session_" + ssid)
			}
			log.Printf("User `%s` just now Logged Out", wfUser.Username)
			ctx.Writer.Header().Add("Set-Cookie", "cention-suiteSSID=guest; Path=/;MaxAge=-1")
		}
		return nil
	} else {
		log.Println("destroyAuthCookie() - Error to decode ", err)
		return err
	}
}
func Logout(ctx *gin.Context) error {
	return destroyAuthCookie(ctx)
}
func Middleware(pathPrefix string) func(*gin.Context) {
	return func(ctx *gin.Context) {
		if strings.HasPrefix(ctx.Request.RequestURI, "/debug/pprof/") {
			ctx.Next()
			return
		}
		var currUser *workflow.User
		ssid, err := getCookieHashKey(ctx)
		if err != nil {
			log.Println("Middleware() - getCookieHashKey(): Error ", err)
			ctx.AbortWithStatus(HTTP_FORBIDDEN_ACCESS)
			ctx.Redirect(HTTP_MOVED_PERMANENTLY, pathPrefix+"login")
		}
		var id int = -1
		var wfUser *wf.User
		if checkingMemcache() {
			if err = gobcache.GetFromMemcache("Session_"+ssid, &id); err != nil {
				log.Println("[Set(Cookie)] Memcache key `Session` is empty!", err)
			}
			if id != -1 {
				currUser = controllers.FetchUserObject(id)
				wfUser, err = wf.QueryUser_byHashLogin(ssid)
				if err != nil {
					log.Println("Middleware() - QueryUser_byHashLogin(): Cookie not found in Memcache: ", err)
					ctx.AbortWithStatus(HTTP_UNAUTHORIZE_ACCESS)
					ctx.Redirect(HTTP_MOVED_PERMANENTLY, pathPrefix+"login")
				}
			} else {
				wfUser, err = wf.QueryUser_byHashLogin(ssid)
				if err != nil {
					log.Println("Middleware() - QueryUser_byHashLogin(): Cookie not matched in database: ", err)
					ctx.AbortWithStatus(HTTP_UNAUTHORIZE_ACCESS)
					ctx.Redirect(HTTP_MOVED_PERMANENTLY, pathPrefix+"login")
				}
				if wfUser != nil {
					if wfUser.Active {
						if err = gobcache.SaveInMemcache("Session_"+ssid, wfUser.Id); err != nil {
							log.Println("[`SaveInMemcache`] Error on saving:", err)
						}
						currUser = controllers.FetchUserObject(wfUser.Id)
					} else {
						log.Println("Middleware() - Creating session - Webframework user is not active, User Id: ", wfUser.Id)
						ctx.AbortWithStatus(HTTP_UNAUTHORIZE_ACCESS)
						ctx.Redirect(HTTP_MOVED_PERMANENTLY, pathPrefix+"login")
					}
				} else {
					log.Println("Middleware() - Creating session - Webframework user is nil ")
					ctx.AbortWithStatus(HTTP_UNAUTHORIZE_ACCESS)
					ctx.Redirect(HTTP_MOVED_PERMANENTLY, pathPrefix+"login")
				}
			}
		} else {
			wfUser, err = wf.QueryUser_byHashLogin(ssid)
			if err != nil {
				log.Println("Middleware() - QueryUser_byHashLogin(): ", err)
				ctx.AbortWithStatus(HTTP_UNAUTHORIZE_ACCESS)
				ctx.Redirect(HTTP_MOVED_PERMANENTLY, pathPrefix+"login")
			}

			if wfUser != nil {
				if wfUser.Active {
					currUser = controllers.FetchUserObject(wfUser.Id)
				} else {
					log.Println("Middleware() - Webframework user is not active, User Id: ", wfUser.Id)
					ctx.AbortWithStatus(HTTP_UNAUTHORIZE_ACCESS)
					ctx.Redirect(HTTP_MOVED_PERMANENTLY, pathPrefix+"login")
				}

			} else {
				log.Println("Middleware() - Webframework user is nil")
				/*Resolve from Headers were already written. Wanted to override status code 301 with 200*/
				ctx.AbortWithStatus(HTTP_UNAUTHORIZE_ACCESS)
				ctx.Redirect(HTTP_MOVED_PERMANENTLY, pathPrefix+"login")
			}
		}

		ctx.Keys = make(map[string]interface{})
		ctx.Keys["loggedInUser"] = currUser
		ctx.Keys["wfUser"] = wfUser
		ctx.Next()
	}
}
