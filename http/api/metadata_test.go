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

func TestMetadataAPICreateDeployPlanAggregate(t *testing.T) {
	router, mock := setupRouterWithMockDB(t)

	reqBody := map[string]any{
		"project": map[string]any{
			"git_id":   1001,
			"name":     "q-demo-project",
			"repo_url": "https://github.com/richer421/q-demo.git",
		},
		"business_unit": map[string]any{
			"name":        "q-demo",
			"description": "demo business unit",
		},
		"ci_config": map[string]any{
			"name":           "q-demo-ci",
			"image_registry": "harbor.local",
			"image_repo":     "q-demo/q-demo",
			"image_tag_rule": map[string]any{"type": "commit"},
			"build_spec": map[string]any{
				"branch": "main",
			},
		},
		"cd_config": map[string]any{
			"name":          "q-demo-cd",
			"render_engine": "helm",
			"values_yaml":   "replicaCount: 1\n",
			"release_strategy": map[string]any{
				"deployment_mode": "rolling",
				"batch_rule": map[string]any{
					"batch_count":  1,
					"batch_ratio":  []float64{1},
					"trigger_type": "auto",
					"interval":     0,
				},
			},
			"git_ops": map[string]any{
				"enabled":       true,
				"repo_url":      "https://github.com/richer421/q-demo-gitops.git",
				"branch":        "main",
				"app_root":      "apps",
				"manifest_root": "manifests",
			},
		},
		"instance_config": map[string]any{
			"name":          "q-demo-dev",
			"env":           "dev",
			"instance_type": "deployment",
			"spec": map[string]any{
				"deployment": map[string]any{},
			},
			"attach_resources": map[string]any{
				"services": map[string]any{
					"q-demo": map[string]any{
						"metadata": map[string]any{"name": "q-demo"},
					},
				},
			},
		},
		"deploy_plan": map[string]any{
			"name":        "q-demo-dev-plan",
			"description": "demo deploy plan",
		},
	}

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `projects`").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO `business_units`").WillReturnResult(sqlmock.NewResult(2, 1))
	mock.ExpectExec("INSERT INTO `ci_configs`").WillReturnResult(sqlmock.NewResult(3, 1))
	mock.ExpectExec("INSERT INTO `cd_configs`").WillReturnResult(sqlmock.NewResult(4, 1))
	mock.ExpectExec("INSERT INTO `instance_configs`").WillReturnResult(sqlmock.NewResult(5, 1))
	mock.ExpectExec("INSERT INTO `deploy_plans`").WillReturnResult(sqlmock.NewResult(6, 1))
	mock.ExpectCommit()

	data, err := json.Marshal(reqBody)
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/deploy-plans", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := decodeBody(t, w.Body)
	assert.Equal(t, float64(0), resp["code"])
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestMetadataAPIGetDeployPlan(t *testing.T) {
	router, mock := setupRouterWithMockDB(t)

	rows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name", "description", "business_unit_id", "ci_config_id", "cd_config_id", "instance_config_id"}).
		AddRow(6, time.Now(), time.Now(), "q-demo-dev-plan", "demo deploy plan", 2, 3, 4, 5)
	mock.ExpectQuery("SELECT \\* FROM `deploy_plans` WHERE `deploy_plans`.`id` = .*").WillReturnRows(rows)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/deploy-plans/6", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := decodeBody(t, w.Body)
	assert.Equal(t, float64(0), resp["code"])
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestMetadataAPIGetDeployPlanFullSpec(t *testing.T) {
	router, mock := setupRouterWithMockDB(t)

	now := time.Now()
	mock.ExpectQuery("SELECT \\* FROM `deploy_plans` WHERE `deploy_plans`.`id` = .*").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "created_at", "updated_at", "name", "description",
			"business_unit_id", "ci_config_id", "cd_config_id", "instance_config_id",
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
	mock.ExpectQuery("SELECT \\* FROM `instance_configs` WHERE `instance_configs`.`id` = .*").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "created_at", "updated_at", "name", "business_unit_id",
			"env", "instance_type", "spec", "attach_resources",
		}).AddRow(5, now, now, "q-demo-dev", 2, "dev", "deployment", []byte(`{"deployment":{}}`), []byte(`{"services":{}}`)))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/deploy-plans/6/full-spec", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := decodeBody(t, w.Body)
	assert.Equal(t, float64(0), resp["code"])
	data := resp["data"].(map[string]any)
	assert.Equal(t, float64(6), data["deploy_plan"].(map[string]any)["id"])
	assert.Equal(t, float64(3), data["ci_config"].(map[string]any)["id"])
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestMetadataAPIGetBusinessUnitFullSpec(t *testing.T) {
	router, mock := setupRouterWithMockDB(t)

	mock.ExpectQuery("SELECT \\* FROM `business_units` WHERE `business_units`.`id` = .*").
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name", "description", "project_id"}).
			AddRow(2, time.Now(), time.Now(), "q-demo", "demo business unit", 1))
	mock.ExpectQuery("SELECT \\* FROM `projects` WHERE `projects`.`id` = .*").
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "git_id", "name", "repo_url"}).
			AddRow(1, time.Now(), time.Now(), 1001, "q-demo-project", "https://github.com/richer421/q-demo.git"))
	mock.ExpectQuery("SELECT \\* FROM `ci_configs` WHERE `ci_configs`.`business_unit_id` = .*").
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name", "business_unit_id", "image_registry", "image_repo", "image_tag_rule", "build_spec"}).
			AddRow(3, time.Now(), time.Now(), "q-demo-ci", 2, "harbor.local", "q-demo/q-demo", []byte(`{"type":"commit"}`), []byte(`{"branch":"main"}`)))
	mock.ExpectQuery("SELECT \\* FROM `cd_configs` WHERE `cd_configs`.`business_unit_id` = .*").
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name", "business_unit_id", "render_engine", "values_yaml", "release_strategy", "git_ops"}).
			AddRow(4, time.Now(), time.Now(), "q-demo-cd", 2, "helm", "replicaCount: 1\n", []byte(`{"deployment_mode":"rolling","batch_rule":{"batch_count":1,"batch_ratio":[1],"trigger_type":"auto","interval":0}}`), []byte(`{"enabled":true,"repo_url":"https://github.com/richer421/q-demo-gitops.git","branch":"main","app_root":"apps","manifest_root":"manifests"}`)))
	mock.ExpectQuery("SELECT \\* FROM `instance_configs` WHERE `instance_configs`.`business_unit_id` = .*").
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name", "business_unit_id", "env", "instance_type", "spec", "attach_resources"}).
			AddRow(5, time.Now(), time.Now(), "q-demo-dev", 2, "dev", "deployment", []byte(`{"deployment":{}}`), []byte(`{"services":{"q-demo":{"metadata":{"name":"q-demo"}}}}`)))
	mock.ExpectQuery("SELECT \\* FROM `deploy_plans` WHERE `deploy_plans`.`business_unit_id` = .*").
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name", "description", "business_unit_id", "ci_config_id", "cd_config_id", "instance_config_id"}).
			AddRow(6, time.Now(), time.Now(), "q-demo-dev-plan", "demo deploy plan", 2, 3, 4, 5))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/business-units/2/full-spec", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := decodeBody(t, w.Body)
	assert.Equal(t, float64(0), resp["code"])
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestMetadataAPISeedDemoSetup(t *testing.T) {
	router, mock := setupRouterWithMockDB(t)

	mock.ExpectQuery("SELECT \\* FROM `projects` WHERE `projects`.`name` = .*").
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "git_id", "name", "repo_url"}).
			AddRow(1, time.Now(), time.Now(), 1001, "q-demo-project", "https://github.com/richer421/q-demo.git"))
	mock.ExpectQuery("SELECT \\* FROM `business_units` WHERE `business_units`.`project_id` = .*").
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name", "description", "project_id"}).
			AddRow(2, time.Now(), time.Now(), "q-demo", "demo business unit", 1))
	mock.ExpectQuery("SELECT \\* FROM `business_units` WHERE `business_units`.`id` = .*").
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name", "description", "project_id"}).
			AddRow(2, time.Now(), time.Now(), "q-demo", "demo business unit", 1))
	mock.ExpectQuery("SELECT \\* FROM `projects` WHERE `projects`.`id` = .*").
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "git_id", "name", "repo_url"}).
			AddRow(1, time.Now(), time.Now(), 1001, "q-demo-project", "https://github.com/richer421/q-demo.git"))
	mock.ExpectQuery("SELECT \\* FROM `ci_configs` WHERE `ci_configs`.`business_unit_id` = .*").
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name", "business_unit_id", "image_registry", "image_repo", "image_tag_rule", "build_spec"}).
			AddRow(3, time.Now(), time.Now(), "q-demo-ci", 2, "harbor.local", "q-demo/q-demo", []byte(`{"type":"commit"}`), []byte(`{"branch":"main"}`)))
	mock.ExpectQuery("SELECT \\* FROM `cd_configs` WHERE `cd_configs`.`business_unit_id` = .*").
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name", "business_unit_id", "render_engine", "values_yaml", "release_strategy", "git_ops"}).
			AddRow(4, time.Now(), time.Now(), "q-demo-cd", 2, "helm", "replicaCount: 1\n", []byte(`{"deployment_mode":"rolling","batch_rule":{"batch_count":1,"batch_ratio":[1],"trigger_type":"auto","interval":0}}`), []byte(`{"enabled":true,"repo_url":"https://github.com/richer421/q-demo-gitops.git","branch":"main","app_root":"apps","manifest_root":"manifests"}`)))
	mock.ExpectQuery("SELECT \\* FROM `instance_configs` WHERE `instance_configs`.`business_unit_id` = .*").
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name", "business_unit_id", "env", "instance_type", "spec", "attach_resources"}).
			AddRow(5, time.Now(), time.Now(), "q-demo-dev", 2, "dev", "deployment", []byte(`{"deployment":{}}`), []byte(`{"services":{"q-demo":{"metadata":{"name":"q-demo"}}}}`)))
	mock.ExpectQuery("SELECT \\* FROM `deploy_plans` WHERE `deploy_plans`.`business_unit_id` = .*").
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name", "description", "business_unit_id", "ci_config_id", "cd_config_id", "instance_config_id"}).
			AddRow(6, time.Now(), time.Now(), "q-demo-dev-plan", "demo deploy plan", 2, 3, 4, 5))
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT \\* FROM `ci_configs` WHERE `ci_configs`.`id` = .*").
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name", "business_unit_id", "image_registry", "image_repo", "image_tag_rule", "build_spec"}).
			AddRow(3, time.Now(), time.Now(), "q-demo-ci", 2, "harbor.local", "q-demo/q-demo", []byte(`{"type":"commit"}`), []byte(`{"branch":"main"}`)))
	mock.ExpectExec("INSERT INTO `ci_configs`").WillReturnResult(sqlmock.NewResult(3, 1))
	mock.ExpectQuery("SELECT \\* FROM `instance_configs` WHERE `instance_configs`.`id` = .*").
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name", "business_unit_id", "env", "instance_type", "spec", "attach_resources"}).
			AddRow(5, time.Now(), time.Now(), "q-demo-dev", 2, "dev", "deployment", []byte(`{"deployment":{}}`), []byte(`{"services":{"q-demo":{"metadata":{"name":"q-demo"}}}}`)))
	mock.ExpectExec("INSERT INTO `instance_configs`").WillReturnResult(sqlmock.NewResult(5, 1))
	mock.ExpectCommit()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/demo/seed", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := decodeBody(t, w.Body)
	assert.Equal(t, float64(0), resp["code"])
	require.NoError(t, mock.ExpectationsWereMet())
}
