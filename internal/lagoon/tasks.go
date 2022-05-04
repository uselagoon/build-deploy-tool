package lagoon

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/uselagoon/build-deploy-tool/internal/helpers"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"
	"time"
)

type Task struct {
	Name      string `json:"name"`
	Command   string `json:"command"`
	Namespace string `json:"namespace"`
	Service   string `json:"service"`
	Shell     string `json:"shell"`
}

func NewTask() Task {
	return Task{
		Command:   "",
		Namespace: "",
		Service:   "cli",
		Shell:     "sh",
	}
}

func (t Task) String() string {
	return fmt.Sprintf("{command: '%v', ns: '%v', service: '%v', shell:'%v'}", t.Command, t.Namespace, t.Service, t.Shell)
}

//TODO: do we want some kind of validation?

//TODO: build get config for kubernetes
// This will either be in cluster or out of cluster - we start with out of cluster to test
// TODO: BMK - ensure that this is responsive to the context
func GetK8sClient(config *rest.Config) (*kubernetes.Clientset, error) {
	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	return clientset, nil
}

func getConfig() (*rest.Config, error) {
	var kubeconfig *string
	kubeconfig = new(string)
	*kubeconfig = helpers.GetEnv("KUBECONFIG", "/home/bomoko/.kube/config", false)

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}
	return config, err
}

//TODO: code to exec pre/post rollout tasks
func ExecuteTaskInEnvironment(task Task) error {
	fmt.Println("Executing task ", task)

	return nil
}

func log(data string) {
	fmt.Printf("LOG[%v]:%v\n", time.Now().Format(time.Kitchen), data)
}

func ExecPod(
	podName, namespace string,
	command []string,
	//stdin io.Reader,
	tty bool,
	container string,
) (string, string, error) {

	restCfg, err := getConfig()
	if err != nil {
		return "", "", err
	}

	clientset, err := GetK8sClient(restCfg)
	if err != nil {
		return "", "", fmt.Errorf("unable to create client: %v", err)
	}

	depClient := clientset.AppsV1().Deployments(namespace)

	lagoonServiceLabel := "lagoon.sh/service=" + podName

	deployments, err := depClient.List(context.TODO(), v1.ListOptions{
		LabelSelector: lagoonServiceLabel,
	})
	if len(deployments.Items) == 0 {
		return "", "", errors.New("No deployments found matching label: " + lagoonServiceLabel)
	}

	deployment := &deployments.Items[0]

	// we want to scale the replicas here to 1, at least, before attempting the exec
	podReady := false
	numIterations := 0
	for ; !podReady; numIterations++ {
		if numIterations > 10 { //break if there's some reason we can't scale the pod
			return "", "", errors.New("Failed to scale pods for " + deployment.Name)
		}
		if deployment.Status.ReadyReplicas == 0 {
			log("No ready replicas found, scaling up")
			scale, err := clientset.AppsV1().Deployments(namespace).GetScale(context.TODO(), deployment.Name, v1.GetOptions{})
			if err != nil {
				return "", "", err
			}
			scale.Spec.Replicas = 1
			depClient.UpdateScale(context.TODO(), deployment.Name, scale, v1.UpdateOptions{})
			time.Sleep(3 * time.Second)
			deployment, err = depClient.Get(context.TODO(), deployment.Name, v1.GetOptions{})
			if err != nil {
				return "", "", err
			}
		} else {
			podReady = true
		}
	}

	//grab pod - for now we'll copy precisely what the build script does and use the labels

	podClient := clientset.CoreV1().Pods(namespace)
	clientList, err := podClient.List(context.TODO(), v1.ListOptions{
		LabelSelector: lagoonServiceLabel,
	})

	if err != nil {
		return "", "", err
	}

	var pod corev1.Pod
	foundRunningPod := false
	for _, i2 := range clientList.Items {
		if i2.Status.Phase == "Running" {
			if container != "" {
				//is this container contained herein?
				for _, c := range i2.Spec.Containers {
					if c.Name != container {
						continue
					}
				}
			}
			pod = i2
			foundRunningPod = true
			break
		}
	}
	if !foundRunningPod {
		return "", "", errors.New("Unable to find running Pod for namespace: " + namespace)
	}
	fmt.Println("Going to exec into ", pod.Name)

	req := clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(pod.Name).
		Namespace(namespace).
		SubResource("exec")

	scheme := runtime.NewScheme()

	if err := corev1.AddToScheme(scheme); err != nil {
		return "", "", fmt.Errorf("error adding to scheme: %v", err)
	}
	if len(command) == 0 {
		command = []string{"sh"}
	}

	parameterCodec := runtime.NewParameterCodec(scheme)
	req.VersionedParams(&corev1.PodExecOptions{
		Container: container,
		Command:   command,
		//Stdin:   stdin != nil,
		Stdout: true,
		Stderr: true,
		TTY:    tty,
	}, parameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(restCfg, "POST", req.URL())
	if err != nil {
		return "", "", fmt.Errorf("error while creating Executor: %v", err)
	}

	var stdout, stderr bytes.Buffer
	err = exec.Stream(remotecommand.StreamOptions{
		//Stdin:  stdin, //TODO: does this need to be implemented?
		Stdout: &stdout,
		Stderr: &stderr,
		Tty:    tty,
	})
	if err != nil {
		return "", "", fmt.Errorf("error in Stream: %v", err)
	}

	return stdout.String(), stderr.String(), nil

}
