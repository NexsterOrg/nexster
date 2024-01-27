package rbac

import (
	"github.com/euroteltr/rbac"
)

// roles
const Reviewer string = "reviewer"
const BdOwner string = "bdOwner"
const student string = "student"

type Interface interface {
	IsGranted(roleId string, perm *rbac.Permission, actions ...rbac.Action) bool
	HasPrivileagesForDesiredRoles(roleIds []string, perm *rbac.Permission, actions ...rbac.Action) bool
	HasDesiredRole(role string, roleIds []string) bool
}

type permission struct {
	ManageBoardingAds    *rbac.Permission
	ManageBoardingOwners *rbac.Permission
}

type action struct {
	Accept   rbac.Action
	Reject   rbac.Action
	Suspend  rbac.Action
	Activate rbac.Action
}

type role struct {
	reviewer *rbac.Role
	bdOwner  *rbac.Role
	student  *rbac.Role
}
