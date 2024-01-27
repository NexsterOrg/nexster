package rbac

import (
	"fmt"

	"github.com/euroteltr/rbac"
)

type RbacGuard struct {
	cmd    *rbac.RBAC
	Perm   *permission
	Action *action
	Role   *role
}

var _ Interface = (*RbacGuard)(nil)

func NewRbacGuard() *RbacGuard {
	r := rbac.New(nil)

	perm := &permission{}
	action := &action{}
	role := &role{}

	// Define new actions
	action.Accept = rbac.Action("accept")
	action.Reject = rbac.Action("reject")
	action.Suspend = rbac.Action("suspend")
	action.Activate = rbac.Action("activate")

	// mangeBoardingAds permission
	mangeBoardingAds, err := r.RegisterPermission("mangeBoardingAds", "Manage boarding ad resource", rbac.CRUD, action.Accept, action.Reject)
	if err != nil {
		panic(fmt.Errorf("unable to register permission manageBoardingAds: %v", err))
	}
	perm.ManageBoardingAds = mangeBoardingAds

	// mangeBoardingOwner permission
	mangeBoardingOwners, err := r.RegisterPermission("mangeBoardingOwners", "Manage boarding owner resource", rbac.CRUD, action.Accept, action.Reject,
		action.Suspend, action.Activate)
	if err != nil {
		panic(fmt.Errorf("unable to register permission manageBoardingOwners: %v", err))
	}
	perm.ManageBoardingOwners = mangeBoardingOwners

	// Add new permission here.

	// Add roles
	// TODO: Add support to hierachical role support. (eg: admin, subAdmin etc)
	reviewerRole, err := r.RegisterRole(reviewer, "Ad reviewer role")
	if err != nil {
		panic(fmt.Errorf("can not add reviewer role: %v", err))
	}
	bdOwnerRole, err := r.RegisterRole(BdOwner, "Ad bdOwner role")
	if err != nil {
		panic(fmt.Errorf("can not add boarding owner role: %v", err))
	}
	studentRole, err := r.RegisterRole(student, "Student role")
	if err != nil {
		panic(fmt.Errorf("can not add boarding owner role: %v", err))
	}
	// Add new roles here.

	// Granting privileages - reviewer
	if err = r.Permit(reviewerRole.ID, mangeBoardingAds, rbac.Create, rbac.Read, rbac.Update, rbac.Delete,
		action.Accept, action.Reject); err != nil {
		panic(fmt.Errorf("can not permit mangeBoardingAds permissions to role %s", reviewerRole.ID))
	}
	if err = r.Permit(reviewerRole.ID, mangeBoardingOwners, rbac.Create, rbac.Read, rbac.Update, rbac.Delete,
		action.Accept, action.Reject, action.Suspend, action.Activate); err != nil {
		panic(fmt.Errorf("can not permit mangeBoardingOwners permissions to role %s", reviewerRole.ID))
	}

	// Granting privileages - bdOwner
	if err = r.Permit(bdOwnerRole.ID, mangeBoardingAds, rbac.Create, rbac.Read, rbac.Update, rbac.Delete); err != nil {
		panic(fmt.Errorf("can not permit mangeBoardingAds permissions to role %s", bdOwnerRole.ID))
	}
	if err = r.Permit(bdOwnerRole.ID, mangeBoardingOwners, rbac.Create, rbac.Read, rbac.Update, rbac.Delete); err != nil {
		panic(fmt.Errorf("can not permit mangeBoardingOwners permissions to role %s", bdOwnerRole.ID))
	}

	// Granting privileages - student (Need to add other privileages for students)
	// TODO: STUDENTS SHOULD NOT HAVE PRIVILEAGES TO ACCPET OR REJECT ADS. THIS IS ONLY TILL REVIEWER ACCOUNT BECOME AVAILABLE.
	// REMOVE action.Accept, action.Reject in future.
	if err = r.Permit(studentRole.ID, mangeBoardingAds, rbac.Read, action.Accept, action.Reject); err != nil {
		panic(fmt.Errorf("can not permit mangeBoardingAds permissions to role %s", studentRole.ID))
	}
	// Add new grant privileages

	role.reviewer = reviewerRole
	role.bdOwner = bdOwnerRole
	role.student = studentRole

	return &RbacGuard{
		cmd:    r,
		Perm:   perm,
		Action: action,
		Role:   role,
	}
}

func (r *RbacGuard) IsGranted(roleId string, perm *rbac.Permission, actions ...rbac.Action) bool {
	return r.cmd.IsGranted(roleId, perm, actions...)
}

func (r *RbacGuard) HasDesiredRole(roleIds []string, perm *rbac.Permission, actions ...rbac.Action) bool {
	for _, role := range roleIds {
		if r.cmd.IsGranted(role, perm, actions...) {
			return true
		}
	}
	return false
}
