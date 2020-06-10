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

// InventoryRole checks role
func InventoryRole(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userId := session.GetInt(r.Context(), "userID")
		ok := checkRole("inventory", userId)
		if ok {
			next.ServeHTTP(w, r)
		} else {
			helpers.ClientError(w, http.StatusUnauthorized)
		}
	})
}

// CreditRole checks role
func CreditRole(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userId := session.GetInt(r.Context(), "userID")
		ok := checkRole("credit", userId)
		if ok {
			next.ServeHTTP(w, r)
		} else {
			helpers.ClientError(w, http.StatusUnauthorized)
		}
	})
}

// FinderRole checks role
func FinderRole(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userId := session.GetInt(r.Context(), "userID")
		ok := checkRole("finder", userId)
		if ok {
			next.ServeHTTP(w, r)
		} else {
			helpers.ClientError(w, http.StatusUnauthorized)
		}
	})
}

// StaffRole checks role
func StaffRole(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userId := session.GetInt(r.Context(), "userID")
		ok := checkRole("staff", userId)
		if ok {
			next.ServeHTTP(w, r)
		} else {
			helpers.ClientError(w, http.StatusUnauthorized)
		}
	})
}

// EmailRole checks role
func EmailRole(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userId := session.GetInt(r.Context(), "userID")
		ok := checkRole("email", userId)
		if ok {
			next.ServeHTTP(w, r)
		} else {
			helpers.ClientError(w, http.StatusUnauthorized)
		}
	})
}

// TestDriveRole checks role
func TestDriveRole(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userId := session.GetInt(r.Context(), "userID")
		ok := checkRole("test_drive", userId)
		if ok {
			next.ServeHTTP(w, r)
		} else {
			helpers.ClientError(w, http.StatusUnauthorized)
		}
	})
}

// WordRole checks role
func WordRole(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userId := session.GetInt(r.Context(), "userID")
		ok := checkRole("word", userId)
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
