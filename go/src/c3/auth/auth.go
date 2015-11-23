package auth

/*
 * Implements auth middleware for cention applications.
 * Expects to be running in Gin framework
 */

import (
	"c3/ferite"
	"c3/ferite/serialize"
	wf "c3/osm/webframework"
	"c3/web/controllers"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gobcache"
	"github.com/gorilla/securecookie"
)

const (
	HTTP_UNAUTHORIZE_ACCESS = 401
	HTTP_FORBIDDEN_ACCESS   = 403
	HTTP_FOUND              = 302
)

var (
	hashKey                       = securecookie.GenerateRandomKey(64)
	blockKey                      = securecookie.GenerateRandomKey(32)
	sc                            = securecookie.New(hashKey, blockKey)
	ERROR_USER_PASS_MISMATCH      = errors.New("Username or Password doesnt match!")
	ERROR_USER_PASS_EMPTY         = errors.New("Username or Password is empty!")
	ERROR_CACHE_MISSED            = errors.New("Cache missed somehow!")
	ERROR_ON_SECURECOOKIE_HASHKEY = errors.New("HashKey doesn't match with secureCookie")
	ERROR_COOKIE_NOT_FOUND        = errors.New("Cookie: cention-suiteSSID=? not found")
	ERROR_WF_USER_NULL            = errors.New("webframework user is null")
	ERROR_MEMCACHE_FAILED         = errors.New("Memcache not running")
)

type AuthCookieManager struct {
	UserId        int
	LastLoginTime int64
	LoggedIn      bool
}

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

func decodeCookie(ctx *gin.Context) (string, error) {
	cookie, err := getCookieHashKey(ctx)
	if err != nil {
		return "", err
	}
	if cookie != "" && cookie != "guest" {
		return cookie, nil
		//TODO Mujibur: Do we need to verify the cookie? Not sure? Why?
		// Because if somehow webserver restarted the securecookie got vanish from memory
		// So its tried to create new. :)
		//		value := make(map[string]interface{})
		//		if err = sc.Decode("cention-suiteSSID", cookie, &value); err == nil {
		//			log.Printf("Cookie[`cention-suiteSSID`: %v] and params %v", cookie, value)
		//			return nil
		//		}
	}
	return "", ERROR_COOKIE_NOT_FOUND
}

func CheckOrCreateAuthCookie(ctx *gin.Context) error {
	ssid, err := decodeCookie(ctx)
	if err != nil {
		err := createNewAuthCookie(ctx)
		if err != nil {
			return err
		}
		return nil
	}
	user := ctx.Request.FormValue("username")
	pass := ctx.Request.FormValue("password")
	if user == "" && pass == "" {
		acm := new(AuthCookieManager)
		if checkingMemcache() {
			if err := gobcache.GetFromMemcache("Session_"+ssid, &acm); err != nil {
				log.Println("[GetFromMemcache] key `Session` is empty!")
				return err
			}
			if acm.LoggedIn && acm.UserId != 0 {
				return nil
			}
		}
		return ERROR_USER_PASS_EMPTY
	} else {
		log.Println("!!-- Seiting cookie informations to memcache. First time login.")
		wfUser, err := validateUser(user, pass)
		if err != nil {
			log.Println("Error on CheckOrCreateAuthCookie() - validateUser: ", err)
			return err
		}
		if wfUser != nil {
			lastLoginTime := time.Now().Unix()
			acm := new(AuthCookieManager)
			acm.UserId = wfUser.Id
			acm.LoggedIn = true
			acm.LastLoginTime = lastLoginTime
			if checkingMemcache() {
				if err = gobcache.SaveInMemcache("Session_"+ssid, acm); err != nil {
					log.Println("[`SaveInMemcache`] Error on saving:", err)
					return err
				}
				log.Printf("CentionAuth: User `%s` just now Logged In", wfUser.Username)
				return nil
			} else {
				return ERROR_MEMCACHE_FAILED
			}
		}
	}
	return ERROR_WF_USER_NULL
}

func validateUser(user, pass string) (*wf.User, error) {
	wu, err := wf.QueryUser_byLogin(user)
	if err != nil {
		return nil, err
	}
	if wu != nil {
		if wu.Active && wu.Password == encodePassword(pass) {
			return wu, nil
		}
	}
	return nil, ERROR_USER_PASS_MISMATCH
}
func encodePassword(p string) string {
	ep := sha256.New()
	_, err := ep.Write([]byte(p))
	if err != nil {
		log.Println("encodePassword() - Error on Sha256: ", err)
		return ""
	}
	return base64.StdEncoding.EncodeToString(ep.Sum(nil))
}
func createNewAuthCookie(ctx *gin.Context) error {
	value := map[string]interface{}{
		"cookie-set-date": time.Now().Unix(),
	}
	encoded, err := sc.Encode("cention-suiteSSID", value)
	if err != nil {
		log.Printf("createNewAuthCookie(): Error %v, creating `guest` cookie", err)
		cookie := fmt.Sprintf("cention-suiteSSID=%s; Path=/", "guest")
		ctx.Writer.Header().Add("Set-Cookie", cookie)
		return err
	}
	cookie := fmt.Sprintf("cention-suiteSSID=%s; Path=/", encoded)
	ctx.Writer.Header().Add("Set-Cookie", cookie)
	return nil
}
func CheckAuthCookie(ctx *gin.Context) (bool, error) {
	cookie, err := decodeCookie(ctx)
	if err != nil {
		return false, err
	}
	//	acm := &AuthCookieManager{
	//		UserId:   2,
	//		LoggedIn: true,
	//	}
	//	s, err := serialize.ToNative("SSID", acm)
	//	tests := struct {
	//		name string
	//		list interface{}
	//		want string
	//	}{
	//		name: "integer number",
	//		list: 42,
	//	}
	ss := ferite.NewArray(2, 559922377, true)
	s, err := serialize.ToNative("SSID", ss)
	if err != nil {
		log.Println("Error:", err)
	}
	log.Println("FFSession_" + cookie)
	err = gobcache.SetRawToMemcache("FFSession_"+cookie, interface{}(s))
	if err != nil {
		log.Println("Errorw:", err)
	}
	ok, err := validateByBrowserCookie(cookie)
	if err != nil {
		return false, err
	}
	if !ok {
		return false, err
	}
	return true, nil
}

