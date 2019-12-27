package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/rajendraventurit/radicaapi/domain"
	"github.com/rajendraventurit/radicaapi/lib/env"
	"github.com/rajendraventurit/radicaapi/lib/handler"
	"github.com/rajendraventurit/radicaapi/lib/routetable"
	"github.com/rajendraventurit/radicaapi/lib/serror"
	"github.com/rajendraventurit/radicaapi/lib/token"
)

// UserRoutes returns the domain routes
func UserRoutes(env *env.Env) routetable.RouteTable {
	rt := routetable.NewRouteTable()
	rt.Add(routetable.Route{
		Category: "User",
		Name:     "User login",
		Method:   "POST",
		Input:    `{"email": "name", "password": "abc"}`,
		Output:   `{"user_id": 0, "first_name": "", "last_name": "", "email": "", "created_on": "", "updated_on": "", "deleted": false, "token": "abc", "roles": [1]}`,
		Path:     "/api/v1/user/login",
		Handler:  handler.Handler{Env: env, Fn: HandleLogin},
		Insecure: true,
	},
		routetable.Route{
			Category: "User",
			Name:     "Create user",
			Method:   "POST",
			Input:    `{"first_name": "", "last_name": "", "email": "", "password": ""}`,
			Path:     "/api/v1/user",
			Handler:  handler.Handler{Env: env, Fn: HandleCreateUser},
			Insecure: true,
		},

		routetable.Route{
			Category: "User",
			Name:     "Users Activity",
			Method:   "POST",
			Input:    `{"device_id": "2fc4b5912826ad1", "activity_type": "open_app/dashboard/setting"}`,
			Path:     "/api/v1/user/activity",
			Handler:  handler.Handler{Env: env, Fn: HandleUserActivity},
			Insecure: true,
		},

		routetable.Route{
			Category: "User",
			Name:     "Users disease",
			Method:   "POST",
			Input:    `{ "disease": "test disease","symtoms": "test1,test2,test3","disease_date":"2019-08-22 11:05:05","dbm":25,"onscreen_time":4}`,
			Path:     "/api/v1/user/createdisease",
			Handler:  handler.Handler{Env: env, Fn: HandleCreateDisease},
		},

		routetable.Route{
			Category: "User",
			Name:     "Users disease",
			Method:   "GET",
			Input:    `{}`,
			Path:     "/api/v1/user/disease",
			Handler:  handler.Handler{Env: env, Fn: HandleGetDisease},
		},

		// routetable.Route{
		// 	Category:    "User",
		// 	Name:        "Delete user",
		// 	Method:      "DELETE",
		// 	Input:       `{"user_id": 345}`,
		// 	Path:        "/api/v1/user",
		// 	Handler:     handler.Handler{Env: env, Fn: HandleDeleteUser},
		// 	Permissions: []int64{domain.PermManageUsers},
		// },
		// routetable.Route{
		// 	Category:    "User",
		// 	Name:        "Change Password",
		// 	Method:      "PUT",
		// 	Input:       `{"password": ""}`,
		// 	Path:        "/api/v1/user/password",
		// 	Handler:     handler.Handler{Env: env, Fn: HandleChangePassword},
		// 	Permissions: []int64{domain.PermManageSelf},
		// },
		// routetable.Route{
		// 	Category: "User",
		// 	Name:     "Request reset password",
		// 	Method:   "POST",
		// 	Input:    `{"email": "", "reset_url": ""}`,
		// 	Path:     "/api/v1/user/password/reset",
		// 	Handler:  handler.Handler{Env: env, Fn: HandleRequestResetPassword},
		// 	Insecure: true,
		// },
		// routetable.Route{
		// 	Category: "User",
		// 	Name:     "Validate and reset password",
		// 	Method:   "PUT",
		// 	Input:    `{"email": "", "password": "", "token": ""}`,
		// 	Path:     "/api/v1/user/password/reset",
		// 	Handler:  handler.Handler{Env: env, Fn: HandleResetPassword},
		// 	Insecure: true,
		// },
		// routetable.Route{
		// 	Category:    "User",
		// 	Name:        "Get User",
		// 	Method:      "GET",
		// 	Input:       `?user_id=1`,
		// 	Path:        "/api/v1/user",
		// 	Handler:     handler.Handler{Env: env, Fn: HandleGetUser},
		// 	Permissions: []int64{domain.PermManageSelf},
		// },
	)
	return rt
}

// HandleLogin will login a user
func HandleLogin(env *env.Env, w http.ResponseWriter, r *http.Request) error {
	p := struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}{}
	if err := decodeJSON(r.Body, &p); err != nil {
		return err
	}
	user, err := domain.Authenticate(env.DB, p.Email, p.Password)
	if err != nil {
		return sendJSON1(w, "", false, "Username and password did not match", http.StatusUnauthorized)
	}
	//return sendJSON(w, user)
	return sendJSON1(w, user, true, "Loggedin Successfully", http.StatusOK)
}

// HandleCreateUser will create a new org and user
func HandleCreateUser(env *env.Env, w http.ResponseWriter, r *http.Request) error {
	p := struct {
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Email     string `json:"email"`
		Password  string `json:"password"`
	}{}
	if err := decodeJSON(r.Body, &p); err != nil {
		return err
	}
	defer r.Body.Close()

	if p.FirstName == "" && p.LastName == "" {
		return serror.NewBadRequest(fmt.Errorf("first / last name required"), "HandleCreteUser", "first / last name required")
	}

	// Create user
	usr, err := domain.CreateUser(env.DB, p.FirstName, p.LastName, p.Email, p.Password)
	if err != nil {
		return serror.Error{
			Code:    http.StatusBadRequest,
			Err:     err,
			Context: "user.Create",
			Msg:     err.Error(),
		}
	}
	usr.Password = ""
	return sendJSON(w, usr)
}

