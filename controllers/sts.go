/*
  Templates for creating object instances
*/

package controllers

import (
	"context"

	directoryv1alpha1 "github.com/ForgeRock/ds-operator/api/v1alpha1"
	"github.com/pkg/errors"
	apps "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func (r *DirectoryServiceReconciler) reconcileSTS(ctx context.Context, ds *directoryv1alpha1.DirectoryService) (ctrl.Result, error) {
	var sts apps.StatefulSet
	sts.Name = ds.Name
	sts.Namespace = ds.Namespace

	_, err := ctrl.CreateOrUpdate(ctx, r, &sts, func() error {
		// todo:
		// Fill in STS fields. If the object already exists this should only update fields.
		// ModifyStatefulSet(ds,&sts)
		r.Log.V(8).Info("CreateorUpdate statefulset", "sts", sts)

		var err error
		// does the sts not exist yet? Is this the right check?
		if sts.CreationTimestamp.IsZero() {
			err = createDSStatefulSet(ds, &sts)
			_ = controllerutil.SetControllerReference(ds, &sts, r.Scheme)
			r.Log.V(8).Info("Created New sts from template", "sts", sts)
		} else {
			// If the sts exists already - we want to update any fields to bring its state into
			// alignment with the Custom Resource
			err = updateDSStatefulSet(ds, &sts)
		}

		r.Log.V(8).Info("sts after update/create", "sts", sts)
		return err

	})
	if err != nil {
		return ctrl.Result{}, errors.Wrap(err, "unable to CreateOrUpdate StateFulSet")
	}
	return ctrl.Result{}, nil
}

// This function updates an existing statefulset to match settings in the custom resource
// TODO: What kinds of things should we update?
func updateDSStatefulSet(ds *directoryv1alpha1.DirectoryService, sts *apps.StatefulSet) error {

	// Copy our expected replicas to the statefulset
	sts.Spec.Replicas = ds.Spec.Replicas

	// copy the current sts replicas up the ds status
	ds.Status.CurrentReplicas = &sts.Status.CurrentReplicas

	return nil
}

// https://godoc.org/k8s.io/api/apps/v1#StatefulSetSpec
func createDSStatefulSet(ds *directoryv1alpha1.DirectoryService, sts *apps.StatefulSet) error {

	// TODO: What is the canonical go way of using these contants in a template. Go wants a pointer to these
	// not a constant
	var fsGroup int64 = 0
	var forgerockUser int64 = 11111

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
						"app":      ds.Name,
						"affinity": "directory", // for anti-affinity
					},
				},
				Spec: v1.PodSpec{
					// We use anti affinity to spread the pods out over host node

					Affinity: &v1.Affinity{
						// NodeAffinity:    &v1.NodeAffinity{},
						// PodAffinity:     &v1.PodAffinity{},
						PodAntiAffinity: &v1.PodAntiAffinity{
							//RequiredDuringSchedulingIgnoredDuringExecution:  []v1.PodAffinityTerm{},
							PreferredDuringSchedulingIgnoredDuringExecution: []v1.WeightedPodAffinityTerm{
								{
									Weight: 100,
									PodAffinityTerm: v1.PodAffinityTerm{
										LabelSelector: &metav1.LabelSelector{
											MatchLabels: map[string]string{},
											MatchExpressions: []metav1.LabelSelectorRequirement{
												{
													Key:      "affinity",
													Operator: "In",
													Values:   []string{"directory"},
												},
											},
										},
										TopologyKey: "kubernetes.io/hostname",
									},
								},
							},
						},
					},
					// Tolerate any nodes tainted with kubectl taint nodes node1 key=directory:NoSchedule
					Tolerations: []v1.Toleration{
						{
							Key:      "key",
							Operator: "Equal",
							Value:    "directory",
							Effect:   "NoSchedule",
						},
					},
					InitContainers: []v1.Container{
						{
							Name:            "init",
							Image:           ds.Spec.Image,
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
								{
									Name:      "passwords",
									MountPath: "/var/run/secrets/opendj-passwords",
								},
							},
							// TODO: Do we just hard code resource requirements for the init container? Or copy the main container
							Resources: v1.ResourceRequirements{
								Limits: v1.ResourceList{
									"memory": resource.MustParse("1024Mi"),
								},
								Requests: v1.ResourceList{
									"cpu":    resource.MustParse("250m"),
									"memory": resource.MustParse("1024Mi"),
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
							Image: ds.Spec.Image,
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
							Resources: ds.DeepCopy().Spec.Resources,
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
									SecretName: ds.Spec.Keystores.KeyStoreSecretName,
								},
							},
						},
						{
							Name: "passwords",
							VolumeSource: v1.VolumeSource{
								Secret: &v1.SecretVolumeSource{
									// todo
									//SecretName: ds.Spec.SecretReferencePasswords,
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
