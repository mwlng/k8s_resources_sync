package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

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

	"github.com/mwlng/k8s_resources_sync/pkg/helpers"
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
	srcEksClusterName := flag.String("source_cluster_name", "", "Source k8s cluster name")
	rootPath := flag.String("rootpath", "", "Specified root path of k8s resource manifest files")

	deploymentFlag := flag.Bool("deployment", false, "Sync k8s deployment resources")
	serviceFlag := flag.Bool("service", false, "Sync k8s service resources")
	cronFlag := flag.Bool("cronjob", false, "Sync k8s cron job resources")
	saFlag := flag.Bool("serviceaccount", false, "Sync k8s service account resources")
	crFlag := flag.Bool("clusterrole", false, "Sync k8s cluster role resources")
	crbFlag := flag.Bool("clusterrole", false, "Sync k8s cluster role binding resources")

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

	sourceKubeConfig, err := helpers.GetKubeConfig(*srcEksClusterName, *kubeconfig)
	if err != nil {
		panic(err)
	}

	klog.Infof("Starting to sync k8s resources from %s in %s ...", sourceKubeConfig.Host, *environ)
	if *deploymentFlag {
		klog.Infof("Syncing k8s deployment resources to %s ...", targetKubeConfig.Host)
		deployments := helpers.LoadDeploymentYamlFiles(eksFilesRootPath)
		for _, d := range deployments {
			klog.Infof("* Deployment: %s\n", d.ObjectMeta.Name)
		}
		deployments = helpers.SyncDeployments(sourceKubeConfig, deployments)
		//PrintDeployments(deployments)
		helpers.ApplyDeployments(targetKubeConfig, deployments)
	} else if *serviceFlag {
		klog.Infof("Syncing k8s service resources to %s ...", targetKubeConfig.Host)
		services := helpers.LoadServiceYamlFiles(eksFilesRootPath)
		for _, s := range services {
			klog.Infof("* Service: %s\n", s.ObjectMeta.Name)
		}
		services = helpers.SyncServices(sourceKubeConfig, services, *environ)
		//PrintServices(services)
		helpers.ApplyServices(targetKubeConfig, services)
	} else if *cronFlag {
		klog.Infof("Syncing k8s cron jobs to %s ...", targetKubeConfig.Host)
		cronJobs := helpers.LoadCronJobYamlFiles(eksFilesRootPath)
		for _, job := range cronJobs {
			klog.Infof("* cron job: %s\n", job.ObjectMeta.Name)
		}
		cronJobs = helpers.SyncCronJobs(sourceKubeConfig, cronJobs)
		//PrintCronJobs(cronJobs)
		helpers.ApplyCronJobs(targetKubeConfig, cronJobs)
	} else if *saFlag {
		klog.Infof("Syncing k8s service accounts to %s ...", targetKubeConfig.Host)
		serviceAccounts := helpers.LoadServiceAccountYamlFiles(eksFilesRootPath)
		for _, account := range serviceAccounts {
			klog.Infof("* service account: %s\n", account.ObjectMeta.Name)
		}
		serviceAccounts = helpers.SyncServiceAccounts(sourceKubeConfig, serviceAccounts)
		//PrintCronJobs(cronJobs)
		helpers.ApplyServiceAccounts(targetKubeConfig, serviceAccounts)
	} else if *crFlag {
		klog.Infof("Syncing k8s cluster roles to %s ...", targetKubeConfig.Host)
		clusterRoles := helpers.LoadClusterRoleYamlFiles(eksFilesRootPath)
		for _, role := range clusterRoles {
			klog.Infof("* cluster role: %s\n", role.ObjectMeta.Name)
		}
		clusterRoles = helpers.SyncClusterRoles(sourceKubeConfig, clusterRoles)
		//PrintClusterRoles(clusterRoles)
		helpers.ApplyClusterRoles(targetKubeConfig, clusterRoles)
	} else if *crbFlag {
		klog.Infof("Syncing k8s cluster role bindings to %s ...", targetKubeConfig.Host)
		clusterRoleBindings := helpers.LoadClusterRoleBindingYamlFiles(eksFilesRootPath)
		for _, roleBinding := range clusterRoleBindings {
			klog.Infof("* cluster role binding: %s\n", roleBinding.ObjectMeta.Name)
		}
		clusterRoleBindings = helpers.SyncClusterRoleBindings(sourceKubeConfig, clusterRoleBindings)
		//PrintClusterRoleBindings(clusterRoleBindings)
		helpers.ApplyClusterRoleBindings(targetKubeConfig, clusterRoleBindings)
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
