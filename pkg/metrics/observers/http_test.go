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

func TestHttpObserver_GetRequestSuccessRate(t *testing.T) {
	expected := ` sum( rate( http_request_duration_seconds_count{ kubernetes_namespace="default", kubernetes_pod_name=~"podinfo-[0-9a-zA-Z]+(-[0-9a-zA-Z]+)", status!~"5.*" }[1m] ) ) / sum( rate( http_request_duration_seconds_count{ kubernetes_namespace="default", kubernetes_pod_name=~"podinfo-[0-9a-zA-Z]+(-[0-9a-zA-Z]+)" }[1m] ) ) * 100`

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

	observer := &HttpObserver{
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

func TestHttpObserver_GetRequestDuration(t *testing.T) {
	expected := ` histogram_quantile( 0.99, sum( rate( http_request_duration_seconds_bucket{ kubernetes_namespace="default", kubernetes_pod_name=~"podinfo-[0-9a-zA-Z]+(-[0-9a-zA-Z]+)" }[1m] ) ) by (le) )`

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		promql := r.URL.Query()["query"][0]
		if promql != expected {
			t.Errorf("\nGot %s \nWanted %s", promql, expected)
		}

		json := `{"status":"success","data":{"resultType":"vector","result":[{"metric":{},"value":[1,"0.100"]}]}}`
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

	observer := &HttpObserver{
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