// HandleUserActivity will track users activity
func HandleUserActivity(env *env.Env, w http.ResponseWriter, r *http.Request) error {
	p := struct {
		DeviceID     string `json:"device_id"`
		ActivityType string `json:"activity_type"`
	}{}

	if err := decodeJSON(r.Body, &p); err != nil {
		return err
	}
	defer r.Body.Close()

	if p.DeviceID == "" && p.ActivityType == "" {
		return serror.NewBadRequest(fmt.Errorf("device_id / activity_type name required"), "HandleCreteUser", "device_id / activity_type  required")
	}

	// Create axctivity
	err := domain.CreateActivity(env.DB, p.DeviceID, p.ActivityType)
	if err != nil {
		return serror.Error{
			Code:    http.StatusBadRequest,
			Err:     err,
			Context: "user.Activity",
			Msg:     err.Error(),
		}
	}

	return nil

}

// HandleCreateDisease will create Disease of user
func HandleCreateDisease(env *env.Env, w http.ResponseWriter, r *http.Request) error {

	claims, err := token.AuthToken(r)

	fmt.Println("err", err)

	if err != nil {
		return serror.New(http.StatusUnauthorized, err, "token.AuthToken", "")
	}

	p := struct {
		Disease      string `json:"disease"`
		Symtoms      string `json:"symtoms"`
		DiseaseDate  string `json:"disease_date"`
		Dbm          int64  `json:"dbm"`
		OnscreenTime int64  `json:"onscreen_time"`
	}{}

	if err := decodeJSON(r.Body, &p); err != nil {
		return err
	}
	defer r.Body.Close()

	// Create axctivity
	err = domain.CreateDisease(env.DB, p.Disease, p.Symtoms, p.DiseaseDate, p.Dbm, p.OnscreenTime, claims.UserID)
	if err != nil {
		return serror.Error{
			Code:    http.StatusBadRequest,
			Err:     err,
			Context: "user.Disease",
			Msg:     err.Error(),
		}
	}

	return nil

}

// HandleGetDisease will get all the disease
func HandleGetDisease(env *env.Env, w http.ResponseWriter, r *http.Request) error {

	claims, err := token.AuthToken(r)

	defer r.Body.Close()

	// get disease of user
	dis, err := domain.GetUserDisease(env.DB, claims.UserID)
	if err != nil {
		return serror.Error{
			Code:    http.StatusBadRequest,
			Err:     err,
			Context: "user.Disease",
			Msg:     err.Error(),
		}
	}

	return sendJSON1(w, dis, true, "Get Disease Successfully", http.StatusOK)

}

// HandleDeleteUser will delete a users
func HandleDeleteUser(env *env.Env, w http.ResponseWriter, r *http.Request) error {
	p := struct {
		UserID int64 `json:"user_id"`
	}{}
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		return err
	}
	defer r.Body.Close()
	if p.UserID == 0 {
		return serror.NewBadRequest(fmt.Errorf("Invalid user_id"), "HandleDeleteUser", "invalid user_id")
	}
	if err := domain.MarkUserDeleted(env.DB, p.UserID); err != nil {
		return err
	}
	return nil
}

// HandleChangePassword will change a users password
func HandleChangePassword(env *env.Env, w http.ResponseWriter, r *http.Request) error {
	claims, err := token.AuthToken(r)
	if err != nil {
		return serror.New(http.StatusUnauthorized, err, "token.AuthToken", "")
	}
	p := struct {
		Password string `json:"password"`
	}{}
	if err := decodeJSON(r.Body, &p); err != nil {
		return err
	}
	defer r.Body.Close()
	if err := domain.UpdatePassword(env.DB, claims.UserID, p.Password); err != nil {
		return serror.NewBadRequest(err, "domain.UpdatePassword", err.Error())
	}
	return nil
}

// HandleRequestResetPassword will send a reset password link to a user
func HandleRequestResetPassword(env *env.Env, w http.ResponseWriter, r *http.Request) error {
	p := struct {
		Email    string `json:"email"`
		ResetURL string `json:"reset_url"`
	}{}
	if err := decodeJSON(r.Body, &p); err != nil {
		return err
	}
	defer r.Body.Close()
	if err := domain.SendResetToken(env.DB, p.Email, p.ResetURL); err != nil {
		return serror.NewBadRequest(err, "domain.SendResetToekn", "Failed to send reset")
	}
	return nil
}

// HandleResetPassword will validate a reset token and change password
func HandleResetPassword(env *env.Env, w http.ResponseWriter, r *http.Request) error {
	p := struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Token    string `json:"token"`
	}{}
	if err := decodeJSON(r.Body, &p); err != nil {
		return err
	}
	defer r.Body.Close()
	if err := domain.ResetPassword(env.DB, p.Email, p.Token, p.Password); err != nil {
		return serror.NewBadRequest(err, "domain.ResetPassword", err.Error())
	}
	return nil
}

// HandleGetUser will return a user
func HandleGetUser(env *env.Env, w http.ResponseWriter, r *http.Request) error {
	uid, err := getQueryInt64(r, "user_id")
	if err != nil {
		return err
	}
	user, err := domain.GetUser(env.DB, uid)
	if err != nil {
		return serror.NewBadRequest(err, "domain.GetUser")
	}
	user.Password = ""
	return sendJSON(w, user)
}
