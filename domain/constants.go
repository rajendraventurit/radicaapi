package domain

import "fmt"

// templates
const defTemplatePath = "/etc/radica/templates"

// invites
const (
	inviteKey      = "freak-lord-angles"
	inviteExpHours = 48
)

// password reset
const resetKey = "beatle-franklin-desk"

// ErrDuplicateName is a duplicate name error
var ErrDuplicateName = fmt.Errorf("Duplicate name")

// ErrDuplicateEmail is a duplicate email error
var ErrDuplicateEmail = fmt.Errorf("User is already accepted")

// User Roles
const (
	RoleUser      = 1
	RoleAdmin     = 2
	RoleSuperUser = 3
)

// Org Type
const (
	OrgOrganization = 1
	OrgLocation     = 2
	OrgDepartment   = 3
)

// Permissions
const (
	PermManageOrg   = 1
	PermManageUsers = 2
	PermManageSelf  = 3
)

// UserStatus is a user's status
type UserStatus int

// User statuses
const (
	StatusInvited UserStatus = iota
	StatusActive
	StatusDeleted
)

func (s UserStatus) String() string {
	switch s {
	case StatusInvited:
		return "Invited"
	case StatusActive:
		return "Active"
	case StatusDeleted:
		return "Deleted"
	}
	return "Unknown"
}

//password related constant
const minPassLength = 8
const maxPassLength = 64
