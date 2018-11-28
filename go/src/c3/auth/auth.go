package auth

/*
 * Implements auth middleware for cention applications.
 * Expects to be running in Gin framework
 */

import (
	"c3/cloud"
	"c3/logger"
	"c3/osm/webframework"
	"c3/osm/workflow"
	"c3/space"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/cention-mujibur-rahman/gobcache"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/securecookie"
	uuid "github.com/satori/go.uuid"
)

var (
	setSecureCookieFlag = true
)

func init() {
	_, err := os.Stat("/cention/var/web.allow.http")
	if err == nil {
		// file exist, do not set "Secure" cookie flag
		setSecureCookieFlag = false
	}
}

type contextKey int

const (
	HTTP_UNAUTHORIZE_ACCESS = 401
	HTTP_FORBIDDEN_ACCESS   = 403
	HTTP_FOUND              = 302
	DayEpoch                = 86400
	CookieExpireAt          = DayEpoch * 30

	userContextKey contextKey = 1
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

func checkingMemcache(log logger.Logger) bool {
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

func getCurrentSession(log logger.Logger, v []byte) (int, int, bool, error) {
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

func fetchFromCache(log logger.Logger, key string) error {
	skey := "Session_" + key
	sItems, err := sessiond.GetRawFromMemcache(skey)
	if err != nil {
		log.Println("[GetRawFromMemcache] key `Session` is empty!")
		return err
	}
	if sItems != nil {
		uid, _, currentlyLogedin, err := getCurrentSession(log, sItems.Value)
		if err != nil {
			return err
		}
		if uid != 0 && currentlyLogedin {
			return nil
		}
	}
	return ERROR_CACHE_MISSED
}

func CreateAuthCookie(ctx *gin.Context, user *workflow.User) bool {
	c3ctx := ctx.Request.Context()
	log := logger.FromContext(c3ctx)
	ssid, err := createNewAuthCookie(ctx)
	if err != nil {
		log.Println(err)
		return false
	}
	lastLoginTime := time.Now().Unix()
	if checkingMemcache(log) {
		sValue := fmt.Sprintf("%v/%v/%v", user.WebframeworkUserID,
			lastLoginTime, true)
		if err = saveToSessiondCache(log, ssid, sValue); err != nil {
			log.Println(err)
			return false
		}
		if err = saveUserIdToCache(log, user.WebframeworkUserID,
			ssid); err != nil {
			log.Println(err)
			return false
		}
		updateUserCurrentLoginIn(c3ctx, user.WebframeworkUserID)
		log.Printf("CentionAuth: User `%s` auto logged In", user.Username)
		cu := FetchUserObject(ctx, user.WebframeworkUserID)
		ctx.Set("loggedInUser", cu)
		ctx.Next()
		return true
	} else {
		return false
	}
}

func FetchUserObject(ctx *gin.Context, wfUId int) *workflow.User {
	c3ctx := ctx.Request.Context()
	user, err := workflow.QueryUser_byWebframeworkUser(c3ctx, wfUId)
	if err != nil {
		log := logger.FromContext(c3ctx)
		log.Printf("error QueryUser_byWebframeworkUser(%d): %s", wfUId, err)
	}
	return user
}

func CheckOrCreateAuthCookie(ctx *gin.Context) error {
	c3ctx := ctx.Request.Context()
	log := logger.FromContext(c3ctx)

	var isOTPLogin bool
	if v, exists := ctx.Get("isOTPLogin"); exists {
		isOTPLogin, _ = v.(bool)
	}

	var user, cloudUsername, pass, ssid string
	var err error
	if isOTPLogin {
		cloudUsername = ctx.Value("otpUsername").(string)
		user, _, err = cloud.SplitUsername(cloudUsername)
		if err != nil {
			log.Printf("FIXME this should not happen invalid cloudUsername?? cloud.SplitUsername(%s): %s", cloudUsername, err)
		}
		pass = ctx.Value("otpPassword").(string)
		ssid, err = createNewAuthCookie(ctx)
	} else {
		user = strings.TrimSpace(ctx.Request.FormValue("username"))
		pass = ctx.Request.FormValue("password")
		ssid, err = decodeCookie(ctx)
	}

	if err != nil {
		_, err := createNewAuthCookie(ctx)
		if err != nil {
			return err
		}
		return nil
	}

	if user == "" && pass == "" {
		if isOTPLogin {
			// Undo set-cookie
			ctx.Writer.Header().Del("Set-Cookie")
		}
		if checkingMemcache(log) {
			if err = fetchFromCache(log, ssid); err != nil {
				if ctx.Request.Method == http.MethodGet && err == memcache.ErrCacheMiss {
					// Browser gave obsolete cookie, give it a new one
					ssid, _ = createNewAuthCookie(ctx)
					if len(ssid) > 0 {
						// And remember this cookie
						updateTimeStampToCache(log, ssid, 0, false)
					}
				}
				return err
			}
			return nil
		}
		return ERROR_USER_PASS_EMPTY
	} else {
		log.Println("!!-- Setting cookie informations to memcache. First time login.")
		wfUser, err := validateUser(c3ctx, user, pass)
		if err != nil {
			log.Println("Error on CheckOrCreateAuthCookie() - validateUser: ", err)
			return err
		}
		if wfUser != nil {
			if checkingMemcache(log) {
				cloudUsername := GetCloudUsername(ctx)
				return LogInUser(c3ctx, ctx, log, wfUser, cloudUsername, ssid)
			} else {
				return ERROR_MEMCACHE_FAILED
			}
		}
	}
	return ERROR_WF_USER_NULL
}

// LogInUser logs in the given user, no questions asked.
func LogInUser(c3ctx context.Context, ctx *gin.Context, log logger.Logger, wfUser *webframework.User, cloudUsername, ssid string) error {
	var err error
	if ssid == "" {
		ssid, err = createNewAuthCookie(ctx)
		if err != nil {
			return err
		}
	}
	lastLoginTime := time.Now().Unix()
	sValue := fmt.Sprintf("%v/%v/%v", wfUser.Id, lastLoginTime, true)
	if err := saveToSessiondCache(log, ssid, sValue); err != nil {
		return err
	}
	if err := saveUserIdToCache(log, wfUser.Id, ssid); err != nil {
		return err
	}
	updateUserCurrentLoginIn(c3ctx, wfUser.Id)
	log.Printf("CentionAuth: User `%s` just now Logged In", wfUser.Username)
	cu := FetchUserObject(ctx, wfUser.Id)
	ctx.Set("loggedInUser", cu)

	// remember the cloud username so that we can
	// later pre-fill the cloud login form at
	// cloud.cention.com when they log out later.
	if cloudUsername != "" {
		type loginData struct {
			Username string
		}
		ld := loginData{Username: cloudUsername}
		err := sessiond.SaveInMemcache(GetCloudCacheKey(ssid), ld)
		if err != nil {
			log.Printf("sessiond.SaveInMemcache: %v", err)
		}
	}

	return nil
}

func GetCloudCacheKey(ssid string) string {
	return "cloud_" + ssid
}

func GetCloudUsername(ctx *gin.Context) (cloudUsername string) {
	v, exist := ctx.Get("cloudUsername")
	if exist {
		cloudUsername, _ = v.(string)
	}
	return
}

func SetCloudUsername(ctx *gin.Context, cloudUsername string) {
	ctx.Set("cloudUsername", cloudUsername)
}

func generateAuthToken() string {
	retString := ""
	hData := make([]byte, 10)
	if _, err := rand.Read(hData); err == nil {
		hashed := sha512.Sum512(hData)
		retString = fmt.Sprintf("%x", hashed)
	}
	return retString
}

func updateUserCurrentLoginIn(c3ctx context.Context, wfUId int) {
	log := logger.FromContext(c3ctx)
	user, err := workflow.QueryUser_byWebframeworkUser(c3ctx, wfUId)
	if err != nil {
		log.Println("Error QueryUser_byWebframeworkUser: ", err)
	}
	user.SetTimestampLastLogin(time.Now().Unix())
	user.SetCurrentlyLoggedIn(true)
	user.SetAuthenticationToken(generateAuthToken())
	if err := user.Save(c3ctx); err != nil {
		log.Println("Error on Save: ", err)
	}
	updateUserStatusInHistory(c3ctx, user, "Login")
}
func updateUserCurrentLoginOut(c3ctx context.Context, wfUId int) {
	log := logger.FromContext(c3ctx)
	user, err := workflow.QueryUser_byWebframeworkUser(c3ctx, wfUId)
	if err != nil {
		log.Println("Error QueryUser_byWebframeworkUser: ", err)
	}
	user.SetTimestampLastLogout(time.Now().Unix())
	user.SetCurrentlyLoggedIn(false)
	user.SetAuthenticationToken("")
	if err := user.Save(c3ctx); err != nil {
		log.Println("Error on Save: ", err)
	}
	updateUserStatusInHistory(c3ctx, user, "Logout")
	log.Printf("- User %d logout", user.Id)
}
func updateUserStatusInHistory(c3ctx context.Context, user *workflow.User, status string) {
	spc := space.FromContext(c3ctx)
	newstat := workflow.NewUserStatusTrack(spc)
	currstat, _ := workflow.QueryUserStatus_byName(c3ctx, status)
	prestat, _ := workflow.QueryUserStatusTrack_getLastStatusByUserID(c3ctx, user.Id)
	chatstat, _ := workflow.QueryUserStatusTrack_getLastStatusByUserIDForChatOn(c3ctx, user.Id)
	timeLapsed := 0

	if user.AcceptChat {
		if chatstat != nil {
			timeLapsed = int(time.Now().Unix() - chatstat.TimestampCreate)
			if timeLapsed <= 24*60*60 && timeLapsed > 0 {
				chatstat.SetTimeSpent(timeLapsed)
				chatstat.Save(c3ctx)
			}
		}
	}
	if prestat != nil {
		if newstat != nil && currstat != nil && prestat.Status != currstat {
			newstat.SetUser(user)
			newstat.SetStatus(currstat)
			newstat.SetSystemGroup(user.SystemGroup)
			newstat.SetTimestampCreate(time.Now().Unix())
			newstat.SetLastUpdatedTimestamp(time.Now().Unix())
			newstat.Save(c3ctx)
		} else {
			prestat.SetTimeSpent(int(time.Now().Unix() - prestat.TimestampCreate))
			prestat.SetLastUpdatedTimestamp(time.Now().Unix())
			prestat.Save(c3ctx)
		}
	}
}
func saveUserIdToCache(log logger.Logger, key int, value string) error {
	sKey := fmt.Sprintf("user/%d", key)
	if err := sessiond.SetRawToMemcache(sKey, value); err != nil {
		log.Println("[`SetRawToMemcache`] Error on saving:", err)
		return err
	}
	return nil
}
func saveToSessiondCache(log logger.Logger, key, value string) error {
	sKey := "Session_" + key
	if err := sessiond.SetRawToMemcache(sKey, value); err != nil {
		log.Println("[`SetRawToMemcache`] Error on saving:", err)
		return err
	}
	return nil
}

func ValidateUser(c3ctx context.Context, user, pass string) (*webframework.User, error) {
	return validateUser(c3ctx, user, pass)
}

func validateUser(c3ctx context.Context, user, pass string) (*webframework.User, error) {
	log := logger.FromContext(c3ctx)
	wu, err := webframework.QueryUser_byLogin(c3ctx, user)
	if err != nil {
		return nil, err
	}
	if wu != nil {
		if wu.Active && wu.Password == encodePassword(log, pass) {
			return wu, nil
		}
	}
	return nil, ERROR_USER_PASS_MISMATCH
}
func encodePassword(log logger.Logger, p string) string {
	ep := sha256.New()
	_, err := ep.Write([]byte(p))
	if err != nil {
		log.Println("encodePassword() - Error on Sha256: ", err)
		return ""
	}
	return base64.StdEncoding.EncodeToString(ep.Sum(nil))
}
func createNewAuthCookie(ctx *gin.Context) (string, error) {
	c3ctx := ctx.Request.Context()
	log := logger.FromContext(c3ctx)
	value := map[string]interface{}{
		"uuid": uuid.NewV4().String(),
	}
	encoded, err := sc.Encode("cention-suiteSSID", value)
	if err != nil {
		log.Printf("createNewAuthCookie(): Error %v, creating `guest` cookie", err)
		cookie := fmt.Sprintf("cention-suiteSSID=%s; Path=/;HttpOnly", "guest")
		if setSecureCookieFlag {
			cookie += ";Secure"
		}
		ctx.Writer.Header().Add("Set-Cookie", cookie)
		return "", err
	}
	expiration := time.Now().Add(CookieExpireAt * time.Second).Format(time.RFC1123)
	cookie := fmt.Sprintf("cention-suiteSSID=%s; Path=/;Expires=%v;Max-Age=%d;HttpOnly", encoded, expiration, CookieExpireAt)
	if setSecureCookieFlag {
		cookie += ";Secure"
	}
	ctx.Writer.Header().Add("Set-Cookie", cookie)
	return encoded, nil
}
func CheckAuthCookie(ctx *gin.Context) (bool, error) {
	c3ctx := ctx.Request.Context()
	log := logger.FromContext(c3ctx)
	cookie, err := decodeCookie(ctx)
	if err != nil {
		return false, err
	}
	ok, err := validateByBrowserCookie(log, cookie)
	if err != nil {
		return false, err
	}
	if !ok {
		return false, err
	}
	return true, nil
}

func fetchFromCacheWithValue(log logger.Logger, key string) (int, int, bool, error) {
	skey := "Session_" + key
	sItems, err := sessiond.GetRawFromMemcache(skey)
	if err != nil {
		log.Println("[GetRawFromMemcache] key `Session` is empty!")
		return 0, 0, false, err
	}
	if sItems != nil {
		uid, timestamp, currentlyLogedin, err := getCurrentSession(log, sItems.Value)
		if err != nil {
			return 0, 0, false, err
		}
		if uid != 0 && currentlyLogedin {
			return uid, timestamp, currentlyLogedin, nil
		}
	}
	return 0, 0, false, ERROR_CACHE_MISSED
}

func validateByBrowserCookie(log logger.Logger, ssid string) (bool, error) {
	if checkingMemcache(log) {
		uid, _, currentlyLogedin, err := fetchFromCacheWithValue(log, ssid)
		if err != nil {
			return false, err
		}
		if !currentlyLogedin {
			log.Println("Cookie exist but user logged out by backend script")
			return false, err
		}
		if err := updateTimeStampToCache(log, ssid, uid, currentlyLogedin); err != nil {
			log.Printf("Error on validateByBrowserCookie(): %v", err)
			return false, err
		}
		log.Printf("CentionAuth: Cookie has verified with this info: %v", uid)
		return true, nil
	}
	return false, ERROR_MEMCACHE_FAILED
}

func destroyAuthCookie(ctx *gin.Context) (cloudUsername string, err error) {
	c3ctx := ctx.Request.Context()
	log := logger.FromContext(c3ctx)
	var ssid string
	ssid, err = decodeCookie(ctx)
	if err != nil {
		return
	}
	if checkingMemcache(log) {
		var uid int
		uid, _, _, err = fetchFromCacheWithValue(log, ssid)
		if err != nil {
			return
		}
		if err = updateTimeStampToCache(log, ssid, uid, false); err != nil {
			log.Printf("Error on validateByBrowserCookie(): %v", err)
			return
		}
		updateUserCurrentLoginOut(c3ctx, uid)
		cloudUsername, err = removeCloudUsernameFromMemcache(log, ssid)
		return
	}
	err = ERROR_MEMCACHE_FAILED
	return
}
func Logout(ctx *gin.Context) (string, error) {
	return destroyAuthCookie(ctx)
}

func removeCloudUsernameFromMemcache(log logger.Logger, ssid string) (cloudUsername string, err error) {
	cloudUsername, err = getCloudUsernameFromMemcache(log, ssid)
	sessiond.DeleteFromMemcache(GetCloudCacheKey(ssid))
	return
}

//GetCloudWorkspace provide cloud's workspace from memcache
func GetCloudWorkspace(ctx *gin.Context) (space string, err error) {
	c3ctx := ctx.Request.Context()
	log := logger.FromContext(c3ctx)
	var ssid string
	ssid, err = decodeCookie(ctx)
	if err != nil {
		return
	}
	user, err := getCloudUsernameFromMemcache(log, ssid)
	if err != nil {
		return
	}
	_, space, err = cloud.SplitUsername(user)
	return
}

func getCloudUsernameFromMemcache(log logger.Logger, ssid string) (cloudUsername string, err error) {
	type loginData struct {
		Username string
	}
	ld := loginData{}
	err = sessiond.GetFromMemcache(GetCloudCacheKey(ssid), &ld)
	if err != nil {
		if !strings.Contains(err.Error(), "memcache: cache miss") {
			log.Printf("sessiond.GetFromMemcache: %v", err)
		}
		return
	}
	cloudUsername = ld.Username
	return
}

func fetchCookieFromRequest(r *http.Request) (string, error) {
	cookie, err := r.Cookie("cention-suiteSSID")
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}
func GetWebframeworkUserFromRequest(r *http.Request) int {
	c3ctx := r.Context()
	log := logger.FromContext(c3ctx)
	ssid, err := fetchCookieFromRequest(r)
	if err != nil {
		return 0
	}
	if ssid == "" {
		return 0
	}
	if checkingMemcache(log) {
		uid, _, currentlyLogedin, err := fetchFromCacheWithValue(log, ssid)
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
		if err := updateTimeStampToCache(log, ssid, uid, currentlyLogedin); err != nil {
			log.Printf("Error on %v", err)
			return 0
		}
		return uid
	}
	return 0
}

func IsLoggedIn(r *http.Request) bool {
	c3ctx := r.Context()
	log := logger.FromContext(c3ctx)
	ssid, err := fetchCookieFromRequest(r)
	if err != nil || ssid == "" {
		// no cookie
		return false
	}
	if !checkingMemcache(log) {
		// memcache is not running?
		return false
	}
	_, _, loggedIn, _ := fetchFromCacheWithValue(log, ssid)
	return loggedIn
}

func updateTimeStampToCache(log logger.Logger, ssid string, uid int, loginStatus bool) error {
	lastTimeGetRequest := time.Now().Unix()
	svalue := fmt.Sprintf("%v/%v/%v", uid, lastTimeGetRequest, loginStatus)
	if err := saveToSessiondCache(log, ssid, svalue); err != nil {
		return err
	}
	return nil
}

func NewContextWithUser(parent context.Context, u *workflow.User) context.Context {
	return context.WithValue(parent, userContextKey, u)
}

func UserFromContext(ctx context.Context) *workflow.User {
	v := ctx.Value(userContextKey)
	u, ok := v.(*workflow.User)
	if ok {
		return u
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
			// /ng/logout implies that the user already logged in, but in
			// case the browser somehow tries to go to /ng/logout when they
			// are already logged out then we just redirect them to "/".
			// Without this they are going to get 401 unauthorized response
			// and the browser page (at /ng/logout) will remain blank.
			if strings.HasSuffix(ctx.Request.RequestURI, "/logout") {
				ctx.Redirect(http.StatusFound, "/")
				return
			}

			ctx.AbortWithStatus(HTTP_UNAUTHORIZE_ACCESS)
			return
		}
		var currUser *workflow.User
		if _, exist := ctx.Get("loggedInUser"); !exist {
			currUser = FetchUserObject(ctx, wfUserId)
			ctx.Set("loggedInUser", currUser)
		}
		ctx.Request = ctx.Request.WithContext(NewContextWithUser(ctx.Request.Context(), currUser))
		ctx.Next()
	}
}
