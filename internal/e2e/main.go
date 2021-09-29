package main

import (

	// Auth
	"bytes"
	"context"
	b64 "encoding/base64"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ForgeRock/ds-operator/pkg/ldap"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/portforward"
	watchtools "k8s.io/client-go/tools/watch"
	"k8s.io/client-go/util/homedir"
)

var (
	podRes = schema.GroupVersionResource{
		Group:    "",
		Version:  "v1",
		Resource: "pods",
	}
	sacRes = schema.GroupVersionResource{
		Group:    "secretagentconfigurations",
		Version:  "v1alpha1",
		Resource: "secret-agent.secrets.forgerock.io",
	}
	dsRes = schema.GroupVersionResource{
		Group:    "directory.forgerock.io",
		Version:  "v1alpha1",
		Resource: "directoryservices",
	}
	secretRes = schema.GroupVersionResource{
		Group:    "",
		Version:  "v1",
		Resource: "secrets",
	}

	// ErrWatchTimeout timeout occured while waiting
	ErrWatchTimeout error = errors.New("wait timeout")

	dsServiceBase = `
apiVersion: directory.forgerock.io/v1alpha1
kind: DirectoryService
metadata:
  name: ds-idrepo
  labels:
    app.kubernetes.io/name: ds
    app.kubernetes.io/testcase: ds-nosnapshot
    app.kubernetes.io/part-of: forgerock
spec:
  image: gcr.io/engineering-devops/ds-idrepo:6831d99-dirty@sha256:45509bac0c5e97796337cd997d878baa9a8190a8ce9b4bbe972e2cbdb6320be0
  replicas: 1
  volumeClaimSpec:
    storageClassName: fast
    accessModes: [ "ReadWriteOnce" ]
    resources:
      requests:
        storage: 2Gi
  snapshots:
    enabled: false
    periodMinutes: 2
    snapshotsRetained: 2
    volumeSnapshotClassName: ds-snapshot-class
  passwords:
    uid=admin:
      secretName: ds-passwords
      key: dirmanager.pw
    uid=monitor:
      secretName: ds-passwords
      key: monitor.pw
    uid=openam_cts,ou=admins,ou=famrecords,ou=openam-session,ou=tokens:
      secretName: ds-env-secrets
      key: AM_STORES_CTS_PASSWORD
    uid=am-identity-bind-account,ou=admins,ou=identities:
      secretName: ds-env-secrets
      key: AM_STORES_USER_PASSWORD
    uid=am-config,ou=admins,ou=am-config:
      secretName: ds-env-secrets
      key: AM_STORES_APPLICATION_PASSWORD
  keystore:
    secretName: ds
  truststore:
    secretName: "platform-ca"
    keyName: "ca.pem"
`
)

// port forwarding needs a thread safe writer
// we want to get this logged just like everything else
// set up a io.Writer interface
type logBufferLevel string

const (
	logBufferError logBufferLevel = "Error"
	logBufferInfo  logBufferLevel = "Info"
)

type SafeBufferedWrite struct {
	logger *zap.Logger
	level  logBufferLevel
}

func (s *SafeBufferedWrite) Write(p []byte) (n int, err error) {
	switch s.level {
	case logBufferError:
		s.logger.Error("port-forward", zap.Binary("error", p))
		return 1, nil
	case logBufferInfo:
		s.logger.Info("port-forward", zap.Binary("info", p))
		return 1, nil
	}
	return 0, nil
}

func isPodReady(event watch.Event) (bool, error) {
	if event.Type == watch.Error {
		err := apierrors.FromObject(event.Object)
		return false, err
	}
	obj := event.Object.(*unstructured.Unstructured)
	phase, _, err := unstructured.NestedString(obj.Object, "status", "phase")
	if phase == "Running" {
		return true, nil
	} else if err != nil {
		return false, err
	} else {
		return false, nil
	}
}

func isDSReady(event watch.Event) (bool, error) {
	if event.Type == watch.Error {
		err := apierrors.FromObject(event.Object)
		return false, err
	}
	obj := event.Object.(*unstructured.Unstructured)
	phase, _, err := unstructured.NestedInt64(obj.Object, "status", "serviceAccountPasswordsUpdatedTime")
	if phase > 0 {
		return true, nil
	} else if err != nil {
		return false, err
	} else {
		return false, nil
	}

}

// K8Resource is something that we want to monitor in kubernetes
type K8Resource struct {
	client dynamic.ResourceInterface
	name   string
}

