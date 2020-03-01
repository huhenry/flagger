package router

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestIngressRouter_Reconcile(t *testing.T) {
	mocks := newFixture(nil)
	router := &IngressRouter{
		logger:            mocks.logger,
		kubeClient:        mocks.kubeClient,
		annotationsPrefix: "custom.ingress.kubernetes.io",
	}

	err := router.Reconcile(mocks.ingressCanary)
	require.NoError(t, err)

	canaryAn := "custom.ingress.kubernetes.io/canary"
	canaryWeightAn := "custom.ingress.kubernetes.io/canary-weight"

	canaryName := fmt.Sprintf("%s-canary", mocks.ingressCanary.Spec.IngressRef.Name)
	inCanary, err := router.kubeClient.ExtensionsV1beta1().Ingresses("default").Get(canaryName, metav1.GetOptions{})
	require.NoError(t, err)

	if _, ok := inCanary.Annotations[canaryAn]; !ok {
		t.Errorf("Canary annotation missing")
	}

	// test initialisation
	if inCanary.Annotations[canaryAn] != "false" {
		t.Errorf("Got canary annotation %v wanted false", inCanary.Annotations[canaryAn])
	}

	if inCanary.Annotations[canaryWeightAn] != "0" {
		t.Errorf("Got canary weight annotation %v wanted 0", inCanary.Annotations[canaryWeightAn])
	}
}

func TestIngressRouter_GetSetRoutes(t *testing.T) {
	mocks := newFixture(nil)
	router := &IngressRouter{
		logger:            mocks.logger,
		kubeClient:        mocks.kubeClient,
		annotationsPrefix: "prefix1.nginx.ingress.kubernetes.io",
	}

	err := router.Reconcile(mocks.ingressCanary)
	require.NoError(t, err)

	p, c, m, err := router.GetRoutes(mocks.ingressCanary)
	require.NoError(t, err)

	p = 50
	c = 50
	m = false

	err = router.SetRoutes(mocks.ingressCanary, p, c, m)
	require.NoError(t, err)

	canaryAn := "prefix1.nginx.ingress.kubernetes.io/canary"
	canaryWeightAn := "prefix1.nginx.ingress.kubernetes.io/canary-weight"

	canaryName := fmt.Sprintf("%s-canary", mocks.ingressCanary.Spec.IngressRef.Name)
	inCanary, err := router.kubeClient.ExtensionsV1beta1().Ingresses("default").Get(canaryName, metav1.GetOptions{})
	require.NoError(t, err)

	if _, ok := inCanary.Annotations[canaryAn]; !ok {
		t.Errorf("Canary annotation missing")
	}

	// test rollout
	if inCanary.Annotations[canaryAn] != "true" {
		t.Errorf("Got canary annotation %v wanted true", inCanary.Annotations[canaryAn])
	}

	if inCanary.Annotations[canaryWeightAn] != "50" {
		t.Errorf("Got canary weight annotation %v wanted 50", inCanary.Annotations[canaryWeightAn])
	}

	p = 100
	c = 0
	m = false

	err = router.SetRoutes(mocks.ingressCanary, p, c, m)
	require.NoError(t, err)

	inCanary, err = router.kubeClient.ExtensionsV1beta1().Ingresses("default").Get(canaryName, metav1.GetOptions{})
	require.NoError(t, err)

	// test promotion
	if inCanary.Annotations[canaryAn] != "false" {
		t.Errorf("Got canary annotation %v wanted false", inCanary.Annotations[canaryAn])
	}

	if inCanary.Annotations[canaryWeightAn] != "0" {
		t.Errorf("Got canary weight annotation %v wanted 0", inCanary.Annotations[canaryWeightAn])
	}
}
