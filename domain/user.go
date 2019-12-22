package domain

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"text/template"
	"time"
	"unicode"

	"github.com/rajendraventurit/radicaapi/lib/db"
	"github.com/rajendraventurit/radicaapi/lib/smtp"
	"github.com/rajendraventurit/radicaapi/lib/token"
	"golang.org/x/crypto/bcrypt"
)

//Userpassword is an object of user password
type Userpassword struct {
	UserPasswordID int64  `db:"user_password_id" json:"user_password_id"`
	UserID         int64  `db:"user_id" json:"user_id"`
	Password       string `db:"password" json:"password"`
}

//UsersActivity keeps track of users activity
type UsersActivity struct {
	ID           int64  `db:"id" json:"id"`
	DeviceID     string `db:"device_id" json:"device_id"`
	ActivityType string `db:"activity_type" json:"activity_type"`
}

// CreateUser creates a user
func CreateUser(ex db.Execer, fn, ln, email, pass string, roles ...int64) (*User, error) {
	if err := validPassword(pass); err != nil {
		return nil, err
	}
	u := NewUser(fn, ln, email)
	hashed, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	u.Password = pass
	u.HashedPass = hashed

	err = u.Create(ex)
	return u, err
}

// CreateActivity creates activity of user
func CreateActivity(ex db.Execer, deviceid, activitytype string) error {

	str := `
		INSERT INTO users_activity
			(device_id, activity_type)
			VALUES
			(?, ?)
		`
	_, err := ex.Exec(str, deviceid, activitytype)
	if err != nil {
		return err
	}

	return nil
}

// UpdateUser will update a user
func UpdateUser(ex db.Execer, u User) error {
	return u.Update(ex)
}

// MarkUserDeleted will mark a user deleted
func MarkUserDeleted(ex db.Execer, userid int64) error {
	str := "UPDATE users SET deleted = true WHERE user_id = ?"
	_, err := ex.Exec(str, userid)
	return err
}

// Authenticate returns true if email/password match
func Authenticate(st db.Storer, email, pass string) (*User, error) {
	usr, err := GetUserWithEmail(st, email)
	if err != nil {
		return nil, err
	}
	if usr.Deleted {
		return nil, ErrUserDeleted
	}
	err = bcrypt.CompareHashAndPassword([]byte(usr.HashedPass), []byte(pass))
	if err != nil {
		return nil, err
	}
	tok, err := token.New(usr.UserID)
	if err != nil {
		return nil, err
	}
	usr.Token = tok
	return usr, nil
}

