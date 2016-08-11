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
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/cention-mujibur-rahman/gobcache"
	"github.com/gin-gonic/gin"
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
	ERROR_MEMCACHE_FAILED         = errors.New("Sessiond not running")
)

var (
	sessiond = gobcache.NewCache("localhost:11311")
)

type AuthCookieManager struct {
	UserId        int
	LastLoginTime int64
	LoggedIn      bool
}

func checkingMemcache() bool {
	conn, err := net.Dial("tcp", "localhost:11311")
	if err != nil {
		log.Println("Sessiond Server is not running! ", err)
		return false
	}
	defer conn.Close()
	return true
}

func getCookieHashKey(ctx *gin.Context) (string, error) {
	var err error
	cookie, err := ctx.Request.Cookie("cention-suiteSSID")
	if err != nil {
		log.Println("getCookieHashKey(): Cookie is empty - ", err)
		return "", err
	}
	if cookie == nil {
		return "", ERROR_CACHE_MISSED
	}
	if cookie.Value == "" || cookie.Value == "guest" {
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
	}
	return "", ERROR_COOKIE_NOT_FOUND
}

func getCurrentSession(v []byte) (int, int, bool, error) {
	sv := string(v)
	sValue := strings.Split(sv, "/")
	if len(sValue) != 3 {
		return 0, 0, false, ERROR_CACHE_MISSED
	}
	uid, err := strconv.Atoi(sValue[0])
	if err != nil {
		log.Printf("Error on Uid conversion: %v", err)
		return 0, 0, false, err
	}
	ts, err := strconv.Atoi(sValue[1])
	if err != nil {
		log.Printf("Error on timestampLastLogin conversion: %v", err)
		return 0, 0, false, err
	}
	currentlyLoggedIn, err := strconv.ParseBool(sValue[2])
	if err != nil {
		log.Printf("Error on currentlyLoggedIn conversion: %v", err)
		return 0, 0, false, err
	}
	return uid, ts, currentlyLoggedIn, nil
}

func fetchFromCache(key string) error {
	skey := "Session_" + key
	sItems, err := sessiond.GetRawFromMemcache(skey)
	if err != nil {
		log.Println("[GetRawFromMemcache] key `Session` is empty!")
		return err
	}
	if sItems != nil {
		uid, _, currentlyLogedin, err := getCurrentSession(sItems.Value)
		if err != nil {
			return err
		}
		if uid != 0 && currentlyLogedin {
			return nil
		}
	}
	return ERROR_CACHE_MISSED
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
	user := strings.TrimSpace(ctx.Request.FormValue("username"))
	pass := ctx.Request.FormValue("password")
	if user == "" && pass == "" {
		if checkingMemcache() {
			if err = fetchFromCache(ssid); err != nil {
				return err
			}
			return nil
		}
		return ERROR_USER_PASS_EMPTY
	} else {
		log.Println("!!-- Setting cookie informations to memcache. First time login.")
		wfUser, err := validateUser(user, pass)
		if err != nil {
			log.Println("Error on CheckOrCreateAuthCookie() - validateUser: ", err)
			return err
		}
		if wfUser != nil {
			lastLoginTime := time.Now().Unix()
			if checkingMemcache() {
				sValue := fmt.Sprintf("%v/%v/%v", wfUser.Id, lastLoginTime, true)
				if err = saveToSessiondCache(ssid, sValue); err != nil {
					return err
				}
				if err = saveUserIdToCache(wfUser.Id, ssid); err != nil {
					return err
				}
				updateUserCurrentLoginIn(wfUser.Id)
				log.Printf("CentionAuth: User `%s` just now Logged In", wfUser.Username)
				return nil
			} else {
				return ERROR_MEMCACHE_FAILED
			}
		}
	}
	return ERROR_WF_USER_NULL
}
func updateUserCurrentLoginIn(wfUId int) {
	user, err := workflow.QueryUser_byWebframeworkUser(wfUId)
	if err != nil {
		log.Println("Error QueryUser_byWebframeworkUser: ", err)
	}
	user.SetTimestampLastLogin(time.Now().Unix())
	user.SetCurrentlyLoggedIn(true)
	if err := user.Save(); err != nil {
		log.Println("Error on Save: ", err)
	}
	updateUserStatusInHistory(user, "Login")
}
func updateUserCurrentLoginOut(wfUId int) {
	user, err := workflow.QueryUser_byWebframeworkUser(wfUId)
	if err != nil {
		log.Println("Error QueryUser_byWebframeworkUser: ", err)
	}
	user.SetTimestampLastLogout(time.Now().Unix())
	user.SetCurrentlyLoggedIn(false)
	if err := user.Save(); err != nil {
		log.Println("Error on Save: ", err)
	}
	updateUserStatusInHistory(user, "Logout")
	log.Printf("- User %d logout", user.Id)
}
func updateUserStatusInHistory(user *workflow.User, status string) {
	newstat := workflow.NewUserStatusTrack()
	currstat, _ := workflow.QueryUserStatus_byName(status)
	prestat, _ := workflow.QueryUserStatusTrack_getLastStatusByUserID(user.Id)
	chatstat, _ := workflow.QueryUserStatusTrack_getLastStatusByUserIDForChatOn(user.Id)
	timeLapsed := 0

	if user.AcceptChat {
		if chatstat != nil {
			timeLapsed = int(time.Now().Unix() - chatstat.TimestampCreate)
			if timeLapsed <= 24*60*60 && timeLapsed > 0 {
				chatstat.SetTimeSpent(timeLapsed)
				chatstat.Save()
			}
		}
	}
	if status == "Logout" && prestat != nil {
		prestat.SetTimeSpent(int(time.Now().Unix() - prestat.TimestampCreate))
		prestat.Save()
	}
	if newstat != nil {
		newstat.SetUser(user)
		newstat.SetStatus(currstat)
		newstat.SetSystemGroup(user.SystemGroup)
		newstat.SetTimestampCreate(time.Now().Unix())
		newstat.Save()
	}
}
func saveUserIdToCache(key int, value string) error {
	sKey := fmt.Sprintf("user/%d", key)
	if err := sessiond.SetRawToMemcache(sKey, value); err != nil {
		log.Println("[`SetRawToMemcache`] Error on saving:", err)
		return err
	}
	return nil
}
func saveToSessiondCache(key, value string) error {
	sKey := "Session_" + key
	if err := sessiond.SetRawToMemcache(sKey, value); err != nil {
		log.Println("[`SetRawToMemcache`] Error on saving:", err)
		return err
	}
	return nil
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
	ok, err := validateByBrowserCookie(cookie)
	if err != nil {
		return false, err
	}
	if !ok {
		return false, err
	}
	return true, nil
}

