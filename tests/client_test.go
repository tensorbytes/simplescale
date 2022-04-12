package tests

import (
	// simplescalev1 "github.com/tensorbytes/simplescale/api/v1alpha1"
	"context"
	"log"
	"testing"
	"time"

	autoscalev1alpha1 "github.com/tensorbytes/simplescale/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlcache "sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

var (
	deployments = []string{
		`{
			"apiVersion": "apps/v1",
			"kind": "Deployment",
			"metadata": {
				"name": "test-env",
				"namespace": "default"
			},
			"spec": {
				"selector":{
					"matchLabels": {
						"app": "test-env"
					}
				},
					
				"template":{
					"metadata": {
						"labels":{
							"app": "test-env"
						}
		
					},
					"spec": {
						"containers":[
							{
								"name": "test",
								"image": "alinux-optimized-tensorflow-registry.cn-hangzhou.cr.aliyuncs.com/tensorflow/tensorflow2",
								"command": ["python3","-m","http.server"]
							}
			
						]
					}
				}
			}
		}`,
	}
)

func TestCreateDeployment(t *testing.T) {
	e := NewTestEnv()
	e.PrepareResource(deployments)
	defer e.Clean()
}

func TestReaderClient(t *testing.T) {
	// prepare env
	e := NewTestEnv()
	e.PrepareResource(deployments)
	defer e.Clean()
	var config = ctrl.GetConfigOrDie()
	mapper, err := apiutil.NewDynamicRESTMapper(config)
	if err != nil {
		log.Println(err)
	}
	scheme := runtime.NewScheme()
	log.Println(scheme)
	clientgoscheme.AddToScheme(scheme)
	autoscalev1alpha1.AddToScheme(scheme)
	// log.Println(scheme)
	clientOptions := client.Options{Scheme: scheme, Mapper: mapper}

	// apiReader 不做缓存，所有操作直接操作 apiserver
	apiReader, err := client.New(config, clientOptions)
	if err != nil {
		log.Println(err)
	}
	log.Println(apiReader)

	namespaceName := client.ObjectKey{
		Namespace: "default",
		Name:      "test-env",
	}
	var deployment appsv1.Deployment
	if err = apiReader.Get(context.TODO(), namespaceName, &deployment); err != nil {
		log.Println(err)
	}
	log.Println(deployment.Name)

	newdeployment := deployment.DeepCopy()
	replicas := int32(4)
	newdeployment.Spec.Replicas = &replicas

	err = apiReader.Patch(context.TODO(), newdeployment, client.Merge, &client.PatchOptions{})
	log.Println(err)

	if err = apiReader.Get(context.TODO(), namespaceName, &deployment); err != nil {
		log.Println(err)
	}
	log.Println(deployment.Name)

	newdeployment = deployment.DeepCopy()
	replicas = int32(2)
	newdeployment.Spec.Replicas = &replicas

	err = apiReader.Patch(context.TODO(), newdeployment, client.Merge, &client.PatchOptions{})
	log.Println(err)

}

func TestWriteClient(t *testing.T) {
	var config = ctrl.GetConfigOrDie()
	mapper, err := apiutil.NewDynamicRESTMapper(config)
	if err != nil {
		log.Println(err)
	}
	scheme := runtime.NewScheme()
	log.Println(scheme)
	clientgoscheme.AddToScheme(scheme)
	autoscalev1alpha1.AddToScheme(scheme)
	// log.Println(scheme)
	clientOptions := client.Options{Scheme: scheme, Mapper: mapper}

	// apiReader 不做缓存，所有操作直接操作 apiserver
	apiReader, err := client.New(config, clientOptions)
	if err != nil {
		log.Println(err)
		return
	}
	log.Println(apiReader)

	// cache 默认是启动一个watch&list
	// 重新list同步时间
	resyncTime := 5 * time.Hour
	cacheReader, err := ctrlcache.New(config, ctrlcache.Options{Scheme: scheme, Mapper: mapper, Resync: &resyncTime})
	if err != nil {
		log.Println(err)
		return
	}
	// 启动cache
	go cacheReader.Start(context.TODO())

	// NewDelegatingClient 具有缓存能力，可以将部分缓存起来
	writeobj, err := client.NewDelegatingClient(client.NewDelegatingClientInput{
		CacheReader: cacheReader,
		Client:      apiReader,
	})
	if err != nil {
		log.Println(err)
		return
	}

	time.Sleep(1 * time.Second)

	namespaceName := client.ObjectKey{
		Namespace: "default",
		Name:      "test-env",
	}
	var deployment appsv1.Deployment
	if err = writeobj.Get(context.TODO(), namespaceName, &deployment); err != nil {
		log.Println(err)
	}
	log.Println(deployment.Name)

	newdeployment := deployment.DeepCopy()
	replicas := int32(3)
	newdeployment.Spec.Replicas = &replicas

	err = writeobj.Patch(context.TODO(), newdeployment, client.Merge, &client.PatchOptions{})
	log.Println(err)

	time.Sleep(1 * time.Second)

	if err = writeobj.Get(context.TODO(), namespaceName, &deployment); err != nil {
		log.Println(err)
	}
	log.Println(deployment.Name)

	newdeployment = deployment.DeepCopy()
	replicas = int32(1)
	newdeployment.Spec.Replicas = &replicas

	err = writeobj.Patch(context.TODO(), newdeployment, client.Merge, &client.PatchOptions{})
	log.Println(err)

}

// 测试创建deployment和assert是否可以正常运行
func TestCreateDeploymentAndAssert(t *testing.T) {
	e := NewTestEnv()
	defer e.Clean()
	objName := "test-obj-update"
	e.CreateDeployment(objName, "50m", "100Mi", 1)
	e.Assert("Deployment", "apps/v1", objName, "default", "spec.template.spec.containers.0.resources.requests.cpu", "50m")

}

func TestCreateAutoScalerAndAssert(t *testing.T) {
	e := NewTestEnv()
	defer e.Clean()
	e.CreateAutoScaler("test-autoscaler", "test-run-once-deployment", "spec.template.spec.containers.0.resources.limits.cpu", "test-scalefactor", "2")
	e.Assert("SimpleAutoScaler", "autoscale.scale.shikanon.com/v1alpha1", "test-autoscaler", "default", "spec.policy.0.scaleFactorObject.name", "test-scalefactor")
	e.Assert("RecommendationScaleFactor", "autoscale.scale.shikanon.com/v1alpha1", "test-scalefactor", "default", "status.scaleFactor", "2")
}