// UpdatePassword updates a users password and writes original to the history table
// It is independent of hashing and expects the password to be hashed
func UpdatePassword(ex db.Execer, userid int64, pass string) error {
	if err := validPassword(pass); err != nil {
		return err
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	str := `
	INSERT INTO user_passwords (user_id, password)
	SELECT user_id, password FROM users WHERE user_id = ?
	`
	_, err = ex.Exec(str, userid)
	if err != nil {
		return err
	}

	str = "UPDATE users SET password = ? WHERE user_id = ?"
	_, err = ex.Exec(str, hashed, userid)
	if err != nil {
		return err
	}
	return nil
}

func validPassword(pass string) error {

	var uppercasePresent bool
	var lowercasePresent bool
	var numberPresent bool
	var specialCharPresent bool

	var passLen int

	for _, ch := range pass {

		switch {
		case unicode.IsNumber(ch):
			numberPresent = true
			passLen++
		case unicode.IsUpper(ch):
			uppercasePresent = true
			passLen++
		case unicode.IsLower(ch):
			lowercasePresent = true
			passLen++
		case unicode.IsPunct(ch) || unicode.IsSymbol(ch):
			specialCharPresent = true
			passLen++
		case ch == ' ':
			passLen++
		}
	}

	if !(minPassLength <= passLen && passLen <= maxPassLength) {
		return fmt.Errorf(fmt.Sprintf("password length must be between %d to %d characters long", minPassLength, maxPassLength))
	}
	if !lowercasePresent {
		return fmt.Errorf("lowercase letter missing")
	}
	if !uppercasePresent {
		return fmt.Errorf("uppercase letter missing")
	}
	if !numberPresent {
		return fmt.Errorf("atleast one numeric character required")
	}
	if !specialCharPresent {
		return fmt.Errorf("special character missing")
	}

	return nil
}

//Ispasswordmatchwithprevfivepass return true if the new password matches with previous five password.
func Ispasswordmatchwithprevfivepass(qr db.Queryer, userid int64, pass string) bool {

	sqlstr := `select user_password_id,user_id,password from user_passwords where user_id=? order by user_password_id limit 5`

	passwords := []Userpassword{}
	err := qr.Select(&passwords, sqlstr, userid)

	fmt.Println("passwords", passwords)

	fmt.Println("err", err)

	if err != nil {
		log.Print("Ispasswordmatchwithprevfivepass", err)
		return false
	}

	for _, password := range passwords {
		err = bcrypt.CompareHashAndPassword([]byte(password.Password), []byte(pass))
		if err == nil {
			return true
		}
	}

	return false
}

// SendResetToken will send a reset token to a user
func SendResetToken(qr db.Queryer, email, resetURL string) error {
	t, err := genResetToken(qr, email)
	if err != nil {
		return err
	}
	tok := encodeResetToken(t)

	rurl := fmt.Sprintf("%s?t=%s", resetURL, tok)
	mailer := smtp.SMTP{}

	fname := fmt.Sprintf("%v/%v", defTemplatePath, "resetpass.html")
	tmp, err := template.ParseFiles(fname)
	if err != nil {
		return err
	}
	p := struct {
		Link string
	}{Link: rurl}
	var b bytes.Buffer
	err = tmp.Execute(&b, p)
	if err != nil {
		return err
	}
	return mailer.Send("Password Reset", b.String(), nil, email)
}

// ResetPassword will reset a users password validated with token
func ResetPassword(db db.Storer, email, tok, newpass string) error {
	valid, err := verifyResetToken(db, email, tok)
	if err != nil {
		return err
	}

	if !valid {
		return fmt.Errorf("Invalid reset token")
	}

	user, err := GetUserWithEmail(db, email)
	if err != nil {
		return err
	}
	if Ispasswordmatchwithprevfivepass(db, user.UserID, newpass) {
		return errors.New("History says that you have used this new password in the past")
	}

	return UpdatePassword(db, user.UserID, newpass)
}

// GetUser will return a user
func GetUser(qy db.Queryer, userid int64) (*User, error) {
	return GetUserWithID(qy, userid)
}

// UserList is a list of users
type UserList struct {
	Users  []User `json:"users"`
	Total  int64  `json:"total"`
	Offset int64  `json:"offset"`
	Limit  int64  `json:"limit"`
}

// Password is a historical user password
type Password struct {
	UserPasswordID int64     `db:"user_password_id" json:"user_password_id"`
	UserID         int64     `db:"user_id" json:"user_id"`
	Password       string    `db:"password" json:"password"`
	CreatedOn      time.Time `db:"created_on" json:"created_on"`
	UpdatedOn      time.Time `db:"updated_on" json:"updated_on"`
}

// Save will commit data to the database db
func (up Password) Save(ex db.Execer) error {
	str := `
	INSERT INTO user_passwords
	(user_id, password)
	VALUES
	(:user_id, :password)
	`
	_, err := ex.NamedExec(str, &up)
	return err
}

// User is a user
type User struct {
	UserID     int64         `db:"user_id" json:"user_id"`
	FirstName  db.NullString `db:"first_name" json:"first_name"`
	LastName   db.NullString `db:"last_name" json:"last_name"`
	Email      string        `db:"email" json:"email"`
	Password   string        `db:"-" json:"password,omitempty"`
	HashedPass []byte        `db:"password" json:"-"`
	CreatedOn  time.Time     `db:"created_on" json:"created_on"`
	UpdatedOn  time.Time     `db:"updated_on" json:"updated_on"`
	Deleted    bool          `db:"deleted" json:"deleted"`
	Token      string        `db:"-" json:"token,omitempty"`
	RolesStr   string        `db:"roles_str" json:"-"`
	Status     string        `json:"status"`
}

//OrganizationUser is object
type OrganizationUser struct {
	OrganizationID int64 `db:"organization_id" json:"organization_id"`
	UserID         int64 `db:"user_id" json:"user_id"`
	Role           int64 `db:"role" json:"role"`
}

// Mobility is a object of Mobility
type Mobility struct {
	ID   int64         `db:"id" json:"id"`
	Name db.NullString `db:"name" json:"name"`
}

// Location is a object of Location
type Location struct {
	ID   int64         `db:"id" json:"id"`
	Name db.NullString `db:"name" json:"name"`
}

// NewUser returns a user
func NewUser(fn, ln, email string) *User {
	return &User{
		FirstName: db.NewNullString(fn),
		LastName:  db.NewNullString(ln),
		Email:     email,
	}
}

// ErrUserDeleted is a deleted user error
var ErrUserDeleted = fmt.Errorf("User has been deleted")

// Create a user
func (u *User) Create(ex db.Execer) error {
	str := `
	INSERT INTO users
		(first_name, last_name, email, password)
		VALUES
		(:first_name, :last_name, :email, :password)
	`
	unhashed := u.Password
	hashed, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashed)
	resp, err := ex.NamedExec(str, u)
	if err != nil {
		return err
	}
	u.Password = unhashed
	id, err := resp.LastInsertId()
	if err != nil {
		return err
	}

	u.UserID = id

	return nil
}

