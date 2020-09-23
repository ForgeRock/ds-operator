/*
   skeleton DS controller
*/

package controllers

import (
	"context"

	"github.com/pkg/errors"

	directoryv1alpha1 "github.com/ForgeRock/ds-operator/api/v1alpha1"
	"github.com/go-logr/logr"
	apps "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
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
	ctx := context.Background()
	var log = r.Log.WithValues("directoryservice", req.NamespacedName)

	log.Info("Started")

	var ds directoryv1alpha1.DirectoryService

	// Load the DirectoryService
	if err := r.Get(ctx, req.NamespacedName, &ds); err != nil {
		log.Info("unable to fetch DirectorService. You can probably ignore this..")
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// finalizer hooks..
	// This registers finalizers for deleting the object
	myFinalizerName := "directory.finalizers.forgerock.com"

	// examine DeletionTimestamp to determine if object is under deletion
	if ds.ObjectMeta.DeletionTimestamp.IsZero() {
		log.Info("Registering finalizer for Directory Service", "name", ds.Name)
		// The object is not being deleted, so if it does not have our finalizer,
		// then lets add the finalizer and update the object. This is equivalent
		// registering our finalizer.
		if !containsString(ds.GetFinalizers(), myFinalizerName) {
			ds.SetFinalizers(append(ds.GetFinalizers(), myFinalizerName))
			if err := r.Update(context.Background(), &ds); err != nil {
				return ctrl.Result{}, err
			}
		}
	} else {
		log.Info("Deleting Directory Service", "name", ds.Name)
		// The object is being deleted
		if containsString(ds.GetFinalizers(), myFinalizerName) {
			// our finalizer is present, so lets handle any external dependency
			if err := r.deleteExternalResources(&ds); err != nil {
				// if fail to delete the external dependency here, return with error
				// so that it can be retried
				return ctrl.Result{}, err
			}

			// remove our finalizer from the list and update it.
			ds.SetFinalizers(removeString(ds.GetFinalizers(), myFinalizerName))
			if err := r.Update(context.Background(), &ds); err != nil {
				return ctrl.Result{}, err
			}
		}

		// Stop reconciliation as the item is being deleted
		return ctrl.Result{}, nil
	}

	// // todo: create a sts...
	// // Move this into a separate function..
	// sts, _ := createDSStatefulSet(&ds)

	// TOOD: Crud method. NOTE: From https://engineering.pivotal.io/post/gp4k-kubebuilder-lessons/
	// In your mutate callback, you should surgically modify individual fields of the object. Don’t overwrite
	// large chunks of the object, or the whole object, as we tried to do initially.

	// This creates a stub sts with only the name/namespace set.
	// The CreateOrUpdate Method will then take this and fill it in the actual values (if the sts exists already)
	var sts apps.StatefulSet
	sts.Name = ds.Name
	sts.Namespace = ds.Namespace

	_, err := ctrl.CreateOrUpdate(ctx, r, &sts, func() error {
		// todo:
		// Fill in STS fields. If the object already exists this should only update fields.
		// ModifyStatefulSet(ds,&sts)
		log.Info("CreateorUpdate statefulset", "sts", sts)

		var err error
		// does the sts not exist yet? Is this the right check?
		if sts.CreationTimestamp.IsZero() {
			err = createDSStatefulSet(&ds, &sts)
			_ = controllerutil.SetControllerReference(&ds, &sts, r.Scheme)
			log.Info("Created New sts from template", "sts", sts)
		} else {
			// If the sts exists already - we want to update any fields to bring its state into
			// alignment with the Custom Resource
			err = updateDSStatefulSet(&ds, &sts)
		}

		log.Info("sts after update/create", "sts", sts)
		return err

	})
	if err != nil {
		return ctrl.Result{}, errors.Wrap(err, "unable to CreateOrUpdate StateFulSet")
	}

	// create or update the service
	var svc v1.Service
	svc.Name = ds.Name
	svc.Namespace = ds.Namespace

	_, err = ctrl.CreateOrUpdate(ctx, r, &svc, func() error {
		log.Info("CreateorUpdate service", "svc", svc)

		var err error
		// does the service not exist yet?
		if svc.CreationTimestamp.IsZero() {
			err = createService(&ds, &svc)
			log.Info("Setting ownerref for service", "svc", svc.Name)
			_ = controllerutil.SetControllerReference(&ds, &svc, r.Scheme)
			log.Info("Created New sts from template", "sts", sts)
		} else {
			// If the sts exists already - we want to update any fields to bring its state into
			// alignment with the Custom Resource
			//err = updateService(&ds, &sts)
			log.Info("TODO: Handle update of ds service")
		}

		log.Info("svc after update/create", "svc", svc)
		return err
	})

	if err != nil {
		return ctrl.Result{}, errors.Wrap(err, "unable to CreateOrUpdate Service")
	}

	return ctrl.Result{}, nil
}

// SetupWithManager stuff
func (r *DirectoryServiceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&directoryv1alpha1.DirectoryService{}).
		Complete(r)
}

