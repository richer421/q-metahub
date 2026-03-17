package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/richer421/q-metahub/infra/mysql/dao"
	"github.com/richer421/q-metahub/pkg/testutil"
)

func setupRouterWithMockDB(t *testing.T) (*gin.Engine, sqlmock.Sqlmock) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	db, mock, err := testutil.NewMockDB()
	require.NoError(t, err)
	dao.SetDefault(db)

	r := gin.New()
	v1 := r.Group("/api/v1")
	RegisterMetadataRoutes(v1)
	return r, mock
}

func decodeBody(t *testing.T, body *bytes.Buffer) map[string]any {
	t.Helper()
	var resp map[string]any
	require.NoError(t, json.Unmarshal(body.Bytes(), &resp))
	return resp
}

func TestMetadataAPICreateInstanceOAM(t *testing.T) {
	router, _ := setupRouterWithMockDB(t)

	reqBody := map[string]any{
		"name":           "q-demo-dev",
		"env":            "dev",
		"schema_version": "v1alpha1",
		"frontend_payload": map[string]any{
			"basic": map[string]any{
				"name": "q-demo",
			},
			"extended": map[string]any{
				"network_mode": "k8s_service",
				"ports":        []int{8080},
			},
		},
	}
	data, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/instance-oams", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := decodeBody(t, w.Body)
	assert.Equal(t, float64(0), resp["code"])
	payload := resp["data"].(map[string]any)
	assert.Equal(t, "q-demo-dev", payload["name"])
	assert.Equal(t, "dev", payload["env"])
}

func TestMetadataAPIGetDeployPlanRuntimeSpec(t *testing.T) {
	router, mock := setupRouterWithMockDB(t)

	now := time.Now()
	mock.ExpectQuery("SELECT \\* FROM `deploy_plans` WHERE `deploy_plans`.`id` = .*").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "created_at", "updated_at", "name", "description",
			"business_unit_id", "ci_config_id", "cd_config_id", "instance_oam_id",
		}).AddRow(6, now, now, "q-demo-dev-plan", "demo deploy plan", 2, 3, 4, 5))
	mock.ExpectQuery("SELECT \\* FROM `business_units` WHERE `business_units`.`id` = .*").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "created_at", "updated_at", "name", "description", "project_id",
		}).AddRow(2, now, now, "q-demo", "demo business unit", 1))
	mock.ExpectQuery("SELECT \\* FROM `projects` WHERE `projects`.`id` = .*").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "created_at", "updated_at", "git_id", "name", "repo_url",
		}).AddRow(1, now, now, 1001, "q-demo-project", "https://github.com/richer421/q-demo.git"))
	mock.ExpectQuery("SELECT \\* FROM `ci_configs` WHERE `ci_configs`.`id` = .*").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "created_at", "updated_at", "name", "business_unit_id",
			"image_registry", "image_repo", "image_tag_rule", "build_spec",
		}).AddRow(3, now, now, "q-demo-ci", 2, "harbor.local", "q-demo/q-demo", []byte(`{"type":"commit"}`), []byte(`{"branch":"main"}`)))
	mock.ExpectQuery("SELECT \\* FROM `cd_configs` WHERE `cd_configs`.`id` = .*").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "created_at", "updated_at", "name", "business_unit_id",
			"render_engine", "values_yaml", "release_strategy", "git_ops",
		}).AddRow(4, now, now, "q-demo-cd", 2, "helm", "replicaCount: 1\n", []byte(`{"deployment_mode":"rolling"}`), []byte(`{"enabled":true}`)))
	mock.ExpectQuery("SELECT \\* FROM `instance_oams` WHERE `instance_oams`.`id` = .*").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "created_at", "updated_at", "name", "business_unit_id",
			"env", "schema_version", "oam_application",
		}).AddRow(5, now, now, "q-demo-dev", 2, "dev", "v1alpha1", []byte(`{"component":{"type":"pod","properties":{"mainContainer":{"name":"q-demo"}}}}`)))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/deploy-plans/6/full-spec", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := decodeBody(t, w.Body)
	assert.Equal(t, float64(0), resp["code"])
	data := resp["data"].(map[string]any)
	assert.Equal(t, float64(6), data["deploy_plan"].(map[string]any)["id"])
	assert.Equal(t, float64(4), data["cd_config"].(map[string]any)["id"])
	assert.Equal(t, "dev", data["instance_oam"].(map[string]any)["env"])
	require.NoError(t, mock.ExpectationsWereMet())
}
