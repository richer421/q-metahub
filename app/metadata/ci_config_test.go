package metadata

import (
	"context"
	"database/sql/driver"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/richer421/q-metahub/app/metadata/vo"
	"github.com/richer421/q-metahub/infra/mysql/dao"
	"github.com/richer421/q-metahub/infra/mysql/model"
	"github.com/richer421/q-metahub/pkg/testutil"
)

func TestToCIConfigVOAppliesBuildSpecDefaults(t *testing.T) {
	createdAt := time.Date(2026, time.March, 18, 10, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2026, time.March, 18, 12, 0, 0, 0, time.UTC)

	entity := &model.CIConfig{
		BaseModel: model.BaseModel{
			ID:        12,
			CreatedAt: createdAt,
			UpdatedAt: updatedAt,
		},
		Name:           "api-server",
		BusinessUnitID: 34,
		ImageRegistry:  "harbor.example.com/project-a",
		ImageRepo:      "api-server",
		ImageTagRule: model.ImageTagRule{
			Type:          "branch",
			WithTimestamp: true,
			WithCommit:    true,
		},
		BuildSpec: model.BuildSpec{},
	}

	item := toCIConfigVO(entity)
	item.DeployPlanRefCount = 2

	if item.ID != entity.ID {
		t.Fatalf("expected id %d, got %d", entity.ID, item.ID)
	}
	if item.CreatedAt != createdAt {
		t.Fatalf("expected created_at %v, got %v", createdAt, item.CreatedAt)
	}
	if item.UpdatedAt != updatedAt {
		t.Fatalf("expected updated_at %v, got %v", updatedAt, item.UpdatedAt)
	}
	if item.DeployPlanRefCount != 2 {
		t.Fatalf("expected deploy plan ref count 2, got %d", item.DeployPlanRefCount)
	}
	if item.BuildSpec.MakefilePath != "./Makefile" {
		t.Fatalf("expected default makefile path, got %q", item.BuildSpec.MakefilePath)
	}
	if item.BuildSpec.DockerfilePath != "./Dockerfile" {
		t.Fatalf("expected default dockerfile path, got %q", item.BuildSpec.DockerfilePath)
	}
}

