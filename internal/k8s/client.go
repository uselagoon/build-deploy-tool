package k8s

import (
	"context"
	"fmt"
	"os"

	mariadbv1 "github.com/amazeeio/dbaas-operator/apis/mariadb/v1"
	mongodbv1 "github.com/amazeeio/dbaas-operator/apis/mongodb/v1"
	postgresv1 "github.com/amazeeio/dbaas-operator/apis/postgres/v1"
	k8upv1 "github.com/k8up-io/k8up/v2/api/v1"
	k8upv1alpha1 "github.com/vshn/k8up/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	networkv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	client "sigs.k8s.io/controller-runtime/pkg/client"
	ctrlfake "sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/yaml"
)

func setScheme() (*runtime.Scheme, error) {
	k8sScheme := runtime.NewScheme()
	// any custom crds etc need to be added to the fake client so it knows about them
	if err := mariadbv1.AddToScheme(k8sScheme); err != nil {
		return nil, err
	}
	if err := mongodbv1.AddToScheme(k8sScheme); err != nil {
		return nil, err
	}
	if err := postgresv1.AddToScheme(k8sScheme); err != nil {
		return nil, err
	}
	if err := k8upv1.AddToScheme(k8sScheme); err != nil {
		return nil, err
	}
	if err := k8upv1alpha1.AddToScheme(k8sScheme); err != nil {
		return nil, err
	}
	if err := batchv1.AddToScheme(k8sScheme); err != nil {
		return nil, err
	}
	if err := appsv1.AddToScheme(k8sScheme); err != nil {
		return nil, err
	}
	if err := networkv1.AddToScheme(k8sScheme); err != nil {
		return nil, err
	}
	if err := corev1.AddToScheme(k8sScheme); err != nil {
		return nil, err
	}
	return k8sScheme, nil
}

func NewClient() (client.Client, error) {
	// read the serviceaccount deployer token first.
	token, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/token")
	if err != nil {
		// read the legacy deployer token if for some reason the serviceaccount is not found.
		token, err = os.ReadFile("/var/run/secrets/lagoon/deployer/token")
		if err != nil {
			return nil, err
		}
	}
	k8sScheme, err := setScheme()
	if err != nil {
		return nil, err
	}

	// generate the rest config for the client.
	config := &rest.Config{
		BearerToken: string(token),
		Host:        "https://kubernetes.default.svc",
		TLSClientConfig: rest.TLSClientConfig{
			Insecure: true,
		},
	}
	// create the client using the rest config.
	return client.New(config, client.Options{
		Scheme: k8sScheme,
	})
}

func NewFakeClient(namespace string) (client.Client, error) {
	k8sScheme, err := setScheme()
	if err != nil {
		return nil, err
	}
	clientBuilder := ctrlfake.NewClientBuilder()
	clientBuilder = clientBuilder.WithScheme(k8sScheme)

	fakeClient := clientBuilder.Build()
	ns := corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}
	err = fakeClient.Create(context.Background(), &ns)
	if err != nil {
		return nil, err
	}
	return fakeClient, nil
}

func SeedFakeData(fakeClient client.Client, namespace string, seedDir string) error {
	dir, err := os.ReadDir(seedDir)
	if err != nil {
		return fmt.Errorf("couldn't read directory %v: %v", seedDir, err)
	}
	for _, r := range dir {
		if r.Name() == ".gitkeep" {
			continue
		}
		seedFile := fmt.Sprintf("%s/%s", seedDir, r.Name())
		sfb, err := os.ReadFile(seedFile)
		if err != nil {
			return fmt.Errorf("couldn't read file %v: %v", seedFile, err)
		}
		u := &unstructured.Unstructured{}
		yaml.Unmarshal(sfb, u)
		u.SetNamespace(namespace)
		err = fakeClient.Create(context.Background(), u)
		if err != nil {
			return err
		}
	}
	return nil
}
