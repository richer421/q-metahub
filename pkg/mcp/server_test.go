package mcp

import (
	"encoding/json"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	metadatavo "github.com/richer421/q-metahub/app/metadata/vo"
)

func TestServerRegistersMetadataTools(t *testing.T) {
	server := NewServer()

	tools := server.metadataToolSpecs()
	names := make([]string, 0, len(tools))
	for _, tool := range tools {
		names = append(names, tool.Name)
	}

	assert.Contains(t, names, "seed_demo_setup")
	assert.Contains(t, names, "get_deploy_plan")
	assert.Contains(t, names, "get_business_unit_full_spec")
}

func TestGetDeployPlanToolUsesDeployPlanIDInput(t *testing.T) {
	server := NewServer()

	tools := server.metadataToolSpecs()
	for _, tool := range tools {
		if tool.Name == "get_deploy_plan" {
			assert.Equal(t, "Get a deploy plan by ID", tool.Description)
			return
		}
	}

	t.Fatal("get_deploy_plan tool not registered")
}

func TestSeedDemoSetupToolReturnsStructuredResult(t *testing.T) {
	server := NewServer()
	payload := &metadatavo.DeployPlanAggregateDTO{
		Project:      metadatavo.ProjectDTO{ID: 1, Name: "q-demo-project"},
		BusinessUnit: metadatavo.BusinessUnitDTO{ID: 2, Name: "q-demo"},
		DeployPlan:   metadatavo.DeployPlanDTO{ID: 6, Name: "q-demo-dev-plan"},
	}

	res, err := server.jsonResult(payload)
	require.NoError(t, err)
	require.Len(t, res.Content, 1)

	text, ok := res.Content[0].(*mcp.TextContent)
	require.True(t, ok)

	var decoded map[string]any
	require.NoError(t, json.Unmarshal([]byte(text.Text), &decoded))
	assert.Equal(t, float64(1), decoded["project"].(map[string]any)["id"])
	assert.Equal(t, float64(6), decoded["deploy_plan"].(map[string]any)["id"])
}
