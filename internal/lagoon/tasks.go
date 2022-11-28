package lagoon

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"strconv"
	"time"

	"github.com/uselagoon/build-deploy-tool/internal/helpers"
	appsv1 "k8s.io/api/apps/v1"
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
	Name                string `json:"name"`
	Command             string `json:"command"`
	Namespace           string `json:"namespace"`
	Service             string `json:"service"`
	Shell               string `json:"shell"`
	Container           string `json:"container"`
	When                string `json:"when"`
	Weight              int    `json:"weight"`
	ScaleWaitTime       int    `json:"scaleWaitTime"`
	ScaleMaxIterations  int    `json:"scaleMaxIterations"`
	RequiresEnvironment bool   `json:"requiresEnvironment"`
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
		//Fall back on out of cluster
		// read the deployer token.
		token, err := ioutil.ReadFile("/var/run/secrets/lagoon/deployer/token")
		if err != nil {
			token, err = ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/token")
			if err != nil {
				return nil, err
			}
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

	stdout, stderr, err := ExecPod(task.Service, task.Namespace, command, false, task.Container, task.ScaleWaitTime, task.ScaleMaxIterations)

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
	waitTime, maxIterations int,
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
	numIterations := 1
	for ; !podReady; numIterations++ {
		if numIterations >= maxIterations { //break if there's some reason we can't scale the pod
			return "", "", errors.New("Failed to scale pods for " + deployment.Name)
		}
		if deployment.Status.ReadyReplicas == 0 {
			fmt.Println(fmt.Sprintf("No ready replicas found, scaling up. Attempt %d/%d", numIterations, maxIterations))

			scale, err := clientset.AppsV1().Deployments(namespace).GetScale(context.TODO(), deployment.Name, v1.GetOptions{})
			if err != nil {
				return "", "", err
			}

			if scale.Spec.Replicas == 0 {
				scale.Spec.Replicas = 1
				depClient.UpdateScale(context.TODO(), deployment.Name, scale, v1.UpdateOptions{})
			}
			time.Sleep(time.Second * time.Duration(waitTime))
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

// The following two functions are shamelessly plucked from https://github.com/uselagoon/lagoon-ssh-portal/pull/104/files

// unidleReplicas checks the unidle-replicas annotation for the number of
// replicas to restore. If the label cannot be read or parsed, 1 is returned.
// The return value is clamped to the interval [1,16].
func unidleReplicas(deploy appsv1.Deployment) int {
	rs, ok := deploy.Annotations["idling.amazee.io/unidle-replicas"]
	if !ok {
		return 1
	}
	r, err := strconv.Atoi(rs)
	if err != nil || r < 1 {
		return 1
	}
	if r > 16 {
		return 16
	}
	return r
}

// unidleNamespace scales all deployments with the
// "idling.amazee.io/watch=true" label up to the number of replicas in the
// "idling.amazee.io/unidle-replicas" label.

var NamespaceUnidlingTimeoutError = errors.New("Unable to scale idled deployments due to timeout")

func UnidleNamespace(ctx context.Context, namespace string, retries int, waitTime int) error {
	restCfg, err := getConfig()
	if err != nil {
		return err
	}

	clientset, err := GetK8sClient(restCfg)
	if err != nil {
		return fmt.Errorf("unable to create client: %v", err)
	}

	deploys, err := clientset.AppsV1().Deployments(namespace).List(ctx, v1.ListOptions{
		LabelSelector: "idling.amazee.io/watch=true",
	})
	if err != nil {
		return fmt.Errorf("couldn't select deploys by label: %v", err)
	}
	for _, deploy := range deploys.Items {
		// check if idled
		s, err := clientset.AppsV1().Deployments(namespace).
			GetScale(ctx, deploy.Name, v1.GetOptions{})
		if err != nil {
			return fmt.Errorf("couldn't get deployment scale: %v", err)
		}
		if s.Spec.Replicas > 0 {
			continue
		}
		// scale up the deployment
		sc := *s
		sc.Spec.Replicas = int32(unidleReplicas(deploy))
		_, err = clientset.AppsV1().Deployments(namespace).
			UpdateScale(ctx, deploy.Name, &sc, v1.UpdateOptions{})
		if err != nil {
			return fmt.Errorf("couldn't scale deployment: %v", err)
		}
	}

	// Let's wait for the various deployments to scale
	scaled := true
	scaledDeps := make(map[string]bool)
	for countdown := retries; len(deploys.Items) > 0 && countdown > 0; countdown-- {
		time.Sleep(time.Second * time.Duration(waitTime))
		for _, deploy := range deploys.Items {
			s, err := clientset.AppsV1().Deployments(namespace).Get(ctx, deploy.Name, v1.GetOptions{})
			if err != nil {
				return err
			}
			if s.Status.ReadyReplicas > 0 {
				scaledDeps[deploy.Name] = true
			}
		}
		if len(scaledDeps) == len(deploys.Items) {
			scaled = true
			break
		} else {
			scaled = false
		}
	}

	if !scaled {
		return NamespaceUnidlingTimeoutError
	}

	return nil
}

func init() {
	//TODO: will potentially be useful to wire this up to the global debug into
	debug = true
}
