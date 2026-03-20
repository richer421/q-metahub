package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/richer421/q-metahub/infra/mysql/dao"
	"github.com/richer421/q-metahub/pkg/testutil"
)

type deployPlanPageResponse struct {
	Items    []deployPlanItemResponse `json:"items"`
	Total    int64                    `json:"total"`
	Page     int                      `json:"page"`
	PageSize int                      `json:"page_size"`
}

type deployPlanItemResponse struct {
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	ID             int64  `json:"id"`
	BusinessUnitID int64  `json:"business_unit_id"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	CIConfigID     int64  `json:"ci_config_id"`
	CDConfigID     int64  `json:"cd_config_id"`
	InstanceOAMID  int64  `json:"instance_oam_id"`
	Env            string `json:"env"`
	CIConfigName   string `json:"ci_config_name"`
	CDConfigName   string `json:"cd_config_name"`
	InstanceName   string `json:"instance_name"`
	LastStatus     string `json:"last_status"`
	LastTime       string `json:"last_time"`
}

func TestListBusinessUnitDeployPlans(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, mock, err := testutil.NewMockDB()
	if err != nil {
		t.Fatalf("new mock db: %v", err)
	}
	dao.SetDefault(db)

	planUpdatedAt := time.Date(2026, 3, 20, 8, 30, 0, 0, time.UTC)
	planRows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "name", "description", "business_unit_id", "ci_config_id", "cd_config_id", "instance_oam_id"}).
		AddRow(21, time.Date(2026, 3, 19, 8, 30, 0, 0, time.UTC), planUpdatedAt, nil, "api-server-prod", "", 1, 11, 12, 13)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `deploy_plans` WHERE `deploy_plans`.`business_unit_id` = ? AND `deploy_plans`.`deleted_at` IS NULL ORDER BY `deploy_plans`.`updated_at` DESC LIMIT ?")).
		WithArgs(1, 10).
		WillReturnRows(planRows)

	ciRows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "name", "business_unit_id", "image_tag_rule", "build_spec"}).
		AddRow(11, time.Now(), time.Now(), nil, "ci-api-server", 1, []byte(`{"type":"branch"}`), []byte(`{"branch":"main"}`))
	mock.ExpectQuery(`SELECT \* FROM ` + "`ci_configs`" + ` WHERE .*` + "`id`" + ` .*\\? AND .*` + "`deleted_at`" + ` IS NULL`).
		WithArgs(11).
		WillReturnRows(ciRows)

	cdRows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "name", "business_unit_id", "release_region", "release_env", "render_engine", "values_yaml", "release_strategy", "git_ops"}).
		AddRow(12, time.Now(), time.Now(), nil, "cd-api-server-prod", 1, "cn-north", "prod", "helm", "", []byte(`{"deployment_mode":"rolling","batch_rule":{"batch_count":1,"batch_ratio":[1],"trigger_type":"auto","interval":0}}`), []byte(`{"enabled":true}`))
	mock.ExpectQuery(`SELECT \* FROM ` + "`cd_configs`" + ` WHERE .*` + "`id`" + ` .*\\? AND .*` + "`deleted_at`" + ` IS NULL`).
		WithArgs(12).
		WillReturnRows(cdRows)

	oamRows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "name", "business_unit_id", "env", "schema_version", "oam_application"}).
		AddRow(13, time.Now(), time.Now(), nil, "inst-api-prod", 1, "prod", "v1alpha1", []byte(`{"apiVersion":"q.oam/v1alpha1","kind":"InstanceApplication","component":{"name":"inst-api-prod","type":"pod","properties":{"mainContainer":{"name":"main"}}}}`))
	mock.ExpectQuery(`SELECT \* FROM ` + "`instance_oams`" + ` WHERE .*` + "`id`" + ` .*\\? AND .*` + "`deleted_at`" + ` IS NULL`).
		WithArgs(13).
		WillReturnRows(oamRows)

	router := gin.New()
	v1 := router.Group("/api/v1")
	v1.GET("/business-units/:id/deploy-plans", ListBusinessUnitDeployPlans)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/business-units/1/deploy-plans?page=1&page_size=10", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", recorder.Code)
	}

	var response testResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if response.Code != 0 {
		t.Fatalf("unexpected response code: %d, body=%s", response.Code, recorder.Body.String())
	}

	var page deployPlanPageResponse
	if err := json.Unmarshal(response.Data, &page); err != nil {
		t.Fatalf("unmarshal data: %v", err)
	}

	if page.Total != 1 || page.Page != 1 || page.PageSize != 10 {
		t.Fatalf("unexpected page metadata: %+v", page)
	}
	if len(page.Items) != 1 {
		t.Fatalf("unexpected items: %+v", page.Items)
	}
	if page.Items[0].Name != "api-server-prod" || page.Items[0].Env != "prod" {
		t.Fatalf("unexpected item core fields: %+v", page.Items[0])
	}
	if page.Items[0].CIConfigName != "ci-api-server" || page.Items[0].CDConfigName != "cd-api-server-prod" || page.Items[0].InstanceName != "inst-api-prod" {
		t.Fatalf("unexpected item relation fields: %+v", page.Items[0])
	}
	if !page.Items[0].UpdatedAt.Equal(planUpdatedAt) {
		t.Fatalf("unexpected item updated_at: %+v", page.Items[0])
	}
	if page.Items[0].LastStatus != "pending" || page.Items[0].LastTime != planUpdatedAt.Format(time.RFC3339) {
		t.Fatalf("unexpected item status fields: %+v", page.Items[0])
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestGetDeployPlan(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, mock, err := testutil.NewMockDB()
	if err != nil {
		t.Fatalf("new mock db: %v", err)
	}
	dao.SetDefault(db)

	planUpdatedAt := time.Date(2026, 3, 20, 9, 10, 0, 0, time.UTC)
	planRows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "name", "description", "business_unit_id", "ci_config_id", "cd_config_id", "instance_oam_id"}).
		AddRow(21, time.Date(2026, 3, 20, 8, 0, 0, 0, time.UTC), planUpdatedAt, nil, "api-server-prod", "prod 发布计划", 1, 11, 12, 13)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `deploy_plans` WHERE `deploy_plans`.`id` = ? AND `deploy_plans`.`deleted_at` IS NULL ORDER BY `deploy_plans`.`id` LIMIT ?")).
		WithArgs(21, 1).
		WillReturnRows(planRows)

	ciRows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "name", "business_unit_id", "image_tag_rule", "build_spec"}).
		AddRow(11, time.Now(), time.Now(), nil, "ci-api-server", 1, []byte(`{"type":"branch"}`), []byte(`{"branch":"main"}`))
	mock.ExpectQuery(`SELECT \* FROM ` + "`ci_configs`" + ` WHERE .*` + "`id`" + ` .*\\? AND .*` + "`deleted_at`" + ` IS NULL`).
		WithArgs(11, 1).
		WillReturnRows(ciRows)

	cdRows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "name", "business_unit_id", "release_region", "release_env", "render_engine", "values_yaml", "release_strategy", "git_ops"}).
		AddRow(12, time.Now(), time.Now(), nil, "cd-api-server-prod", 1, "cn-north", "prod", "helm", "", []byte(`{"deployment_mode":"rolling","batch_rule":{"batch_count":1,"batch_ratio":[1],"trigger_type":"auto","interval":0}}`), []byte(`{"enabled":true}`))
	mock.ExpectQuery(`SELECT \* FROM ` + "`cd_configs`" + ` WHERE .*` + "`id`" + ` .*\\? AND .*` + "`deleted_at`" + ` IS NULL`).
		WithArgs(12, 1).
		WillReturnRows(cdRows)

	oamRows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "name", "business_unit_id", "env", "schema_version", "oam_application"}).
		AddRow(13, time.Now(), time.Now(), nil, "inst-api-prod", 1, "prod", "v1alpha1", []byte(`{"apiVersion":"q.oam/v1alpha1","kind":"InstanceApplication","component":{"name":"inst-api-prod","type":"pod","properties":{"mainContainer":{"name":"main"}}}}`))
	mock.ExpectQuery(`SELECT \* FROM ` + "`instance_oams`" + ` WHERE .*` + "`id`" + ` .*\\? AND .*` + "`deleted_at`" + ` IS NULL`).
		WithArgs(13, 1).
		WillReturnRows(oamRows)

	router := gin.New()
	v1 := router.Group("/api/v1")
	v1.GET("/deploy-plans/:id", GetDeployPlan)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/deploy-plans/21", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", recorder.Code)
	}

	var response testResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if response.Code != 0 {
		t.Fatalf("unexpected response code: %d, body=%s", response.Code, recorder.Body.String())
	}

	var item deployPlanItemResponse
	if err := json.Unmarshal(response.Data, &item); err != nil {
		t.Fatalf("unmarshal data: %v", err)
	}

	if item.ID != 21 || item.Name != "api-server-prod" || item.Description != "prod 发布计划" {
		t.Fatalf("unexpected item fields: %+v", item)
	}
	if item.CIConfigID != 11 || item.CDConfigID != 12 || item.InstanceOAMID != 13 {
		t.Fatalf("unexpected relation ids: %+v", item)
	}
	if item.CIConfigName != "ci-api-server" || item.CDConfigName != "cd-api-server-prod" || item.InstanceName != "inst-api-prod" || item.Env != "prod" {
		t.Fatalf("unexpected relation names: %+v", item)
	}
	if item.LastStatus != "pending" || item.LastTime != planUpdatedAt.Format(time.RFC3339) {
		t.Fatalf("unexpected status fields: %+v", item)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestCreateBusinessUnitDeployPlan(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, mock, err := testutil.NewMockDB()
	if err != nil {
		t.Fatalf("new mock db: %v", err)
	}
	dao.SetDefault(db)

	businessUnitRows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "name", "description", "project_id"}).
		AddRow(1, time.Now(), time.Now(), nil, "api-server", "", 101)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `business_units` WHERE `business_units`.`id` = ? AND `business_units`.`deleted_at` IS NULL ORDER BY `business_units`.`id` LIMIT ?")).
		WithArgs(1, 1).
		WillReturnRows(businessUnitRows)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `deploy_plans` WHERE `deploy_plans`.`business_unit_id` = ? AND `deploy_plans`.`name` = ? AND `deploy_plans`.`deleted_at` IS NULL ORDER BY `deploy_plans`.`id` LIMIT ?")).
		WithArgs(1, "api-server-prod", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	ciRows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "name", "business_unit_id", "image_tag_rule", "build_spec"}).
		AddRow(11, time.Now(), time.Now(), nil, "ci-api-server", 1, []byte(`{"type":"branch"}`), []byte(`{"branch":"main"}`))
	mock.ExpectQuery(`SELECT \* FROM ` + "`ci_configs`" + ` WHERE .*` + "`id`" + ` .*\\? AND .*` + "`deleted_at`" + ` IS NULL`).
		WithArgs(11, 1).
		WillReturnRows(ciRows)

	cdRows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "name", "business_unit_id", "release_region", "release_env", "render_engine", "values_yaml", "release_strategy", "git_ops"}).
		AddRow(12, time.Now(), time.Now(), nil, "cd-api-server-prod", 1, "cn-north", "prod", "helm", "", []byte(`{"deployment_mode":"rolling","batch_rule":{"batch_count":1,"batch_ratio":[1],"trigger_type":"auto","interval":0}}`), []byte(`{"enabled":true}`))
	mock.ExpectQuery(`SELECT \* FROM ` + "`cd_configs`" + ` WHERE .*` + "`id`" + ` .*\\? AND .*` + "`deleted_at`" + ` IS NULL`).
		WithArgs(12, 1).
		WillReturnRows(cdRows)

	oamRows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "name", "business_unit_id", "env", "schema_version", "oam_application"}).
		AddRow(13, time.Now(), time.Now(), nil, "inst-api-prod", 1, "prod", "v1alpha1", []byte(`{"apiVersion":"q.oam/v1alpha1","kind":"InstanceApplication","component":{"name":"inst-api-prod","type":"pod","properties":{"mainContainer":{"name":"main"}}}}`))
	mock.ExpectQuery(`SELECT \* FROM ` + "`instance_oams`" + ` WHERE .*` + "`id`" + ` .*\\? AND .*` + "`deleted_at`" + ` IS NULL`).
		WithArgs(13, 1).
		WillReturnRows(oamRows)

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `deploy_plans`")).
		WithArgs(
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			"api-server-prod",
			"prod 发布计划",
			int64(1),
			int64(11),
			int64(12),
			int64(13),
		).
		WillReturnResult(sqlmock.NewResult(21, 1))

	router := gin.New()
	v1 := router.Group("/api/v1")
	v1.POST("/business-units/:id/deploy-plans", CreateBusinessUnitDeployPlan)

	body := bytes.NewBufferString(`{"name":"api-server-prod","description":"prod 发布计划","ci_config_id":11,"cd_config_id":12,"instance_oam_id":13}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/business-units/1/deploy-plans", body)
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", recorder.Code)
	}

	var response testResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if response.Code != 0 {
		t.Fatalf("unexpected response code: %d, body=%s", response.Code, recorder.Body.String())
	}

	var item deployPlanItemResponse
	if err := json.Unmarshal(response.Data, &item); err != nil {
		t.Fatalf("unmarshal data: %v", err)
	}
	if item.ID != 21 || item.Name != "api-server-prod" || item.Description != "prod 发布计划" {
		t.Fatalf("unexpected created item: %+v", item)
	}
	if item.CIConfigID != 11 || item.CDConfigID != 12 || item.InstanceOAMID != 13 {
		t.Fatalf("unexpected relation ids: %+v", item)
	}
	if item.Env != "prod" || item.CIConfigName != "ci-api-server" || item.CDConfigName != "cd-api-server-prod" || item.InstanceName != "inst-api-prod" {
		t.Fatalf("unexpected relation names: %+v", item)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestUpdateDeployPlan(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, mock, err := testutil.NewMockDB()
	if err != nil {
		t.Fatalf("new mock db: %v", err)
	}
	dao.SetDefault(db)

	createdAt := time.Date(2026, 3, 20, 8, 0, 0, 0, time.UTC)
	currentUpdatedAt := time.Date(2026, 3, 20, 9, 0, 0, 0, time.UTC)
	currentRows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "name", "description", "business_unit_id", "ci_config_id", "cd_config_id", "instance_oam_id"}).
		AddRow(21, createdAt, currentUpdatedAt, nil, "api-server-prod", "old", 1, 11, 12, 13)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `deploy_plans` WHERE `deploy_plans`.`id` = ? AND `deploy_plans`.`deleted_at` IS NULL ORDER BY `deploy_plans`.`id` LIMIT ?")).
		WithArgs(21, 1).
		WillReturnRows(currentRows)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `deploy_plans` WHERE `deploy_plans`.`business_unit_id` = ? AND `deploy_plans`.`name` = ? AND `deploy_plans`.`id` <> ? AND `deploy_plans`.`deleted_at` IS NULL ORDER BY `deploy_plans`.`id` LIMIT ?")).
		WithArgs(1, "api-server-prod-v2", 21, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	ciRows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "name", "business_unit_id", "image_tag_rule", "build_spec"}).
		AddRow(14, time.Now(), time.Now(), nil, "ci-api-server-v2", 1, []byte(`{"type":"branch"}`), []byte(`{"branch":"main"}`))
	mock.ExpectQuery(`SELECT \* FROM ` + "`ci_configs`" + ` WHERE .*` + "`id`" + ` .*\\? AND .*` + "`deleted_at`" + ` IS NULL`).
		WithArgs(14, 1).
		WillReturnRows(ciRows)

	cdRows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "name", "business_unit_id", "release_region", "release_env", "render_engine", "values_yaml", "release_strategy", "git_ops"}).
		AddRow(15, time.Now(), time.Now(), nil, "cd-api-server-gray", 1, "cn-north", "gray", "helm", "", []byte(`{"deployment_mode":"rolling","batch_rule":{"batch_count":1,"batch_ratio":[1],"trigger_type":"auto","interval":0}}`), []byte(`{"enabled":true}`))
	mock.ExpectQuery(`SELECT \* FROM ` + "`cd_configs`" + ` WHERE .*` + "`id`" + ` .*\\? AND .*` + "`deleted_at`" + ` IS NULL`).
		WithArgs(15, 1).
		WillReturnRows(cdRows)

	oamRows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "name", "business_unit_id", "env", "schema_version", "oam_application"}).
		AddRow(16, time.Now(), time.Now(), nil, "inst-api-gray", 1, "gray", "v1alpha1", []byte(`{"apiVersion":"q.oam/v1alpha1","kind":"InstanceApplication","component":{"name":"inst-api-gray","type":"pod","properties":{"mainContainer":{"name":"main"}}}}`))
	mock.ExpectQuery(`SELECT \* FROM ` + "`instance_oams`" + ` WHERE .*` + "`id`" + ` .*\\? AND .*` + "`deleted_at`" + ` IS NULL`).
		WithArgs(16, 1).
		WillReturnRows(oamRows)

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `deploy_plans`")).
		WillReturnResult(sqlmock.NewResult(0, 1))

	router := gin.New()
	v1 := router.Group("/api/v1")
	v1.PUT("/deploy-plans/:id", UpdateDeployPlan)

	body := bytes.NewBufferString(`{"name":"api-server-prod-v2","description":"gray 发布计划","ci_config_id":14,"cd_config_id":15,"instance_oam_id":16}`)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/deploy-plans/21", body)
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", recorder.Code)
	}

	var response testResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if response.Code != 0 {
		t.Fatalf("unexpected response code: %d, body=%s", response.Code, recorder.Body.String())
	}

	var item deployPlanItemResponse
	if err := json.Unmarshal(response.Data, &item); err != nil {
		t.Fatalf("unmarshal data: %v", err)
	}
	if item.ID != 21 || item.Name != "api-server-prod-v2" || item.Description != "gray 发布计划" {
		t.Fatalf("unexpected updated item: %+v", item)
	}
	if item.CIConfigID != 14 || item.CDConfigID != 15 || item.InstanceOAMID != 16 {
		t.Fatalf("unexpected updated relation ids: %+v", item)
	}
	if item.Env != "gray" || item.InstanceName != "inst-api-gray" {
		t.Fatalf("unexpected updated relation names: %+v", item)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestDeleteDeployPlan(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, mock, err := testutil.NewMockDB()
	if err != nil {
		t.Fatalf("new mock db: %v", err)
	}
	dao.SetDefault(db)

	currentRows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "name", "description", "business_unit_id", "ci_config_id", "cd_config_id", "instance_oam_id"}).
		AddRow(21, time.Now(), time.Now(), nil, "api-server-prod", "", 1, 11, 12, 13)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `deploy_plans` WHERE `deploy_plans`.`id` = ? AND `deploy_plans`.`deleted_at` IS NULL ORDER BY `deploy_plans`.`id` LIMIT ?")).
		WithArgs(21, 1).
		WillReturnRows(currentRows)
	mock.ExpectExec("UPDATE `deploy_plans` SET `deleted_at`=\\? WHERE `deploy_plans`.`id` = \\? AND `deploy_plans`.`deleted_at` IS NULL").
		WithArgs(sqlmock.AnyArg(), 21).
		WillReturnResult(sqlmock.NewResult(0, 1))

	router := gin.New()
	v1 := router.Group("/api/v1")
	v1.DELETE("/deploy-plans/:id", DeleteDeployPlan)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/deploy-plans/21", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", recorder.Code)
	}

	var response testResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if response.Code != 0 {
		t.Fatalf("unexpected response code: %d, body=%s", response.Code, recorder.Body.String())
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
