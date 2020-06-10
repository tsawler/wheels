package clienthandlers

import (
	"github.com/alexedwards/scs/v2"
	"github.com/tsawler/goblender/pkg/config"
	"github.com/tsawler/goblender/pkg/helpers"
	"net/http"
)

//var visitorQueue chan string
var session *scs.SessionManager
var serverName string
var live bool
var domain string
var preferenceMap map[string]string
var inProduction bool

// NewClientMiddleware sets app config for middleware
func NewClientMiddleware(app config.AppConfig) {
	serverName = app.ServerName
	live = app.InProduction
	domain = app.Domain
	preferenceMap = app.PreferenceMap
	session = app.Session
	inProduction = app.InProduction
}

// SomeRole is a sample role
func SomeRole(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userId := session.GetInt(r.Context(), "userID")
		ok := checkRole("some-role", userId)
		if ok {
			next.ServeHTTP(w, r)
		} else {
			helpers.ClientError(w, http.StatusUnauthorized)
		}
	})
}

// checkRole checks roles for the user
func checkRole(role string, userId int) bool {
	user, _ := repo.DB.GetUserById(userId)
	roles := user.Roles

	if _, ok := roles[role]; ok {
		return true
	}
	return false
}
