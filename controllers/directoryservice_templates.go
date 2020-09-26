/*
  Templates for creating object instances
*/

package controllers

import (
	"math/rand"
	"time"

	directoryv1alpha1 "github.com/ForgeRock/ds-operator/api/v1alpha1"
	apps "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

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
									SecretName: ds.Spec.SecretReferencePasswords,
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
	return nil // todo: can this ever fail?
}

// Generate a new secret for the admin passwords
// Sets the passwords to random strings
func createAdminSecret(ds *directoryv1alpha1.DirectoryService, secret *v1.Secret) error {
	secretTemplate := v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Labels:      make(map[string]string),
			Annotations: make(map[string]string),
			Name:        ds.Spec.SecretReferencePasswords,
			Namespace:   ds.Namespace,
		},
		Data: map[string][]byte{
			"dirmanager.pw": []byte(randPassword(24)),
			"monitor.pw":    []byte(randPassword(15)),
		},
	}
	secretTemplate.DeepCopyInto(secret)

	return nil
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!$^#()-+<>")

func init() {
	rand.Seed(time.Now().UnixNano())
}

func randPassword(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
