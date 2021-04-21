package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	//"gopkg.in/yaml.v2"

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

	"github.com/mwlng/k8s_resources_sync/pkg/utils"
)

var (
	eksPaths map[string]string = map[string]string{
		"alpha": "/home/ssm-user/backup/eks/dev/alphaeks/app-services",
		"qa":    "/home/ssm-user/backup/eks/qa/qaeks/app-services",
		"prod":  "/home/ssm-user/backup/eks/prod/prodeks/app-services",
	}

	homeDir string
)

func init() {
	klog.InitFlags(nil)
	homeDir = utils.GetHomeDir()
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
	srcEksClusterName := flag.String("source_cluster_name", "", "k8s source cluster name")
	rootPath := flag.String("rootpath", "", "Specified root path of k8s resource manifest files")

	flag.Set("v", "2")
	flag.Parse()

	if len(*srcEksClusterName) == 0 {
		klog.Infoln("No specified source k8s cluster name, nothing to sync exit !")
		Usage()
		os.Exit(0)
	}

	eksFilesRootPath := eksPaths[*environ]
	if len(*rootPath) > 0 {
		eksFilesRootPath = utils.NormalizePath(*rootPath)
	}

	klog.Infoln("Loading client kubeconfig ...")
	targetKubeConfig, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err)
	}

	sourceKubeConfig, err := GetKubeConfig(*srcEksClusterName, *kubeconfig)
	if err != nil {
		panic(err)
	}

	klog.Infof("Starting to sync k8s resources from %s in %s ...", sourceKubeConfig.Host, *environ)
	if *deploymentFlag {
		klog.Infof("Syncing k8s deployment resources to %s ...", targetKubeConfig.Host)
		deployments := LoadDeploymentYamlFiles(eksFilesRootPath)
		for _, d := range deployments {
			klog.Infof("* Deployment: %s\n", d.ObjectMeta.Name)
		}
		SyncDeployments(sourceKubeConfig, deployments)
		PrintDeployments(deployments)
		//ApplyDeployments(targetKubeConfig, deployments)
	} else if *serviceFlag {
		klog.Infof("Syncing k8s service resources to %s ...", targetKubeConfig.Host)
		services := LoadServiceYamlFiles(eksFilesRootPath)
		for _, s := range services {
			klog.Infof("* Service: %s\n", s.ObjectMeta.Name)
		}
		services = SyncServices(sourceKubeConfig, services)
		PrintServices(services)
	} else if *cronFlag {
		klog.Infof("Syncing k8s cron jobs to %s ...", targetKubeConfig.Host)
		cronJobs := LoadCronJobYamlFiles(eksFilesRootPath)
		for _, job := range cronJobs {
			klog.Infof("* cron job: %s\n", job.ObjectMeta.Name)
		}
		cronJobs = SyncCronJobs(sourceKubeConfig, cronJobs)
		PrintCronJobs(cronJobs)
	} else {
		klog.Infoln("No specified k8s resources to sync, exit !")
		Usage()
	}
}

func Usage() {
	fmt.Println()
	fmt.Printf("Usage of %s:\n", os.Args[0])
	flag.PrintDefaults()
}
