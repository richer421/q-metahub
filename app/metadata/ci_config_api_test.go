package metadata

import (
	"testing"
	"time"

	"github.com/richer421/q-metahub/infra/mysql/model"
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