// https://godoc.org/k8s.io/api/apps/v1#StatefulSetSpec
func createDSStatefulSet(ds *directoryv1alpha1.DirectoryService, sts *apps.StatefulSet) error {

	// TODO: What is the canonical go way of using these contants in a template. Go wants a pointer to these
	// not a constant
	var fsGroup int64 = 0
	var forgerockUser int64 = 11111
	// todo: update proper tag
	var image = "gcr.io/engineering-devops/ds-idrepo:master-ready-for-dev-pipelines-232-gc547ff85e-dirty"

	// Create a template
	stemplate := &apps.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Labels:      make(map[string]string),
			Annotations: make(map[string]string),
			Name:        ds.Name,
			Namespace:   ds.Namespace,
		},
		Spec: apps.StatefulSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": ds.Name,
				},
			},
			// ServiceName: cr.ObjectMeta.Name,
			Replicas: ds.Spec.Replicas,
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": ds.Name,
					},
				},
				Spec: v1.PodSpec{
					InitContainers: []v1.Container{
						{
							Name:            "init",
							Image:           image,
							ImagePullPolicy: v1.PullAlways, // todo: for testing this is good. Remove later?
							//Command: []string{"/opt/opendj/scripts/init-and-restore.sh"},
							Args: []string{"initialize-only"},
							VolumeMounts: []v1.VolumeMount{
								{
									Name:      "data",
									MountPath: "/opt/opendj/data",
								},
								{
									Name:      "secrets",
									MountPath: "/opt/opendj/secrets",
								},
								// TODO: Why does our kustomize sample mount the same secrets in two locations
								// {
								// 	Name:      "secrets",
								// 	MountPath: "/var/run/secrets/opendj",
								// },
								{
									Name:      "passwords",
									MountPath: "/var/run/secrets/opendj-passwords",
								},
							},
							Env: []v1.EnvVar{
								{
									Name:  "DS_SET_UID_ADMIN_AND_MONITOR_PASSWORDS",
									Value: "true",
								},
								{
									Name:  "DS_UID_MONITOR_PASSWORD_FILE",
									Value: "/var/run/secrets/opendj-passwords/monitor.pw",
								},
								{
									Name:  "DS_UID_ADMIN_PASSWORD_FILE",
									Value: "/var/run/secrets/opendj-passwords/dirmanager.pw",
								},
							},
						},
					},
					Containers: []v1.Container{
						{
							Name:  "ds",
							Image: image,
							// ImagePullPolicy: ImagePullPolicy,
							Args: []string{"start-ds"},
							VolumeMounts: []v1.VolumeMount{
								{
									Name:      "data",
									MountPath: "/opt/opendj/data",
								},
								{
									Name:      "secrets",
									MountPath: "/opt/opendj/secrets",
								},
							},
						},
					},
					SecurityContext: &v1.PodSecurityContext{
						FSGroup:   &fsGroup,
						RunAsUser: &forgerockUser,
					},
					Volumes: []v1.Volume{
						{
							Name: "secrets",
							VolumeSource: v1.VolumeSource{
								Secret: &v1.SecretVolumeSource{
									SecretName: "ds", // NOTE: Do we want common secrets for all instances in the NS?
								},
							},
						},
						{
							Name: "passwords",
							VolumeSource: v1.VolumeSource{
								Secret: &v1.SecretVolumeSource{
									SecretName: "ds-passwords",
								},
							},
						},
					},
				},
			},
			VolumeClaimTemplates: []v1.PersistentVolumeClaim{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "data",
						Namespace: ds.Namespace,
						Annotations: map[string]string{
							"pv.beta.kubernetes.io/gid": "0",
						},
					},
					Spec: v1.PersistentVolumeClaimSpec{
						AccessModes: []v1.PersistentVolumeAccessMode{v1.ReadWriteOnce},
						Resources: v1.ResourceRequirements{
							Requests: v1.ResourceList{
								v1.ResourceName(v1.ResourceStorage): resource.MustParse("10Gi"),
							},
						},
					},
				},
			},
		},
	}
	stemplate.DeepCopyInto(sts)
	return nil
}

// Create the service for ds
func createService(ds *directoryv1alpha1.DirectoryService, svc *v1.Service) error {
	svcTemplate := v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Labels:      make(map[string]string),
			Annotations: make(map[string]string),
			Name:        ds.Name,
			Namespace:   ds.Namespace,
		},
		Spec: v1.ServiceSpec{
			ClusterIP: "None", // headless service
			Selector: map[string]string{
				"app": ds.Name,
			},
			Ports: []v1.ServicePort{
				{
					Name: "admin",
					Port: 4444,
				},
				{
					Name: "ldap",
					Port: 1389,
				},
				{
					Name: "ldaps",
					Port: 1636,
				},
				{
					Name: "http",
					Port: 8080,
				},
			},
		},
	}

	svcTemplate.DeepCopyInto(svc)
	return nil
}

// This function updates an existing statefulset to match settings in the custom resource
// TODO: What kinds of things should we update?
func updateDSStatefulSet(ds *directoryv1alpha1.DirectoryService, sts *apps.StatefulSet) error {

	sts.Spec.Replicas = ds.Spec.Replicas
	return nil
}

func (r *DirectoryServiceReconciler) deleteExternalResources(ds *directoryv1alpha1.DirectoryService) error {
	//
	// delete any external resources associated with the ds set
	//
	// Ensure that delete implementation is idempotent and safe to invoke
	// multiple times for same object.
	return nil
}

// Helper functions to check and remove string from a slice of strings.
func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func removeString(slice []string, s string) (result []string) {
	for _, item := range slice {
		if item == s {
			continue
		}
		result = append(result, item)
	}
	return
}
