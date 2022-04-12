package webhook

import (
	"context"
	"encoding/json"
	"log"
	"testing"

	simplescaletests "github.com/tensorbytes/simplescale/tests"
	"github.com/tensorbytes/simplescale/utils"
	appsv1 "k8s.io/api/apps/v1"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
	kubeadmission "sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

func TestJsonPatch(t *testing.T) {
	c, err := utils.GetControllerClient()
	if err != nil {
		panic(err)
	}
	e := simplescaletests.NewTestEnv()
	defer e.Clean()
	e.CreateDeployment("test-jsonpatch", "50m", "100Mi", 1)
	var originDeployment appsv1.Deployment
	err = c.Get(context.TODO(), runtimeclient.ObjectKey{Name: "test-jsonpatch", Namespace: "default"}, &originDeployment)
	if err != nil {
		log.Fatalln(err)
	}
	currentDeployment := originDeployment.DeepCopy()
	replicas := *currentDeployment.Spec.Replicas
	replicas = replicas + 1
	currentDeployment.Spec.Replicas = &replicas
	originJson, err := json.Marshal(originDeployment)
	if err != nil {
		panic(err)
	}
	currentJson, err := json.Marshal(currentDeployment)
	if err != nil {
		log.Fatal(err)
	}
	resp := kubeadmission.PatchResponseFromRaw(originJson, currentJson)
	for _, p := range resp.Patches {
		s, err := p.MarshalJSON()
		if err != nil {
			log.Fatal(err)
		}
		log.Println(string(s))
	}

}
