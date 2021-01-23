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

package v1alpha1

import (
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var directoryservicelog = logf.Log.WithName("directoryservice-resource")

// SetupWebhookWithManager registers the webhook with the manager
func (r *DirectoryService) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// +kubebuilder:webhook:path=/mutate-directory-forgerock-com-forgerock-com-v1alpha1-directoryservice,mutating=true,failurePolicy=fail,groups=directory.forgerock.io.forgerock.io,resources=directoryservices,verbs=create;update,versions=v1alpha1,name=mdirectoryservice.kb.io

var _ webhook.Defaulter = &DirectoryService{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *DirectoryService) Default() {
	directoryservicelog.Info("default", "name", r.Name)

	// TODO(user): fill in your defaulting logic.

	// If no replicas specified - default to 1
	if r.Spec.Replicas == nil {
		r.Spec.Replicas = new(int32)
		*r.Spec.Replicas = 1
	}
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
// +kubebuilder:webhook:verbs=create;update,path=/validate-directory-forgerock-com-forgerock-com-v1alpha1-directoryservice,mutating=false,failurePolicy=fail,groups=directory.forgerock.io.forgerock.io,resources=directoryservices,versions=v1alpha1,name=vdirectoryservice.kb.io

var _ webhook.Validator = &DirectoryService{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *DirectoryService) ValidateCreate() error {
	directoryservicelog.Info("validate create", "name", r.Name)
	var allErrs field.ErrorList

	// TODO(user): fill in your validation logic upon object creation.
	if *r.Spec.Replicas > 6 {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("Replicas"), *r.Spec.Replicas, "must not exceed 6"))
	}

	if len(allErrs) == 0 {
		return nil
	}

	return apierrors.NewInvalid(
		schema.GroupKind{Group: "directory.forgerock.io", Kind: "DirectoryService"},
		r.Name, allErrs)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *DirectoryService) ValidateUpdate(old runtime.Object) error {
	directoryservicelog.Info("validate update", "name", r.Name)

	// TODO(user): fill in your validation logic upon object update.
	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *DirectoryService) ValidateDelete() error {
	directoryservicelog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}
