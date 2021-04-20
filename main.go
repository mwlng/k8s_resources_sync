package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	//"gopkg.in/yaml.v2"

	corev1 "k8s.io/api/core/v1"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"k8s.io/klog/v2"

	//
	// Uncomment to load all auth plugins
	// _ "k8s.io/client-go/plugin/pkg/client/auth"
	//
	// Or uncomment to load specific auth plugins
	// _ "k8s.io/client-go/plugin/pkg/client/auth/azure"
	// _ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	// _ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	// _ "k8s.io/client-go/plugin/pkg/client/auth/openstack"

	"github.com/mwlng/k8s_resources_sync/pkg/k8s_resources"
)

var (
	eksPaths map[string]string = map[string]string{
		"alpha": "/home/ssm-user/backup/eks/dev/alphaeks/app-services",
		"qa":    "/home/ssm-user/backup/eks/qa/qaeks/app-services",
		"prod":  "/home/ssm-user/backup/eks/prod/prodeks/app-services",
	}
)

func init() {
	klog.InitFlags(nil)
}

func main() {
	defer func() {
		klog.Flush()
	}()

	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}

	environ := flag.String("e", "alpha", "Target environment")
	deploymentFlag := flag.Bool("deployment", false, "Sync k8s deployment resources")
	serviceFlag := flag.Bool("service", false, "Sync k8s service resources")
	cronFlag := flag.Bool("cronjob", false, "Sync k8s cron job resources")

	flag.Set("v", "2")
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err)
	}

	klog.Infof("Starting to sync k8s resources in %s ...", *environ)
	if *deploymentFlag {
		klog.Infof("Syncing k8s deployments resources to %s ...", environ)
		syncDeployments(environ, config)
	} else if *serviceFlag {
		klog.Infof("Syncing k8s services to %s ...", environ)
		syncServices(environ, config)
	} else if *cronFlag {
		klog.Infof("Syncing k8s cron jobs to %s ...", environ)
		syncCronJobs(environ, config)
	} else {
		klog.Infoln("No specified k8s resources to sync, exit !")
		Usage()
	}
}

func Usage() {
	fmt.Printf("Usage of %s:\n", os.Args[0])
	flag.PrintDefaults()
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func syncDeployments(environ *string, kubeConfig *rest.Config) {
	// List Deployments
	fmt.Printf("Listing deployments in namespace %q:\n", corev1.NamespaceDefault)
	deployment, err := k8s_resources.NewDeployment(kubeConfig, corev1.NamespaceDefault)
	if err != nil {
		panic(err)
	}

	list, err := deployment.ListDeloyments()
	if err != nil {
		panic(err)
	}

	for _, d := range list.Items {
		names := []string{}
		for _, n := range strings.Split(d.Name, "-") {
			if strings.ToLower(n) != "deployment" {
				names = append(names, n)
			}
		}
		app_name := strings.Join(names, "-")
		deployment_manifest := fmt.Sprintf("%s/%s/deployment-%s.yaml", eksPaths[*environ], app_name, app_name)
		if fileExists(deployment_manifest) {
			for _, c := range d.Spec.Template.Spec.Containers {
				fmt.Printf("* Found %s (image: %s)\n", d.Name, c.Image)
			}
		} else {
			fmt.Printf("x Can't find deployement manifest file: %s \n", deployment_manifest)
		}
	}
}

func syncServices(environ *string, kubeConfig *rest.Config) {

}

func syncCronJobs(environ *string, kubeConfig *rest.Config) {

}

func int32Ptr(i int32) *int32 { return &i }

func prompt() {
	fmt.Printf("-> Press Return key to continue.")
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		break
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}
	fmt.Println()
}
