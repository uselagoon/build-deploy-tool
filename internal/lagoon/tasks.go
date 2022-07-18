package lagoon

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/uselagoon/build-deploy-tool/internal/helpers"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"
)

var debug bool

// Task .
type Task struct {
	Name      string `json:"name"`
	Command   string `json:"command"`
	Namespace string `json:"namespace"`
	Service   string `json:"service"`
	Shell     string `json:"shell"`
	Container string `json:"container"`
	When      string `json:"when"`
}

// NewTask .
func NewTask() Task {
	return Task{
		Command:   "",
		Namespace: "",
		Service:   "cli",
		Shell:     "sh",
	}
}

type DeploymentMissingError struct {
	ErrorText string
}

func (e *DeploymentMissingError) Error() string {
	return e.ErrorText
}

type PodScalingError struct {
	ErrorText string
}

func (e *PodScalingError) Error() string {
	return e.ErrorText
}

func (t Task) String() string {
	return fmt.Sprintf("{command: '%v', ns: '%v', service: '%v', shell:'%v'}", t.Command, t.Namespace, t.Service, t.Shell)
}

// GetK8sClient .
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
	*kubeconfig = helpers.GetEnv("KUBECONFIG", "", false)

	if *kubeconfig == "" {
		//return nil, fmt.Errorf("Unable to find a valid KUBECONFIG")
		//Fall back on out of cluster

		// read the deployer token.
		token, err := ioutil.ReadFile("/var/run/secrets/lagoon/deployer/token")
		if err != nil {
			return nil, err
		}
		// generate the rest config for the client.
		restCfg := &rest.Config{
			BearerToken: string(token),
			Host:        "https://kubernetes.default.svc",
			TLSClientConfig: rest.TLSClientConfig{
				Insecure: true,
			},
		}
		return restCfg, nil
	}
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}
	return config, err
}

// ExecuteTaskInEnvironment .
func ExecuteTaskInEnvironment(task Task) error {
	if debug {
		fmt.Printf("Executing task '%v':'%v'\n", task.Name, task.Command)
	}
	command := make([]string, 0, 5)
	if task.Shell != "" {
		command = append(command, task.Shell)
	} else {
		command = append(command, "sh")
	}

	command = append(command, "-c")
	command = append(command, task.Command)

	stdout, stderr, err := ExecPod(task.Service, task.Namespace, command, false, task.Container)

	if err != nil {
		fmt.Printf("*** Task '%v' failed - STDOUT and STDERR follows ***\n", task.Name)
	}

	if len(stdout) > 0 {
		fmt.Printf("*** Task STDOUT ***\n %v \n *** STDOUT Ends ***\n", stdout)
	}
	if len(stderr) > 0 {
		fmt.Printf("*** Task STDERR ***\n %v \n *** STDERR Ends ***\n", stderr)
	}

	return err
}

// ExecPod .
func ExecPod(
	podName, namespace string,
	command []string,
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
	if err != nil {
		return "", "", err
	}

	if len(deployments.Items) == 0 {
		return "", "", &DeploymentMissingError{ErrorText: "No deployments found matching label: " + lagoonServiceLabel}
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
			fmt.Println("No ready replicas found, scaling up")

			scale, err := clientset.AppsV1().Deployments(namespace).GetScale(context.TODO(), deployment.Name, v1.GetOptions{})
			if err != nil {
				return "", "", err
			}

			if scale.Spec.Replicas == 0 {
				scale.Spec.Replicas = 1
				depClient.UpdateScale(context.TODO(), deployment.Name, scale, v1.UpdateOptions{})
			}

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
		if i2.Status.Phase == "Running" && i2.ObjectMeta.DeletionTimestamp == nil {
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
		return "", "", &PodScalingError{
			ErrorText: "Unable to find running Pod for namespace: " + namespace,
		}
	}
	if debug {
		fmt.Println("Going to exec into ", pod.Name)
	}
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
		Stdout:    true,
		Stderr:    true,
		TTY:       tty,
	}, parameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(restCfg, "POST", req.URL())
	if err != nil {
		return "", "", fmt.Errorf("error while creating Executor: %v", err)
	}

	var stdout, stderr bytes.Buffer
	err = exec.Stream(remotecommand.StreamOptions{
		Stdout: &stdout,
		Stderr: &stderr,
		Tty:    tty,
	})
	if err != nil {
		return stdout.String(), stderr.String(), fmt.Errorf("Error returned: %v", err)
	}

	return stdout.String(), stderr.String(), nil

}

func init() {
	//TODO: will potentially be useful to wire this up to the global debug into
	debug = true
}
