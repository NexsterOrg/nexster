package rbac

import (
	"github.com/euroteltr/rbac"
)

const reviewer string = "reviewer"
const bdOwner string = "bdOwner"

type Interface interface {
	IsGranted(roleId string, perm *rbac.Permission, actions ...rbac.Action) bool
	HasDesiredRole(roleIds []string, perm *rbac.Permission, actions ...rbac.Action) bool
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
}
