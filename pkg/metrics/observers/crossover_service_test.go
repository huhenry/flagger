package observers

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	flaggerv1 "github.com/weaveworks/flagger/pkg/apis/flagger/v1beta1"
	"github.com/weaveworks/flagger/pkg/metrics/providers"
)

func TestCrossoverServiceObserver_GetRequestSuccessRate(t *testing.T) {
	expected := ` sum( rate( envoy_cluster_upstream_rq{ kubernetes_namespace="default", envoy_cluster_name="podinfo-canary", envoy_response_code!~"5.*" }[1m] ) ) / sum( rate( envoy_cluster_upstream_rq{ kubernetes_namespace="default", envoy_cluster_name="podinfo-canary" }[1m] ) ) * 100`

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		promql := r.URL.Query()["query"][0]
		if promql != expected {
			t.Errorf("\nGot %s \nWanted %s", promql, expected)
		}

		json := `{"status":"success","data":{"resultType":"vector","result":[{"metric":{},"value":[1,"100"]}]}}`
		w.Write([]byte(json))
	}))
	defer ts.Close()

	client, err := providers.NewPrometheusProvider(flaggerv1.MetricTemplateProvider{
		Type:      "prometheus",
		Address:   ts.URL,
		SecretRef: nil,
	}, nil)
	if err != nil {
		t.Fatal(err)
	}

	observer := &CrossoverServiceObserver{
		client: client,
	}

	val, err := observer.GetRequestSuccessRate(flaggerv1.MetricTemplateModel{
		Name:      "podinfo",
		Namespace: "default",
		Target:    "podinfo",
		Service:   "podinfo",
		Interval:  "1m",
	})
	require.NoError(t, err)

	if val != 100 {
		t.Errorf("Got %v wanted %v", val, 100)
	}
}

func TestCrossoverServiceObserver_GetRequestDuration(t *testing.T) {
	expected := ` histogram_quantile( 0.99, sum( rate( envoy_cluster_upstream_rq_time_bucket{ kubernetes_namespace="default", envoy_cluster_name="podinfo-canary" }[1m] ) ) by (le) )`

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		promql := r.URL.Query()["query"][0]
		if promql != expected {
			t.Errorf("\nGot %s \nWanted %s", promql, expected)
		}

		json := `{"status":"success","data":{"resultType":"vector","result":[{"metric":{},"value":[1,"100"]}]}}`
		w.Write([]byte(json))
	}))
	defer ts.Close()

	client, err := providers.NewPrometheusProvider(flaggerv1.MetricTemplateProvider{
		Type:      "prometheus",
		Address:   ts.URL,
		SecretRef: nil,
	}, nil)
	if err != nil {
		t.Fatal(err)
	}

	observer := &CrossoverServiceObserver{
		client: client,
	}

	val, err := observer.GetRequestDuration(flaggerv1.MetricTemplateModel{
		Name:      "podinfo",
		Namespace: "default",
		Target:    "podinfo",
		Service:   "podinfo",
		Interval:  "1m",
	})
	require.NoError(t, err)

	if val != 100*time.Millisecond {
		t.Errorf("Got %v wanted %v", val, 100*time.Millisecond)
	}
}