func validateByBrowserCookie(ssid string) (bool, error) {
	acm := new(AuthCookieManager)
	if checkingMemcache() {
		if err := gobcache.GetFromMemcache("Session_"+ssid, &acm); err != nil {
			log.Println("[GetFromMemcache] key `Session` is empty!")
			return false, err
		}
		if acm.LoggedIn && acm.UserId != 0 {
			log.Println("Time after", acm.LastLoginTime)
			if err := updateTimeStampToMemcacheSSID(ssid, acm.UserId, true); err != nil {
				log.Printf("Error on validateByBrowserCookie(): %v", err)
				return false, err
			}
			log.Println("Time after")
			log.Printf("CentionAuth: Cookie has vrified with this info: %v", acm)
			return true, nil
		}
		log.Printf("CentionAuth redirect to login: Cookie info: %v", acm)
		return false, nil
	}
	return false, ERROR_MEMCACHE_FAILED
	/*
		//TODO Mujibur: I wrote and kept intentionally for future :P
		// Please dont argue with that. because wanting to get rif of from database while logging in and session checks
		wfv, err := wf.QueryVoucher_byHashLogin(ssid)
		if err != nil {
			log.Println("validateByOnlyCookie(): Error on loading - wf.QueryVoucher_byHashLogin", err)
			return false, err
		}
		if wfv != nil {
			if wfv.Loggedin && wfv.Active {
				return true, nil
			}
		}
		//Because wfv null is not an error.
		return false, nil
	*/
}

//Not used now.
//Kept that for future if anyhow is needed
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
	ssid, err := decodeCookie(ctx)
	if err != nil {
		return err
	}
	if checkingMemcache() {
		acm := new(AuthCookieManager)
		if err = gobcache.GetFromMemcache("Session_"+ssid, &acm); err != nil {
			log.Println("[GetFromMemcache] key `Session` is empty!")
			return err
		}
		if acm.LoggedIn && acm.UserId != 0 {
			if err = updateTimeStampToMemcacheSSID(ssid, acm.UserId, false); err != nil {
				log.Printf("Error on destroyAuthCookie(): %v", err)
				return err
			}
			return nil
		}
	}
	return ERROR_MEMCACHE_FAILED
}
func Logout(ctx *gin.Context) error {
	return destroyAuthCookie(ctx)
}

func fetchCookieFromRequest(r *http.Request) (string, error) {
	cookie, err := r.Cookie("cention-suiteSSID")
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}
func GetWebframeworkUserFromRequest(r *http.Request) int {
	ssid, err := fetchCookieFromRequest(r)
	if err != nil {
		log.Printf("Error on getting SSID: %v", err)
		return 0
	}
	if ssid == "" {
		return 0
	}
	acm := new(AuthCookieManager)
	if checkingMemcache() {
		if err := gobcache.GetFromMemcache("Session_"+ssid, &acm); err != nil {
			log.Println("[GetFromMemcache] key `Session` is empty!")
			return 0
		}
		log.Printf("Acm: %v", acm)
		if !acm.LoggedIn {
			return 0
		}
		//Update timestamp for every request
		if err := updateTimeStampToMemcacheSSID(ssid, acm.UserId, true); err != nil {
			log.Printf("Error on %v", err)
			return 0
		}
		return acm.UserId
	}
	return 0
}

func updateTimeStampToMemcacheSSID(ssid string, uid int, loginStatus bool) error {
	acm1 := new(AuthCookieManager)
	acm := new(AuthCookieManager)
	lastTimeGetRequest := time.Now().Unix()
	acm.LastLoginTime = lastTimeGetRequest
	acm.UserId = uid
	acm.LoggedIn = loginStatus
	log.Println("Time after:", acm.LastLoginTime)
	if err := gobcache.SaveInMemcache("Session_"+ssid, acm); err != nil {
		log.Println("[`SaveInMemcache`] Error on saving:", err)
		return err
	}
	if err := gobcache.GetFromMemcache("Session_"+ssid, &acm1); err != nil {
		log.Println("[`SaveInMemcache`] Error on saving:", err)
		return err
	}
	log.Println("HHHHHHH:", acm1)
	return nil
}

func Middleware() func(*gin.Context) {
	return func(ctx *gin.Context) {
		if strings.HasPrefix(ctx.Request.RequestURI, "/debug/pprof/") {
			ctx.Next()
			return
		}
		wfUserId := GetWebframeworkUserFromRequest(ctx.Request)
		if wfUserId == 0 {
			ctx.AbortWithStatus(HTTP_UNAUTHORIZE_ACCESS)
			return
		}
		currUser := controllers.FetchUserObject(wfUserId)
		ctx.Keys = make(map[string]interface{})
		ctx.Keys["loggedInUser"] = currUser
		ctx.Next()
	}
}