// Update a user
func (u User) Update(ex db.Execer) error {
	str := `
	UPDATE users
	SET first_name = :first_name,
	last_name = :last_name,
	email = :email
	WHERE user_id = ?
	`
	_, err := ex.NamedExec(str, u)
	return err
}

// IsUserDeleted returns true if user has been deleted
func IsUserDeleted(qr db.Queryer, userid int64) bool {
	str := "SELECT deleted FROM users WHERE user_id = ?"
	del := false
	err := qr.Get(&del, str, userid)
	if err != nil {
		return false
	}
	return del
}

// GetUserWithID will return a user
func GetUserWithID(qr db.Queryer, userid int64) (*User, error) {
	str := "SELECT * FROM users WHERE user_id = ?"
	user := User{}
	err := qr.Get(&user, str, userid)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// GetUserWithEmail will return a user
func GetUserWithEmail(qr db.Queryer, email string) (*User, error) {
	str := "SELECT * FROM users WHERE email = ?"
	user := User{}
	err := qr.Get(&user, str, email)
	if err != nil {
		return nil, err
	}

	return &user, err
}

// GetUserID returns a user id from an email
func GetUserID(qr db.Queryer, email string) (int64, error) {
	str := "SELECT user_id FROM users WHERE email = ?"
	uid := int64(0)
	err := qr.Get(&uid, str, email)
	return uid, err
}

//IsUserExist will return true if user exist
func IsUserExist(qr db.Queryer, email string) bool {
	str := "SELECT count(*) as cnt FROM users WHERE email = ?"
	uid := int64(0)
	err := qr.Get(&uid, str, email)
	if err != nil {
		log.Println("err", err)
		return false
	}
	log.Println("uid", err)
	if uid > 0 {
		return true
	}

	return false

}

// GetUserCount returns number of users not deleted
func GetUserCount(qy db.Queryer) (int64, error) {
	str := "SELECT count(*) from users WHERE deleted = false"
	cnt := int64(0)
	err := qy.Get(&cnt, str)
	return cnt, err
}

// GetUsersByRole returns users with matching role
func GetUsersByRole(qy db.Queryer, roleid int64) ([]User, error) {
	str := `
	SELECT u.user_id, u.first_name, u.last_name, u.email, u.created_on, u.updated_on
	FROM users u
	INNER JOIN user_roles ur
	ON ur.user_id = u.user_id
	WHERE u.deleted = false
	AND ur.role_id = ?
	`
	users := []User{}
	err := qy.Select(&users, str, roleid)
	return users, err
}
