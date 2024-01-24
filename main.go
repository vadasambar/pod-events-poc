package main

import (
	"context"
	"path/filepath"
	"time"

	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	typedv1core "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/homedir"
)

func main() {
	kubeconfig := filepath.Join(homedir.HomeDir(), ".kube", "config")
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		log.Info("connecting with config in-cluster")
		config, err = rest.InClusterConfig()
		if err != nil {
			log.Fatal(err)
		}
	}
	if err != nil {
		log.Fatal(err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	pod, err := clientset.CoreV1().Pods("default").Get(context.TODO(), "nginx", metav1.GetOptions{})
	if err != nil {
		log.Fatal(err)
	}

	scheme := runtime.NewScheme()
	_ = v1.AddToScheme(scheme)

	// all the good events stuff is here
	eventBroadcaster := record.NewBroadcaster()
	// logs events to stdout at V4 level
	eventBroadcaster.StartStructuredLogging(1)
	eventBroadcaster.StartRecordingToSink(&typedv1core.EventSinkImpl{Interface: clientset.CoreV1().Events("")})
	eventRecorder := eventBroadcaster.NewRecorder(scheme, v1.EventSource{Component: "my-component"})
	eventRecorder.Event(pod, v1.EventTypeNormal, "this is a", "test event")
	// uncomment this to log a warning type event
	// eventRecorder.Event(pod, corev1.EventTypeWarning, "this is a", "test event")
	time.Sleep(time.Hour * 1)
	eventBroadcaster.Shutdown()
}
