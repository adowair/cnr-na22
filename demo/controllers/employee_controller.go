/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"

	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	cnrna22v1 "github.com/adowair/cnr-na22/api/v1"
)

// EmployeeReconciler reconciles a Employee object
type EmployeeReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

const employeeFinalizerName = "github.com.adowair.cnr-na22/finalizer"

//+kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=cnr-na22.my.domain,resources=employees,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=cnr-na22.my.domain,resources=employees/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=cnr-na22.my.domain,resources=employees/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Employee object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *EmployeeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// Get the employee object on which the request was made.
	var employee cnrna22v1.Employee
	if err := r.Get(ctx, req.NamespacedName, &employee); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// If the delete timestamp is not set, the object was created or updated.
	if employee.ObjectMeta.DeletionTimestamp.IsZero() {
		// Add the employee to their team's roster. Reconcile state.
		if err := r.addEmployeeToRoster(ctx, &employee); err != nil {
			log.Error(err, "error adding employee to roster")
			return ctrl.Result{}, err
		}
		// Add a finalizer to control the employee object's deletion
		if err := r.maybeAddFinalizer(ctx, &employee); err != nil {
			log.Error(err, "error adding finalizer")
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	// Otherwise the object was deleted. Reconcile state.
	if err := r.removeEmployeeFromRoster(ctx, &employee); err != nil {
		log.Error(err, "error removing employee from roster")
		return ctrl.Result{}, err
	}
	// Remove the finalizer to allow deletion to proceed.
	if err := r.removeFinalizer(ctx, &employee); err != nil {
		log.Error(err, "error removing finalizer")
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

// maybeAddFinalizer adds a finalizer to the object if one does not already
// exist. If the object was only updated in this request, it should already
// have a finalizer. Learn more here https://kubernetes.io/docs/concepts/overview/working-with-objects/finalizers/
func (r *EmployeeReconciler) maybeAddFinalizer(ctx context.Context, e *cnrna22v1.Employee) error {
	if !controllerutil.ContainsFinalizer(e, employeeFinalizerName) {
		controllerutil.AddFinalizer(e, employeeFinalizerName)
		return r.Update(ctx, e)
	}
	return nil
}

// maybeRemoveFinalizer removes the finalizer from the employee object. If
// the finalizer is not found, then the object is being deleted already.
func (r *EmployeeReconciler) removeFinalizer(ctx context.Context, e *cnrna22v1.Employee) error {
	change := controllerutil.RemoveFinalizer(e, employeeFinalizerName)
	if change {
		return r.Update(ctx, e)
	}
	return nil
}

// rosterName returns, for a given employee, the name of the roster they should
// should belong to. This is derived from the employee's namespace
func rosterName(e *cnrna22v1.Employee) string {
	// Make team name a valid Kubernetes name
	return fmt.Sprintf("%s-team-roster", e.Namespace)
}

// rosterKey returns, for an employee, the key with which they should be added
// to the roster (a configMap)'s Data.
func rosterKey(e *cnrna22v1.Employee) string {
	// Use the object's Kubernetes name. This is guaranteed to be unique
	// in one namespace.
	return e.Name
}

// getRoster gets the roster object to which this employee would belong and
// copies it into cm if it exists.
func (r *EmployeeReconciler) getRoster(ctx context.Context, e *cnrna22v1.Employee, cm *v1.ConfigMap) error {
	rosterCMKey := client.ObjectKey{
		Namespace: v1.NamespaceDefault,
		Name:      rosterName(e),
	}
	return r.Get(ctx, rosterCMKey, cm)
}

// addEmployeeToRoster manages the lifecycle of the roster object to which an
// employee belongs. This function will add the employee to their team's roster.
// If a roster does not exist for the employee's team, one is created.
func (r *EmployeeReconciler) addEmployeeToRoster(ctx context.Context, e *cnrna22v1.Employee) error {
	var rosterCM v1.ConfigMap
	err := r.getRoster(ctx, e, &rosterCM)

	rosterName := rosterName(e)
	e.Status.Roster = rosterName
	if err := r.Update(ctx, e, &client.UpdateOptions{}); err != nil {
		return err
	}

	switch {
	// Roster exists, just update it.
	case err == nil:
		rosterCM.Data[rosterKey(e)] = e.Spec.Name
		return r.Update(ctx, &rosterCM)
	// Roster does not exist. Add metadata and create it.
	case apierrors.IsNotFound(err):
		rosterCM.ObjectMeta = metav1.ObjectMeta{
			Namespace: v1.NamespaceDefault,
			Name:      rosterName,
		}
		rosterCM.Data = make(map[string]string)
		rosterCM.Data[rosterKey(e)] = e.Spec.Name
		return r.Create(ctx, &rosterCM)
	// Some unexpected error.
	default:
		return err
	}
}

// removeEmployeeFromRoster manages the lifecycle of the roster object to which
// an employee belongs. This function will remove the employee from the roster.
// If the roster is then empty, it is deleted.
func (r *EmployeeReconciler) removeEmployeeFromRoster(ctx context.Context, e *cnrna22v1.Employee) error {
	var rosterCM v1.ConfigMap
	err := r.getRoster(ctx, e, &rosterCM)

	delete(rosterCM.Data, rosterKey(e))
	switch {
	// Roster exists. Update the roster, or delete it if empty.
	case err == nil:
		if len(rosterCM.Data) == 0 {
			return r.Delete(ctx, &rosterCM)
		}
		return r.Update(ctx, &rosterCM)
	// Roster already deleted. Stop.
	case apierrors.IsNotFound(err):
		return nil
	// Some unexpected error.
	default:
		return err
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *EmployeeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cnrna22v1.Employee{}).
		Complete(r)
}