func TestListBusinessUnitCIConfigsFiltersByName(t *testing.T) {
	db, mock, err := testutil.NewMockDB()
	if err != nil {
		t.Fatalf("create mock db: %v", err)
	}
	dao.SetDefault(db)

	rows := sqlmock.NewRows([]string{
		"id", "created_at", "updated_at", "name", "business_unit_id", "image_registry", "image_repo", "image_tag_rule", "build_spec",
	}).AddRow(
		1,
		time.Date(2026, time.March, 18, 10, 0, 0, 0, time.UTC),
		time.Date(2026, time.March, 18, 12, 0, 0, 0, time.UTC),
		"api-server",
		34,
		"harbor.example.com/project-a",
		"api-server",
		jsonValue(t, `{"type":"branch","with_timestamp":true,"with_commit":true}`),
		jsonValue(t, `{"makefile_path":"./Makefile","dockerfile_path":"./Dockerfile"}`),
	)

	mock.ExpectQuery("SELECT \\* FROM `ci_configs` WHERE .*`business_unit_id` = \\? AND .*`name` LIKE \\?.* ORDER BY .*`updated_at` DESC LIMIT \\?").
		WithArgs(34, "%api%", 10).
		WillReturnRows(rows)

	page, err := App.ListBusinessUnitCIConfigs(context.Background(), 34, 1, 10, "api")
	if err != nil {
		t.Fatalf("list business unit ci configs: %v", err)
	}
	if page.Total != 1 {
		t.Fatalf("expected total 1, got %d", page.Total)
	}
	if page.Page != 1 || page.PageSize != 10 {
		t.Fatalf("expected page=1 page_size=10, got page=%d page_size=%d", page.Page, page.PageSize)
	}
	if len(page.Items) != 1 {
		t.Fatalf("expected one item, got %d", len(page.Items))
	}
	if page.Items[0].Name != "api-server" {
		t.Fatalf("expected filtered item name api-server, got %q", page.Items[0].Name)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestNormalizeCreateCIConfigAppliesDefaults(t *testing.T) {
	req := vo.CreateCIConfigReq{
		Name: "  API-SERVER  ",
		ImageTagRule: vo.CIConfigImageTagRuleVO{
			Type: "branch",
		},
		BuildSpec: vo.CIConfigBuildSpecVO{},
	}

	entity, err := normalizeCreateCIConfig(34, req, "API Server")
	if err != nil {
		t.Fatalf("normalize create ci config: %v", err)
	}

	if entity.Name != "API-SERVER" {
		t.Fatalf("expected trimmed name, got %q", entity.Name)
	}
	if entity.ImageRegistry != "api-server" {
		t.Fatalf("expected default image registry, got %q", entity.ImageRegistry)
	}
	if entity.ImageRepo != "api-server" {
		t.Fatalf("expected generated image repo, got %q", entity.ImageRepo)
	}
	if entity.BuildSpec.MakefilePath != "./Makefile" {
		t.Fatalf("expected default makefile path, got %q", entity.BuildSpec.MakefilePath)
	}
	if entity.BuildSpec.MakeCommand != "make build" {
		t.Fatalf("expected default make command, got %q", entity.BuildSpec.MakeCommand)
	}
	if entity.BuildSpec.DockerfilePath != "./Dockerfile" {
		t.Fatalf("expected default dockerfile path, got %q", entity.BuildSpec.DockerfilePath)
	}
	if entity.BuildSpec.DockerContext != "." {
		t.Fatalf("expected default docker context, got %q", entity.BuildSpec.DockerContext)
	}
	if entity.BuildSpec.BuildArgs == nil || len(entity.BuildSpec.BuildArgs) != 0 {
		t.Fatalf("expected empty build args map, got %+v", entity.BuildSpec.BuildArgs)
	}
}

func TestNormalizeCreateCIConfigRejectsInvalidCustomTemplate(t *testing.T) {
	req := vo.CreateCIConfigReq{
		Name: "api-server",
		ImageTagRule: vo.CIConfigImageTagRuleVO{
			Type:     "custom",
			Template: "release/${foo}",
		},
	}

	_, err := normalizeCreateCIConfig(34, req, "api-server")
	if err == nil || !strings.Contains(err.Error(), "invalid image tag rule template") {
		t.Fatalf("expected invalid image tag rule template error, got %v", err)
	}
}

func TestCreateBusinessUnitCIConfigRejectsMissingBusinessUnit(t *testing.T) {
	db, mock, err := testutil.NewMockDB()
	if err != nil {
		t.Fatalf("create mock db: %v", err)
	}
	dao.SetDefault(db)

	mock.ExpectQuery("SELECT \\* FROM `business_units` WHERE .*`id` = \\?.* ORDER BY .*`id` LIMIT \\?").
		WithArgs(34, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}))

	_, err = App.CreateBusinessUnitCIConfig(context.Background(), 34, vo.CreateCIConfigReq{
		Name: "api-server",
		ImageTagRule: vo.CIConfigImageTagRuleVO{
			Type: "branch",
		},
		BuildSpec: vo.CIConfigBuildSpecVO{},
	})
	if err == nil || !strings.Contains(err.Error(), "business unit not found") {
		t.Fatalf("expected business unit not found error, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestMergeCIConfigUpdatePreservesHiddenBuildSpecFields(t *testing.T) {
	current := &model.CIConfig{
		Name:          "api-server",
		ImageRegistry: "harbor.example.com/project-a",
		ImageRepo:     "api-server",
		ImageTagRule: model.ImageTagRule{
			Type:          "branch",
			WithTimestamp: true,
			WithCommit:    true,
		},
		BuildSpec: model.BuildSpec{
			Branch:         stringPtr("main"),
			MakefilePath:   "./ops/Makefile",
			MakeCommand:    "build",
			DockerfilePath: "./deploy/Dockerfile",
			DockerContext:  ".",
			BuildArgs: map[string]string{
				"GO_ENV": "prod",
			},
		},
	}

	req := vo.UpdateCIConfigReq{
		Name: stringPtr("api-server-v2"),
		BuildSpec: &vo.CIConfigBuildSpecVO{
			MakefilePath: "./Makefile",
		},
	}

	next, err := mergeCIConfigUpdate(current, req)
	if err != nil {
		t.Fatalf("merge ci config update: %v", err)
	}

	if next.Name != "api-server-v2" {
		t.Fatalf("expected updated name, got %q", next.Name)
	}
	if next.ImageRepo != "api-server" {
		t.Fatalf("expected image repo unchanged, got %q", next.ImageRepo)
	}
	if next.BuildSpec.MakefilePath != "./Makefile" {
		t.Fatalf("expected overridden makefile path, got %q", next.BuildSpec.MakefilePath)
	}
	if next.BuildSpec.DockerfilePath != "./deploy/Dockerfile" {
		t.Fatalf("expected dockerfile path preserved, got %q", next.BuildSpec.DockerfilePath)
	}
	if next.BuildSpec.MakeCommand != "build" {
		t.Fatalf("expected make command preserved, got %q", next.BuildSpec.MakeCommand)
	}
	if next.BuildSpec.Branch == nil || *next.BuildSpec.Branch != "main" {
		t.Fatalf("expected branch preserved, got %+v", next.BuildSpec.Branch)
	}
	if next.BuildSpec.BuildArgs["GO_ENV"] != "prod" {
		t.Fatalf("expected build args preserved, got %+v", next.BuildSpec.BuildArgs)
	}
}

func TestUpdateCIConfigRejectsDuplicateNameInBusinessUnit(t *testing.T) {
	db, mock, err := testutil.NewMockDB()
	if err != nil {
		t.Fatalf("create mock db: %v", err)
	}
	dao.SetDefault(db)

	currentRows := sqlmock.NewRows([]string{
		"id", "created_at", "updated_at", "name", "business_unit_id", "image_registry", "image_repo", "image_tag_rule", "build_spec",
	}).AddRow(
		12,
		time.Date(2026, time.March, 18, 10, 0, 0, 0, time.UTC),
		time.Date(2026, time.March, 18, 12, 0, 0, 0, time.UTC),
		"api-server",
		34,
		"harbor.example.com/project-a",
		"api-server",
		jsonValue(t, `{"type":"branch"}`),
		jsonValue(t, `{}`),
	)
	duplicateRows := sqlmock.NewRows([]string{
		"id", "created_at", "updated_at", "name", "business_unit_id", "image_registry", "image_repo", "image_tag_rule", "build_spec",
	}).AddRow(
		15,
		time.Date(2026, time.March, 18, 11, 0, 0, 0, time.UTC),
		time.Date(2026, time.March, 18, 13, 0, 0, 0, time.UTC),
		"api-server-v2",
		34,
		"harbor.example.com/project-a",
		"api-server",
		jsonValue(t, `{"type":"branch"}`),
		jsonValue(t, `{}`),
	)

	mock.ExpectQuery("SELECT \\* FROM `ci_configs` WHERE .*`id` = \\?.* ORDER BY .*`id` LIMIT \\?").
		WithArgs(12, 1).
		WillReturnRows(currentRows)
	mock.ExpectQuery("SELECT \\* FROM `ci_configs` WHERE .*`business_unit_id` = \\? AND .*`name` = \\? AND .*`id` <> \\?.* ORDER BY .*`id` LIMIT \\?").
		WithArgs(34, "api-server-v2", 12, 1).
		WillReturnRows(duplicateRows)

	_, err = App.UpdateCIConfig(context.Background(), 12, vo.UpdateCIConfigReq{
		Name: stringPtr("api-server-v2"),
	})
	if err == nil || !strings.Contains(err.Error(), "ci config name already exists") {
		t.Fatalf("expected duplicate name error, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestDeleteCIConfigRejectsReferencedDeployPlans(t *testing.T) {
	db, mock, err := testutil.NewMockDB()
	if err != nil {
		t.Fatalf("create mock db: %v", err)
	}
	dao.SetDefault(db)

	currentRows := sqlmock.NewRows([]string{
		"id", "created_at", "updated_at", "name", "business_unit_id", "image_registry", "image_repo", "image_tag_rule", "build_spec",
	}).AddRow(
		12,
		time.Date(2026, time.March, 18, 10, 0, 0, 0, time.UTC),
		time.Date(2026, time.March, 18, 12, 0, 0, 0, time.UTC),
		"api-server",
		34,
		"harbor.example.com/project-a",
		"api-server",
		jsonValue(t, `{"type":"branch"}`),
		jsonValue(t, `{}`),
	)

	mock.ExpectQuery("SELECT \\* FROM `ci_configs` WHERE .*`id` = \\?.* ORDER BY .*`id` LIMIT \\?").
		WithArgs(12, 1).
		WillReturnRows(currentRows)
	mock.ExpectQuery("SELECT count\\(\\*\\) FROM `deploy_plans` WHERE .*`ci_config_id` = \\?").
		WithArgs(12).
		WillReturnRows(sqlmock.NewRows([]string{"count(*)"}).AddRow(2))

	err = App.DeleteCIConfig(context.Background(), 12)
	if err == nil || !strings.Contains(err.Error(), "禁止删除") {
		t.Fatalf("expected delete guard error, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestDeleteCIConfigSucceedsWhenUnreferenced(t *testing.T) {
	db, mock, err := testutil.NewMockDB()
	if err != nil {
		t.Fatalf("create mock db: %v", err)
	}
	dao.SetDefault(db)

	currentRows := sqlmock.NewRows([]string{
		"id", "created_at", "updated_at", "name", "business_unit_id", "image_registry", "image_repo", "image_tag_rule", "build_spec",
	}).AddRow(
		12,
		time.Date(2026, time.March, 18, 10, 0, 0, 0, time.UTC),
		time.Date(2026, time.March, 18, 12, 0, 0, 0, time.UTC),
		"api-server",
		34,
		"harbor.example.com/project-a",
		"api-server",
		jsonValue(t, `{"type":"branch"}`),
		jsonValue(t, `{}`),
	)

	mock.ExpectQuery("SELECT \\* FROM `ci_configs` WHERE .*`id` = \\?.* ORDER BY .*`id` LIMIT \\?").
		WithArgs(12, 1).
		WillReturnRows(currentRows)
	mock.ExpectQuery("SELECT count\\(\\*\\) FROM `deploy_plans` WHERE .*`ci_config_id` = \\?").
		WithArgs(12).
		WillReturnRows(sqlmock.NewRows([]string{"count(*)"}).AddRow(0))
	mock.ExpectExec("UPDATE `ci_configs` SET `deleted_at`=\\? WHERE `ci_configs`.`id` = \\? AND `ci_configs`.`deleted_at` IS NULL").
		WithArgs(sqlmock.AnyArg(), 12).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := App.DeleteCIConfig(context.Background(), 12); err != nil {
		t.Fatalf("delete ci config: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestGetCIConfigAppliesBuildSpecDefaults(t *testing.T) {
	db, mock, err := testutil.NewMockDB()
	if err != nil {
		t.Fatalf("create mock db: %v", err)
	}
	dao.SetDefault(db)

	rows := sqlmock.NewRows([]string{
		"id", "created_at", "updated_at", "name", "business_unit_id", "image_registry", "image_repo", "image_tag_rule", "build_spec",
	}).AddRow(
		12,
		time.Date(2026, time.March, 18, 10, 0, 0, 0, time.UTC),
		time.Date(2026, time.March, 18, 12, 0, 0, 0, time.UTC),
		"api-server",
		34,
		"harbor.example.com/project-a",
		"api-server",
		jsonValue(t, `{"type":"branch"}`),
		jsonValue(t, `{}`),
	)

	mock.ExpectQuery("SELECT \\* FROM `ci_configs` WHERE .*`id` = \\?.* ORDER BY .*`id` LIMIT \\?").
		WithArgs(12, 1).
		WillReturnRows(rows)

	item, err := App.GetCIConfig(context.Background(), 12)
	if err != nil {
		t.Fatalf("get ci config: %v", err)
	}
	if item.BuildSpec.MakefilePath != "./Makefile" || item.BuildSpec.DockerfilePath != "./Dockerfile" {
		t.Fatalf("expected defaulted build paths, got %+v", item.BuildSpec)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func jsonValue(t *testing.T, raw string) driver.Value {
	t.Helper()
	return []byte(raw)
}

func stringPtr(value string) *string {
	return &value
}
