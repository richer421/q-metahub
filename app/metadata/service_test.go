package metadata

import (
	"database/sql/driver"
	"encoding/json"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/richer421/q-metahub/app/metadata/vo"
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

func TestServiceCreateDeployPlanAggregate(t *testing.T) {
	mock := setupMockDAO(t)
	svc := NewService()

	now := time.Now()
	createReq := &vo.CreateDeployPlanAggregateReq{
		Project: vo.CreateProjectReq{
			GitID:   1001,
			Name:    "q-demo-project",
			RepoURL: "https://github.com/richer421/q-demo.git",
		},
		BusinessUnit: vo.CreateBusinessUnitReq{
			Name:        "q-demo",
			Description: "demo business unit",
		},
		CIConfig: vo.CreateCIConfigReq{
			Name:          "q-demo-ci",
			ImageRegistry: "harbor.local",
			ImageRepo:     "q-demo/q-demo",
			ImageTagRule: map[string]any{
				"type": "commit",
			},
			BuildSpec: map[string]any{
				"branch":          "main",
				"dockerfile_path": "./Dockerfile",
				"docker_context":  ".",
				"make_command":    "build",
			},
		},
		CDConfig: vo.CreateCDConfigReq{
			Name:         "q-demo-cd",
			RenderEngine: "helm",
			ValuesYAML:   "replicaCount: 1\n",
			ReleaseStrategy: map[string]any{
				"deployment_mode": "rolling",
				"batch_rule": map[string]any{
					"batch_count":  1,
					"batch_ratio":  []float64{1},
					"trigger_type": "auto",
					"interval":     0,
				},
			},
			GitOps: map[string]any{
				"enabled":       true,
				"repo_url":      "https://github.com/richer421/q-demo-gitops.git",
				"branch":        "main",
				"app_root":      "apps",
				"manifest_root": "manifests",
			},
		},
		InstanceConfig: vo.CreateInstanceConfigReq{
			Name:          "q-demo-dev",
			Env:           "dev",
			SchemaVersion: "v1alpha1",
			OAMApplication: map[string]any{
				"apiVersion": "q.oam/v1alpha1",
				"kind":       "InstanceApplication",
				"component": map[string]any{
					"name": "q-demo",
					"type": "pod",
					"properties": map[string]any{
						"mainContainer": map[string]any{
							"name":  "q-demo",
							"image": "IMAGE",
						},
					},
				},
			},
			FrontendPayload: map[string]any{
				"basic": map[string]any{
					"name": "q-demo-dev",
				},
			},
		},
		DeployPlan: vo.CreateDeployPlanReq{
			Name:        "q-demo-dev-plan",
			Description: "demo deploy plan",
		},
	}

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `projects`").
		WithArgs(anyTime(), anyTime(), createReq.Project.GitID, createReq.Project.Name, createReq.Project.RepoURL).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO `business_units`").
		WithArgs(anyTime(), anyTime(), createReq.BusinessUnit.Name, createReq.BusinessUnit.Description, int64(1)).
		WillReturnResult(sqlmock.NewResult(2, 1))
	mock.ExpectExec("INSERT INTO `ci_configs`").
		WithArgs(anyTime(), anyTime(), createReq.CIConfig.Name, int64(2), createReq.CIConfig.ImageRegistry, createReq.CIConfig.ImageRepo, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(3, 1))
	mock.ExpectExec("INSERT INTO `cd_configs`").
		WithArgs(anyTime(), anyTime(), createReq.CDConfig.Name, int64(2), createReq.CDConfig.RenderEngine, createReq.CDConfig.ValuesYAML, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(4, 1))
	mock.ExpectExec("INSERT INTO `instance_oams`").
		WithArgs(anyTime(), anyTime(), createReq.InstanceConfig.Name, int64(2), createReq.InstanceConfig.Env, createReq.InstanceConfig.SchemaVersion, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(5, 1))
	mock.ExpectExec("INSERT INTO `deploy_plans`").
		WithArgs(anyTime(), anyTime(), createReq.DeployPlan.Name, createReq.DeployPlan.Description, int64(2), int64(3), int64(4), int64(5)).
		WillReturnResult(sqlmock.NewResult(6, 1))
	mock.ExpectCommit()

	res, err := svc.CreateDeployPlanAggregate(t.Context(), createReq)
	require.NoError(t, err)
	require.NotNil(t, res)

	assert.Equal(t, int64(1), res.Project.ID)
	assert.Equal(t, int64(2), res.BusinessUnit.ID)
	assert.Equal(t, int64(6), res.DeployPlan.ID)
	assert.WithinDuration(t, now, res.Project.CreatedAt, time.Minute)
	require.NoError(t, mock.ExpectationsWereMet())
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

	deployPlanRows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name", "description", "business_unit_id", "ci_config_id", "cd_config_id", "instance_config_id"}).
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
	require.Len(t, res.InstanceConfigs, 1)
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
	assert.Equal(t, int64(5), res.InstanceConfig.ID)
	assert.Equal(t, "q-demo-project", res.Project.Name)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestServiceSeedDemoSetupIsIdempotent(t *testing.T) {
	mock := setupMockDAO(t)
	svc := NewService()

	projectRows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "git_id", "name", "repo_url"}).
		AddRow(1, time.Now(), time.Now(), 1001, "q-demo-project", "https://github.com/richer421/q-demo.git")
	businessUnitRows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name", "description", "project_id"}).
		AddRow(2, time.Now(), time.Now(), "q-demo", "demo business unit", 1)
	ciRows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name", "business_unit_id", "image_registry", "image_repo", "image_tag_rule", "build_spec"}).
		AddRow(3, time.Now(), time.Now(), "q-demo-ci", 2, "harbor.local", "q-demo/q-demo", []byte(`{"type":"commit"}`), []byte(`{"branch":"main"}`))
	cdRows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name", "business_unit_id", "render_engine", "values_yaml", "release_strategy", "git_ops"}).
		AddRow(4, time.Now(), time.Now(), "q-demo-cd", 2, "helm", "replicaCount: 1\n", []byte(`{"deployment_mode":"rolling","batch_rule":{"batch_count":1,"batch_ratio":[1],"trigger_type":"auto","interval":0}}`), []byte(`{"enabled":true,"repo_url":"https://github.com/richer421/q-demo-gitops.git","branch":"main","app_root":"apps","manifest_root":"manifests"}`))
	instanceRows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name", "business_unit_id", "env", "schema_version", "oam_application", "frontend_payload"}).
		AddRow(5, time.Now(), time.Now(), "q-demo-dev", 2, "dev", "v1alpha1", []byte(`{"component":{}}`), []byte(`{"basic":{"name":"q-demo-dev"}}`))
	deployPlanRows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name", "description", "business_unit_id", "ci_config_id", "cd_config_id", "instance_config_id"}).
		AddRow(6, time.Now(), time.Now(), "q-demo-dev-plan", "demo deploy plan", 2, 3, 4, 5)

	mock.ExpectQuery("SELECT \\* FROM `projects` WHERE `projects`.`name` = .*").
		WillReturnRows(projectRows)
	mock.ExpectQuery("SELECT \\* FROM `business_units` WHERE `business_units`.`project_id` = .*").
		WillReturnRows(businessUnitRows)
	mock.ExpectQuery("SELECT \\* FROM `business_units` WHERE `business_units`.`id` = .*").
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name", "description", "project_id"}).
			AddRow(2, time.Now(), time.Now(), "q-demo", "demo business unit", 1))
	mock.ExpectQuery("SELECT \\* FROM `projects` WHERE `projects`.`id` = .*").
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "git_id", "name", "repo_url"}).
			AddRow(1, time.Now(), time.Now(), 1001, "q-demo-project", "https://github.com/richer421/q-demo.git"))
	mock.ExpectQuery("SELECT \\* FROM `ci_configs` WHERE `ci_configs`.`business_unit_id` = .*").
		WillReturnRows(ciRows)
	mock.ExpectQuery("SELECT \\* FROM `cd_configs` WHERE `cd_configs`.`business_unit_id` = .*").
		WillReturnRows(cdRows)
	mock.ExpectQuery("SELECT \\* FROM `instance_oams` WHERE `instance_oams`.`business_unit_id` = .*").
		WillReturnRows(instanceRows)
	mock.ExpectQuery("SELECT \\* FROM `deploy_plans` WHERE `deploy_plans`.`business_unit_id` = .*").
		WillReturnRows(deployPlanRows)
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT \\* FROM `ci_configs` WHERE `ci_configs`.`id` = .*").
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name", "business_unit_id", "image_registry", "image_repo", "image_tag_rule", "build_spec"}).
			AddRow(3, time.Now(), time.Now(), "q-demo-ci", 2, "harbor.local", "q-demo/q-demo", []byte(`{"type":"commit"}`), []byte(`{"branch":"main"}`)))
	mock.ExpectExec("INSERT INTO `ci_configs`").
		WithArgs(anyTime(), anyTime(), "q-demo-ci", int64(2), defaultDemoImageRegistry, "q-demo/q-demo", sqlmock.AnyArg(), sqlmock.AnyArg(), int64(3), anyTime()).
		WillReturnResult(sqlmock.NewResult(3, 1))
	mock.ExpectQuery("SELECT \\* FROM `instance_oams` WHERE `instance_oams`.`id` = .*").
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name", "business_unit_id", "env", "schema_version", "oam_application", "frontend_payload"}).
			AddRow(5, time.Now(), time.Now(), "q-demo-dev", 2, "dev", "v1alpha1", []byte(`{"component":{}}`), []byte(`{"basic":{"name":"q-demo-dev"}}`)))
	mock.ExpectExec("INSERT INTO `instance_oams`").
		WithArgs(anyTime(), anyTime(), "q-demo-dev", int64(2), "dev", "v1alpha1", sqlmock.AnyArg(), sqlmock.AnyArg(), int64(5), anyTime()).
		WillReturnResult(sqlmock.NewResult(5, 1))
	mock.ExpectCommit()

	res, err := svc.SeedDemoSetup(t.Context())
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, int64(6), res.DeployPlan.ID)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestServiceSeedDemoSetupReconcilesDemoCIRegistry(t *testing.T) {
	mock := setupMockDAO(t)
	svc := NewService()

	now := time.Now()
	projectRows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "git_id", "name", "repo_url"}).
		AddRow(1, now, now, 1001, "q-demo-project", "https://github.com/richer421/q-demo.git")
	businessUnitRows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name", "description", "project_id"}).
		AddRow(2, now, now, "q-demo", "demo business unit", 1)
	ciRows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name", "business_unit_id", "image_registry", "image_repo", "image_tag_rule", "build_spec"}).
		AddRow(3, now, now, "q-demo-ci", 2, "harbor.local", "q-demo/q-demo", []byte(`{"type":"commit"}`), []byte(`{"branch":"main"}`))
	cdRows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name", "business_unit_id", "render_engine", "values_yaml", "release_strategy", "git_ops"}).
		AddRow(4, now, now, "q-demo-cd", 2, "helm", "replicaCount: 1\n", []byte(`{"deployment_mode":"rolling","batch_rule":{"batch_count":1,"batch_ratio":[1],"trigger_type":"auto","interval":0}}`), []byte(`{"enabled":true,"repo_url":"https://github.com/richer421/q-demo-gitops.git","branch":"main","app_root":"apps","manifest_root":"manifests"}`))
	instanceRows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name", "business_unit_id", "env", "schema_version", "oam_application", "frontend_payload"}).
		AddRow(5, now, now, "q-demo-dev", 2, "dev", "v1alpha1", []byte(`{"component":{}}`), []byte(`{"basic":{"name":"q-demo-dev"}}`))
	deployPlanRows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name", "description", "business_unit_id", "ci_config_id", "cd_config_id", "instance_config_id"}).
		AddRow(6, now, now, "q-demo-dev-plan", "demo deploy plan", 2, 3, 4, 5)

	mock.ExpectQuery("SELECT \\* FROM `projects` WHERE `projects`.`name` = .*").
		WillReturnRows(projectRows)
	mock.ExpectQuery("SELECT \\* FROM `business_units` WHERE `business_units`.`project_id` = .*").
		WillReturnRows(businessUnitRows)
	mock.ExpectQuery("SELECT \\* FROM `business_units` WHERE `business_units`.`id` = .*").
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name", "description", "project_id"}).
			AddRow(2, now, now, "q-demo", "demo business unit", 1))
	mock.ExpectQuery("SELECT \\* FROM `projects` WHERE `projects`.`id` = .*").
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "git_id", "name", "repo_url"}).
			AddRow(1, now, now, 1001, "q-demo-project", "https://github.com/richer421/q-demo.git"))
	mock.ExpectQuery("SELECT \\* FROM `ci_configs` WHERE `ci_configs`.`business_unit_id` = .*").
		WillReturnRows(ciRows)
	mock.ExpectQuery("SELECT \\* FROM `cd_configs` WHERE `cd_configs`.`business_unit_id` = .*").
		WillReturnRows(cdRows)
	mock.ExpectQuery("SELECT \\* FROM `instance_oams` WHERE `instance_oams`.`business_unit_id` = .*").
		WillReturnRows(instanceRows)
	mock.ExpectQuery("SELECT \\* FROM `deploy_plans` WHERE `deploy_plans`.`business_unit_id` = .*").
		WillReturnRows(deployPlanRows)
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT \\* FROM `ci_configs` WHERE `ci_configs`.`id` = .*").
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name", "business_unit_id", "image_registry", "image_repo", "image_tag_rule", "build_spec"}).
			AddRow(3, now, now, "q-demo-ci", 2, "harbor.local", "q-demo/q-demo", []byte(`{"type":"commit"}`), []byte(`{"branch":"main"}`)))
	mock.ExpectExec("INSERT INTO `ci_configs`").
		WithArgs(anyTime(), anyTime(), "q-demo-ci", int64(2), defaultDemoImageRegistry, "q-demo/q-demo", sqlmock.AnyArg(), sqlmock.AnyArg(), int64(3), anyTime()).
		WillReturnResult(sqlmock.NewResult(3, 1))
	mock.ExpectQuery("SELECT \\* FROM `instance_oams` WHERE `instance_oams`.`id` = .*").
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name", "business_unit_id", "env", "schema_version", "oam_application", "frontend_payload"}).
			AddRow(5, now, now, "q-demo-dev", 2, "dev", "v1alpha1", []byte(`{"component":{}}`), []byte(`{"basic":{"name":"q-demo-dev"}}`)))
	mock.ExpectExec("INSERT INTO `instance_oams`").
		WithArgs(anyTime(), anyTime(), "q-demo-dev", int64(2), "dev", "v1alpha1", sqlmock.AnyArg(), sqlmock.AnyArg(), int64(5), anyTime()).
		WillReturnResult(sqlmock.NewResult(5, 1))
	mock.ExpectCommit()

	res, err := svc.SeedDemoSetup(t.Context())
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, defaultDemoImageRegistry, res.CIConfig.ImageRegistry)
	require.NoError(t, mock.ExpectationsWereMet())
}
