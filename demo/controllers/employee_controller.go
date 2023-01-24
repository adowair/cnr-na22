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
	"sigs.k8s.io/controller-runtime/pkg/log"

	cnrna22v1 "github.com/adowair/cnr-na22/api/v1"
)

// EmployeeReconciler reconciles a Employee object
type EmployeeReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

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

	var employee cnrna22v1.Employee
	if err := r.Get(ctx, req.NamespacedName, &employee); err != nil {
		log.Error(err, "unable to get employee from request")
		return ctrl.Result{}, err
	}

	rosterCMName := fmt.Sprintf("team-roster-%s", strings.ReplaceAll(employee.Spec.TeamName, " ", "-"))
	rosterCMName = strings.ToLower(rosterCMName)

	employeeKey := fmt.Sprint(employee.Spec.ID)
	var rosterCM v1.ConfigMap
	rosterCMKey := client.ObjectKey{
		Namespace: req.Namespace,
		Name:      rosterCMName,
	}
	err := r.Get(ctx, rosterCMKey, &rosterCM)

	if employee.ObjectMeta.DeletionTimestamp.IsZero() {
		// Then the object is being created
		switch {
		case apierrors.IsNotFound(err):
			// Then roster does not yet exist. Create it.
			rosterCM = v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      rosterCMKey.Name,
					Namespace: rosterCMKey.Namespace,
				},
				Data: map[string]string{
					employeeKey: employee.Name,
				},
			}
			return ctrl.Result{}, r.Create(ctx, &rosterCM)
		case err == nil:
			// Then roster exists. Update it.
			rosterCM.Data[employeeKey] = employee.Name
			return ctrl.Result{}, r.Update(ctx, &rosterCM)
		}
	} else {
		// Then the object is being deleted
		switch {
		case apierrors.IsNotFound(err):
			// Then the roster is already deleted. Done reconciling.
			return ctrl.Result{}, nil
		case err == nil:
			// The roster exists.
			delete(rosterCM.Data, employeeKey)
			if len(rosterCM.Data) == 0 {
				// The roster is now empty. Delete it.
				err = r.Delete(ctx, &rosterCM)
			} else {
				err = r.Update(ctx, &rosterCM)
			}
			return ctrl.Result{}, err
		}
	}

	log.Error(err, "error getting roster object")
	return ctrl.Result{}, err
}

// SetupWithManager sets up the controller with the Manager.
func (r *EmployeeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cnrna22v1.Employee{}).
		Complete(r)
}
