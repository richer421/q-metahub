package metadata

import (
	"database/sql/driver"
	"encoding/json"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/richer421/q-metahub/infra/mysql/dao"
	"github.com/richer421/q-metahub/pkg/testutil"
)

func setupMockDAO(t *testing.T) sqlmock.Sqlmock {
	t.Helper()
	db, mock, err := testutil.NewMockDB()
	require.NoError(t, err)
	dao.SetDefault(db)
	return mock
}

func anyTime() driver.Value {
	return sqlmock.AnyArg()
}

func TestServiceGetBusinessUnitFullSpec(t *testing.T) {
	mock := setupMockDAO(t)
	svc := NewService()

	projectRows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "git_id", "name", "repo_url"}).
		AddRow(1, time.Now(), time.Now(), 1001, "q-demo-project", "https://github.com/richer421/q-demo.git")
	businessUnitRows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name", "description", "project_id"}).
		AddRow(2, time.Now(), time.Now(), "q-demo", "demo business unit", 1)

	buildSpec, err := json.Marshal(map[string]any{"branch": "main"})
	require.NoError(t, err)
	tagRule, err := json.Marshal(map[string]any{"type": "commit"})
	require.NoError(t, err)
	ciRows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name", "business_unit_id", "image_registry", "image_repo", "image_tag_rule", "build_spec"}).
		AddRow(3, time.Now(), time.Now(), "q-demo-ci", 2, "harbor.local", "q-demo/q-demo", tagRule, buildSpec)

	releaseStrategy, err := json.Marshal(map[string]any{
		"deployment_mode": "rolling",
		"batch_rule": map[string]any{
			"batch_count":  1,
			"batch_ratio":  []float64{1},
			"trigger_type": "auto",
			"interval":     0,
		},
	})
	require.NoError(t, err)
	gitOps, err := json.Marshal(map[string]any{
		"enabled":       true,
		"repo_url":      "https://github.com/richer421/q-demo-gitops.git",
		"branch":        "main",
		"app_root":      "apps",
		"manifest_root": "manifests",
	})
	require.NoError(t, err)
	cdRows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name", "business_unit_id", "render_engine", "values_yaml", "release_strategy", "git_ops"}).
		AddRow(4, time.Now(), time.Now(), "q-demo-cd", 2, "helm", "replicaCount: 1\n", releaseStrategy, gitOps)

	oamApplication, err := json.Marshal(map[string]any{"component": map[string]any{"type": "pod"}})
	require.NoError(t, err)
	frontendPayload, err := json.Marshal(map[string]any{"basic": map[string]any{"name": "q-demo-dev"}})
	require.NoError(t, err)
	instanceRows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name", "business_unit_id", "env", "schema_version", "oam_application", "frontend_payload"}).
		AddRow(5, time.Now(), time.Now(), "q-demo-dev", 2, "dev", "v1alpha1", oamApplication, frontendPayload)

	deployPlanRows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name", "description", "business_unit_id", "ci_config_id", "cd_config_id", "instance_oam_id"}).
		AddRow(6, time.Now(), time.Now(), "q-demo-dev-plan", "demo deploy plan", 2, 3, 4, 5)

	mock.ExpectQuery("SELECT \\* FROM `business_units` WHERE `business_units`.`id` = .*").
		WillReturnRows(businessUnitRows)
	mock.ExpectQuery("SELECT \\* FROM `projects` WHERE `projects`.`id` = .*").
		WillReturnRows(projectRows)
	mock.ExpectQuery("SELECT \\* FROM `ci_configs` WHERE `ci_configs`.`business_unit_id` = .*").
		WillReturnRows(ciRows)
	mock.ExpectQuery("SELECT \\* FROM `cd_configs` WHERE `cd_configs`.`business_unit_id` = .*").
		WillReturnRows(cdRows)
	mock.ExpectQuery("SELECT \\* FROM `instance_oams` WHERE `instance_oams`.`business_unit_id` = .*").
		WillReturnRows(instanceRows)
	mock.ExpectQuery("SELECT \\* FROM `deploy_plans` WHERE `deploy_plans`.`business_unit_id` = .*").
		WillReturnRows(deployPlanRows)

	res, err := svc.GetBusinessUnitFullSpec(t.Context(), 2)
	require.NoError(t, err)
	require.NotNil(t, res)

	assert.Equal(t, int64(2), res.BusinessUnit.ID)
	assert.Equal(t, "q-demo-project", res.Project.Name)
	require.Len(t, res.CIConfigs, 1)
	require.Len(t, res.CDConfigs, 1)
	require.Len(t, res.InstanceOAMs, 1)
	require.Len(t, res.DeployPlans, 1)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestServiceGetDeployPlanFullSpec(t *testing.T) {
	mock := setupMockDAO(t)
	svc := NewService()

	now := time.Now()
	buildSpec, err := json.Marshal(map[string]any{
		"branch":          "main",
		"dockerfile_path": "./Dockerfile",
		"docker_context":  ".",
		"make_command":    "build",
	})
	require.NoError(t, err)
	tagRule, err := json.Marshal(map[string]any{"type": "commit"})
	require.NoError(t, err)
	releaseStrategy, err := json.Marshal(map[string]any{
		"deployment_mode": "rolling",
		"batch_rule": map[string]any{
			"batch_count":  1,
			"batch_ratio":  []float64{1},
			"trigger_type": "auto",
			"interval":     0,
		},
	})
	require.NoError(t, err)
	gitOps, err := json.Marshal(map[string]any{
		"enabled":       true,
		"repo_url":      "https://github.com/richer421/q-demo-gitops.git",
		"branch":        "main",
		"app_root":      "apps",
		"manifest_root": "manifests",
	})
	require.NoError(t, err)
	oamApplication, err := json.Marshal(map[string]any{"component": map[string]any{"type": "pod"}})
	require.NoError(t, err)
	frontendPayload, err := json.Marshal(map[string]any{"basic": map[string]any{"name": "q-demo-dev"}})
	require.NoError(t, err)

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
		}).AddRow(3, now, now, "q-demo-ci", 2, "harbor.local", "q-demo/q-demo", tagRule, buildSpec))
	mock.ExpectQuery("SELECT \\* FROM `cd_configs` WHERE `cd_configs`.`id` = .*").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "created_at", "updated_at", "name", "business_unit_id",
			"render_engine", "values_yaml", "release_strategy", "git_ops",
		}).AddRow(4, now, now, "q-demo-cd", 2, "helm", "replicaCount: 1\n", releaseStrategy, gitOps))
	mock.ExpectQuery("SELECT \\* FROM `instance_oams` WHERE `instance_oams`.`id` = .*").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "created_at", "updated_at", "name", "business_unit_id",
			"env", "schema_version", "oam_application", "frontend_payload",
		}).AddRow(5, now, now, "q-demo-dev", 2, "dev", "v1alpha1", oamApplication, frontendPayload))

	res, err := svc.GetDeployPlanFullSpec(t.Context(), 6)
	require.NoError(t, err)
	require.NotNil(t, res)

	assert.Equal(t, int64(6), res.DeployPlan.ID)
	assert.Equal(t, int64(2), res.BusinessUnit.ID)
	assert.Equal(t, int64(3), res.CIConfig.ID)
	assert.Equal(t, int64(4), res.CDConfig.ID)
	assert.Equal(t, int64(5), res.InstanceOAM.ID)
	assert.Equal(t, "q-demo-project", res.Project.Name)
	require.NoError(t, mock.ExpectationsWereMet())
}