func fetchFromCacheWithValue(key string) (int, int, bool, error) {
	skey := "Session_" + key
	sItems, err := sessiond.GetRawFromMemcache(skey)
	if err != nil {
		log.Println("[GetRawFromMemcache] key `Session` is empty!")
		return 0, 0, false, err
	}
	if sItems != nil {
		uid, timestamp, currentlyLogedin, err := getCurrentSession(sItems.Value)
		if err != nil {
			return 0, 0, false, err
		}
		if uid != 0 && currentlyLogedin {
			return uid, timestamp, currentlyLogedin, nil
		}
	}
	return 0, 0, false, ERROR_CACHE_MISSED
}

func validateByBrowserCookie(ssid string) (bool, error) {
	if checkingMemcache() {
		uid, _, currentlyLogedin, err := fetchFromCacheWithValue(ssid)
		if err != nil {
			return false, err
		}
		if !currentlyLogedin {
			log.Println("Cookie exist but user logged out by backend script")
			return false, err
		}
		if err := updateTimeStampToCache(ssid, uid, currentlyLogedin); err != nil {
			log.Printf("Error on validateByBrowserCookie(): %v", err)
			return false, err
		}
		log.Printf("CentionAuth: Cookie has verified with this info: %v", uid)
		return true, nil
	}
	return false, ERROR_MEMCACHE_FAILED
}

func destroyAuthCookie(ctx *gin.Context) error {
	ssid, err := decodeCookie(ctx)
	if err != nil {
		return err
	}
	if checkingMemcache() {
		uid, _, _, err := fetchFromCacheWithValue(ssid)
		if err != nil {
			return err
		}
		if err := updateTimeStampToCache(ssid, uid, false); err != nil {
			log.Printf("Error on validateByBrowserCookie(): %v", err)
			return err
		}
		updateUserCurrentLoginOut(uid)
		return nil
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
	if checkingMemcache() {
		uid, _, currentlyLogedin, err := fetchFromCacheWithValue(ssid)
		if err != nil {
			if !currentlyLogedin {
				return 0
			} else {
				log.Println("Error: ", err)
			}
		}
		if !currentlyLogedin {
			return 0
		}
		//Update timestamp for every request
		if err := updateTimeStampToCache(ssid, uid, currentlyLogedin); err != nil {
			log.Printf("Error on %v", err)
			return 0
		}
		return uid
	}
	return 0
}

func updateTimeStampToCache(ssid string, uid int, loginStatus bool) error {
	lastTimeGetRequest := time.Now().Unix()
	svalue := fmt.Sprintf("%v/%v/%v", uid, lastTimeGetRequest, loginStatus)
	if err := saveToSessiondCache(ssid, svalue); err != nil {
		return err
	}
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
