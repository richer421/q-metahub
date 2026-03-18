package metadata

import (
	"context"
	"database/sql/driver"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/richer421/q-metahub/infra/mysql/dao"
	"github.com/richer421/q-metahub/infra/mysql/model"
	"github.com/richer421/q-metahub/pkg/testutil"
)

func TestToCIConfigVOAppliesDerivedFields(t *testing.T) {
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
	if item.FullImageRepo != "harbor.example.com/project-a/api-server" {
		t.Fatalf("expected full image repo to be derived, got %q", item.FullImageRepo)
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

	mock.ExpectQuery("SELECT \\* FROM `ci_configs` WHERE .*`business_unit_id` = \\? AND .*`name` LIKE \\? ORDER BY .*`updated_at` DESC LIMIT \\?").
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

func TestGetCIConfigReturnsDerivedFields(t *testing.T) {
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

	mock.ExpectQuery("SELECT \\* FROM `ci_configs` WHERE .*`id` = \\? ORDER BY .*`id` LIMIT \\?").
		WithArgs(12, 1).
		WillReturnRows(rows)

	item, err := App.GetCIConfig(context.Background(), 12)
	if err != nil {
		t.Fatalf("get ci config: %v", err)
	}
	if item.FullImageRepo != "harbor.example.com/project-a/api-server" {
		t.Fatalf("expected derived full image repo, got %q", item.FullImageRepo)
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