// Wait wait for a resource to become ready
func Wait(ctx context.Context, resource *K8Resource, conditionFn watchtools.ConditionFunc) (bool, error) {
	nameSelector := fields.OneTermEqualSelector("metadata.name", resource.name).String()
	watchOptions := metav1.ListOptions{}
	watchOptions.FieldSelector = nameSelector
	for {
		gottenObjList, err := resource.client.List(ctx, metav1.ListOptions{FieldSelector: nameSelector})
		if apierrors.IsNotFound(err) {
			gottenObjList = &unstructured.UnstructuredList{}
		} else if err != nil && !apierrors.IsNotFound(err) {
			return false, err
		}
		// If the object is present, let's evaluate if the condition has already been met.
		if len(gottenObjList.Items) != 0 {
			// Make an event to use in the condition func, hopefully we don't
			// need to worry about the event type at the moment.
			event := watch.Event{
				"ADDED",
				gottenObjList.Items[0].DeepCopyObject(),
			}
			met, err := conditionFn(event)
			if err != nil {
				return false, err

			} else if met {
				return true, nil
			}

		}
		// The condition has not been met. Set a watch on the object
		watchOptions.ResourceVersion = gottenObjList.GetResourceVersion()
		objWatch, err := resource.client.Watch(ctx, watchOptions)
		if err != nil {
			return false, err
		}
		lastEvent, err := watchtools.UntilWithoutRetry(ctx, objWatch, conditionFn)
		if err == watchtools.ErrWatchClosed {
			continue
		}
		if err != nil || lastEvent == nil {
			return false, err
		}
		return true, nil
	}
}

func main() {

	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	namespace := "max"

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err)
	}
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		panic(err)
	}
	cmd := exec.Command("kubectl", "apply", "-n", namespace, "-f", "-")
	cmd.Stdin = strings.NewReader(dsServiceBase)
	var out bytes.Buffer
	cmd.Stdout = &out
	err = cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("DS Deployment started")

	dsPod := &K8Resource{
		dynamicClient.Resource(podRes).Namespace(namespace),
		"ds-idrepo-0",
	}

	// TODO context needs to be set for ENTIRE test
	// then should be a sub context per task with a timeout
	ctx := context.Background()
	// Check pods first, sometimes they don't come up etc, no point in checking
	// on a password set status if the pods don't actually run
	podReady, err := Wait(ctx, dsPod, isPodReady)
	if err != nil {
		log.Fatal(err)
	}
	if podReady {
		fmt.Println("Pod ready")
	} else {
		fmt.Println("Pod not ready anytime")
	}
	dsCRD := &K8Resource{
		dynamicClient.Resource(dsRes).Namespace(namespace),
		"ds-idrepo",
	}
	dsReady, err := Wait(ctx, dsCRD, isDSReady)
	if err != nil {
		log.Fatal(err)
	}
	if dsReady {
		fmt.Println("DS Instance Password set")
	} else {
		fmt.Println("DS instance password wasn't set")
	}
	fmt.Println("DS READY TO RAGE")

	fmt.Println("Creating DS Record")
	// get DS secret
	secClient := dynamicClient.Resource(secretRes).Namespace(namespace)
	result, err := secClient.Get(ctx, "ds-passwords", metav1.GetOptions{})
	if err != nil {
		log.Fatal(err.Error())
	}

	dsAdmin := ""
	passwordB64Encoded, found, err := unstructured.NestedString(result.Object, "data", "dirmanager.pw")
	if err != nil {
		log.Fatal(err.Error())
	} else if found {
		value, err := b64.StdEncoding.DecodeString(passwordB64Encoded)
		if err != nil {
			log.Fatal(err.Error())
		}
		dsAdmin = string(value)
	} else {
		log.Fatal("DS Admin Password Not Found")
	}
	_ = &ldap.DSConnection{
		DN:       "uid=admin",
		Password: dsAdmin,
		URL:      "ldaps://localhost:1636",
	}
	transport, upgrader, err := spdy.RouteTripperFor(config)
	if err != nil {
		log.Fatal(err.Error())
	}
	restclient, err := rest.RESTClientFor(config)
	if err != nil {
		log.Fatal(err.Error())
	}
	req := restclient.Post().Resource("pods").Namespace(namespace).Name("ds-idrepo-0").SubResource("portfoward")

	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, method, req.Url())
	// dialer https://github.com/kubernetes/kubectl/blob/0f88fc6b598b7e883a391a477215afb080ec7733/pkg/cmd/portforward/portforward.go#L133

	logger, err := zap.NewProduction()
	if err != nil {
		fmt.Println(err.Error())
	}
	infoWriter := SafeBufferedWrite{logger: logger, level: logBufferInfo}
	errWriter := SafeBufferedWrite{logger: logger, level: logBufferError}
	stopChannel := make(chan struct{}, 1)
	readyChannel := make(chan struct{})
	fw, err := portforward.NewOnAddresses(dialer, "localhost", 1636, stopChannel, readyChannel, infoWriter, errWriter)
	f

	// restconfig for url https://pkg.go.dev/k8s.io/client-go@v0.22.2/rest#RESTClientFor
	fmt.Println(dsAdmin)
}
