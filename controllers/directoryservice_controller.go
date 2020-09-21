/*


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

	directoryv1alpha1 "github.com/ForgeRock/ds-operator/api/v1alpha1"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// DirectoryServiceReconciler reconciles a DirectoryService object
type DirectoryServiceReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// Reconcile loop for DS controller
// Add in all the RBAC permissions that a DS controller needs. StatefulSets, etc.
// +kubebuilder:rbac:groups=directory.forgerock.com,resources=directoryservices,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=directory.forgerock.com,resources=directoryservices/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=batch,resources=jobs/status,verbs=get
// +kubebuilder:rbac:groups=apps,resources=statefulsets/status,verbs=get
// +kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;create;update;patch;delete
func (r *DirectoryServiceReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	var log = r.Log.WithValues("directoryservice", req.NamespacedName)

	log.Info("Started")

	// your logic here

	return ctrl.Result{}, nil
}

// SetupWithManager stuff
func (r *DirectoryServiceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&directoryv1alpha1.DirectoryService{}).
		Complete(r)
}
