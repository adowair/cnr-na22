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
	"strings"

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

	// Get the employee object on which a request was made.
	var employee cnrna22v1.Employee
	if err := r.Get(ctx, req.NamespacedName, &employee); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// If employee object is being created or updated, the deleted timestamp
	// should not be set.
	if employee.ObjectMeta.DeletionTimestamp.IsZero() {
		if err := r.addEmployeeToRoster(ctx, &employee); err != nil {
			log.Error(err, "error adding employee to roster")
			return ctrl.Result{}, err
		}
		if err := r.maybeAddFinalizer(ctx, &employee); err != nil {
			log.Error(err, "error adding finalizer")
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	// Otherwise the object is being deleted.
	if err := r.removeEmployeeFromRoster(ctx, &employee); err != nil {
		log.Error(err, "error removing employee from roster")
		return ctrl.Result{}, err
	}
	if err := r.maybeRemoveFinalizer(ctx, &employee); err != nil {
		log.Error(err, "error removing finalizer")
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func (r *EmployeeReconciler) maybeAddFinalizer(ctx context.Context, e *cnrna22v1.Employee) error {
	if !controllerutil.ContainsFinalizer(e, employeeFinalizerName) {
		controllerutil.AddFinalizer(e, employeeFinalizerName)
		return r.Update(ctx, e)
	}
	return nil
}

func (r *EmployeeReconciler) maybeRemoveFinalizer(ctx context.Context, e *cnrna22v1.Employee) error {
	if controllerutil.ContainsFinalizer(e, employeeFinalizerName) {
		controllerutil.RemoveFinalizer(e, employeeFinalizerName)
		return r.Update(ctx, e)
	}
	return nil
}

func rosterName(e *cnrna22v1.Employee) string {
	rosterCMName := fmt.Sprintf("team-roster-%s", strings.ReplaceAll(e.Spec.TeamName, " ", "-"))
	return strings.ToLower(rosterCMName)
}

func rosterKey(e *cnrna22v1.Employee) string {
	return e.Name
}

func (r *EmployeeReconciler) getRoster(ctx context.Context, e *cnrna22v1.Employee, cm *v1.ConfigMap) error {
	rosterCMKey := client.ObjectKey{
		Namespace: e.Namespace,
		Name:      rosterName(e),
	}
	return r.Get(ctx, rosterCMKey, cm)
}

func (r *EmployeeReconciler) addEmployeeToRoster(ctx context.Context, e *cnrna22v1.Employee) error {
	var rosterCM v1.ConfigMap
	err := r.getRoster(ctx, e, &rosterCM)

	switch {
	// Roster exists, just update it.
	case err == nil:
		rosterCM.Data[rosterKey(e)] = e.Spec.Name
		return r.Update(ctx, &rosterCM)
	// Roster does not exist. Add metadata and create it.
	case apierrors.IsNotFound(err):
		rosterCM.ObjectMeta = metav1.ObjectMeta{
			Namespace: e.Namespace,
			Name:      rosterName(e),
		}
		rosterCM.Data = make(map[string]string)
		rosterCM.Data[rosterKey(e)] = e.Spec.Name
		return r.Create(ctx, &rosterCM)
	// Some unexpected error.
	default:
		return err
	}
}

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
