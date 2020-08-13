// based on https://github.com/kubernetes-sigs/controller-runtime/tree/master/examples/builtins
package main

import (
	"context"
	"encoding/json"
	"net/http"

	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// +kubebuilder:webhook:path=/mutate-v1-pod,mutating=true,failurePolicy=fail,groups="",resources=pods,verbs=create;update,versions=v1,name=mpod.kb.io

var (
	webhookLogger = ctrl.Log.WithName("webhook")
	windows2019   = "windows-2019"
	windows2004   = "windows-2004"
)

// podAnnotator annotates Pods
type podAnnotator struct {
	Client  client.Client
	decoder *admission.Decoder
}

// podAnnotator adds an annotation to every incoming pods.
func (a *podAnnotator) Handle(ctx context.Context, req admission.Request) admission.Response {
	pod := &corev1.Pod{}

	err := a.decoder.Decode(req, pod)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	if pod.Annotations == nil {
		pod.Annotations = map[string]string{}
	}
	pod.Annotations["example-mutating-admission-webhook"] = "foo"

	os := pod.Spec.NodeSelector["kubernetes.io/os"]
	webhookLogger.Info("HERE")
	webhookLogger.Info(os)
	if os == "windows" {
		webhookLogger.Info("pod set for windows")
		webhookLogger.Info(pod.Spec.NodeName)
		webhookLogger.Info(pod.Status.NominatedNodeName)
		pod.Spec.RuntimeClassName = &windows2004
		c := pod.Spec.Containers[0]
		c.Image = c.Image + "-windows-amd64-2004"
		pod.Spec.Containers[0] = c
	}

	marshaledPod, err := json.Marshal(pod)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	return admission.PatchResponseFromRaw(req.Object.Raw, marshaledPod)
}

// podAnnotator implements admission.DecoderInjector.
// A decoder will be automatically injected.

// InjectDecoder injects the decoder.
func (a *podAnnotator) InjectDecoder(d *admission.Decoder) error {
	a.decoder = d
	return nil
}
