package k8s

import (
	"context"
	"fmt"

	v1 "k8s.io/api/core/v1"
	policyv1beta1 "k8s.io/api/policy/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/wonderivan/logger"
)

const (
	EvictionKind       = "Eviction"
	PolicyGroupVersion = "policy/v1beta1"
)

func EvictNodePods(nodeName string, k8sClient *kubernetes.Clientset) error {
	pods, err := k8sClient.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{
		FieldSelector: "spec.nodeName=" + nodeName,
	})
	if err != nil {
		return err
	}
	for _, i := range pods.Items {
		err := EvictPod(k8sClient, i, PolicyGroupVersion)
		if err != nil {
			return err
		}
	}
	return nil
}

func EvictPod(k8sClient *kubernetes.Clientset, pod v1.Pod, policyGroupVersion string) error {
	deleteOptions := &metav1.DeleteOptions{}
	eviction := &policyv1beta1.Eviction{
		TypeMeta: metav1.TypeMeta{
			APIVersion: policyGroupVersion,
			Kind:       EvictionKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      pod.Name,
			Namespace: pod.Namespace,
		},
		DeleteOptions: deleteOptions,
	}
	return k8sClient.PolicyV1beta1().Evictions(eviction.Namespace).Evict(context.TODO(), eviction)
}

func CordonUnCordon(k8sClient *kubernetes.Clientset, nodeName string, cordoned bool) error {
	node, err := GetNodeByName(k8sClient, nodeName)
	if err != nil {
		return err
	}
	if node.Spec.Unschedulable == cordoned {
		logger.Alert("Node %s is already cordoned: %v", nodeName, cordoned)
		return nil
	}
	node.Spec.Unschedulable = cordoned
	_, err = k8sClient.CoreV1().Nodes().Update(context.TODO(), node, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("Error setting cordoned state for  %s node err: %v", nodeName, err)
	}
	return nil
}
