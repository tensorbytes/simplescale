package main

import (
	"flag"
	"net/http"

	webhookhandler "github.com/tensorbytes/simplescale/webhook"
	kubeflag "k8s.io/component-base/cli/flag"
	klogv2 "k8s.io/klog/v2"
	"k8s.io/klog/v2/klogr"
	"sigs.k8s.io/controller-runtime/pkg/runtime/inject"
	kubeadmission "sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

var (
	certPath   = flag.String("cert", "cert.pem", `The path of cert`)
	keyPath    = flag.String("key", "private.key", `The path of private key`)
	listenPort = flag.String("port", ":8000", `The port of listen`)
)

func main() {
	klogv2.InitFlags(nil)
	kubeflag.InitFlags()

	mutation, err := webhookhandler.NewScaleMutatingHandler()
	if err != nil {
		klogv2.Exitln(err)
	}
	defer mutation.Close()
	go mutation.Sync()

	var logger = klogr.New()
	webhook := kubeadmission.Webhook{
		Handler: mutation,
	}
	_, err = inject.LoggerInto(logger, &webhook)
	if err != nil {
		klogv2.Exitln(err)
	}
	http.HandleFunc("/mutating-resource", webhook.ServeHTTP)
	err = http.ListenAndServeTLS(*listenPort, *certPath, *keyPath, nil)
	if err != nil {
		klogv2.Exitln(err)
	}

}
