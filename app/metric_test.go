package app_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/weaveworks/scope/app"
	"github.com/weaveworks/scope/probe/docker"
	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/test"
	"github.com/weaveworks/scope/test/fixture"
	"github.com/weaveworks/scope/test/reflect"
)

func TestNewMetricStorage(t *testing.T) {
	{
		// It errors for unknown scheme
		s, err := app.NewMetricStorage("foobar://localhost:9090")
		if s != nil {
			t.Errorf("Expected s to be nil, but was: %v", s)
		}
		wantError := "Unknown metrics storage scheme: \"foobar\""
		if err == nil || err.Error() != wantError {
			t.Errorf("Expected err to be %q, but was: %v", wantError, err)
		}
	}
}

func TestMetricListEndpoint(t *testing.T) {
	router := mux.NewRouter()
	app.RegisterMetricRoutes(app.NewCollector(1*time.Minute), nil, router)

	{
		// /api/metrics lists all metrics
		w := httptest.NewRecorder()
		r, _ := http.NewRequest("GET", "/api/metrics", nil)
		router.ServeHTTP(w, r)
		var have []string
		if err := json.NewDecoder(w.Body).Decode(&have); err != nil {
			t.Error(err)
		}
		// TODO: finish filling this in
		want := []string{
			fmt.Sprintf("%s{topology=%q,node=%q}", docker.CPUTotalUsage, report.Container, fixture.ClientContainerNodeID),
		}
		if !reflect.DeepEqual(want, have) {
			t.Error(test.Diff(want, have))
		}
	}
}

func TestMetricQueryEndpoint(t *testing.T) {
	router := mux.NewRouter()
	app.RegisterMetricRoutes(app.NewCollector(1*time.Minute), nil, router)

	{
		// /api/metrics/query queries a metric
		w := httptest.NewRecorder()
		r, _ := http.NewRequest("GET", "/api/metrics/query?q=", nil)
		router.ServeHTTP(w, r)
		var have map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&have); err != nil {
			t.Error(err)
		}
		// TODO: finish filling this in
		want := map[string]interface{}{}
		if !reflect.DeepEqual(want, have) {
			t.Error(test.Diff(want, have))
		}
	}
}
