package metadata

import (
	"reflect"
	"testing"

	"github.com/richer421/q-metahub/app/metadata/vo"
	"github.com/richer421/q-metahub/infra/mysql/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCDConfigModelIncludesReleaseRegionField(t *testing.T) {
	t.Parallel()

	typ := reflect.TypeOf(model.CDConfig{})
	_, ok := typ.FieldByName("ReleaseRegion")

	assert.True(t, ok, "CDConfig should include ReleaseRegion field")
}

func TestCDConfigModelIncludesReleaseEnvField(t *testing.T) {
	t.Parallel()

	typ := reflect.TypeOf(model.CDConfig{})
	_, ok := typ.FieldByName("ReleaseEnv")

	assert.True(t, ok, "CDConfig should include ReleaseEnv field")
}

func TestBuildCDConfigModelRollingUsesDefaults(t *testing.T) {
	t.Parallel()

	entity, err := buildCDConfigModel(7, vo.UpsertCDConfigReq{
		Name:           "api-server-dev",
		ReleaseRegion:  "华东",
		ReleaseEnv:     "开发",
		DeploymentMode: "滚动发布",
	}, nil)
	require.NoError(t, err)

	assert.Equal(t, int64(7), entity.BusinessUnitID)
	assert.Equal(t, "api-server-dev", entity.Name)
	assert.Equal(t, "cn-east", entity.ReleaseRegion)
	assert.Equal(t, "dev", entity.ReleaseEnv)
	assert.Equal(t, "helm", entity.RenderEngine)
	assert.Equal(t, "", entity.ValuesYAML)
	assert.Equal(t, model.DeploymentModeRolling, entity.ReleaseStrategy.DeploymentMode)
	assert.Equal(t, 1, entity.ReleaseStrategy.BatchRule.BatchCount)
	assert.Equal(t, []float64{1}, entity.ReleaseStrategy.BatchRule.BatchRatio)
	if assert.NotNil(t, entity.GitOps) {
		assert.True(t, entity.GitOps.Enabled)
		assert.Equal(t, "main", entity.GitOps.Branch)
		assert.Equal(t, "apps", entity.GitOps.AppRoot)
		assert.Equal(t, "manifests", entity.GitOps.ManifestRoot)
	}
}

func TestBuildCDConfigModelCanaryMapsAdvancedFields(t *testing.T) {
	t.Parallel()

	manualAdjust := true
	timeout := 600
	trafficBatchCount := 3

	entity, err := buildCDConfigModel(8, vo.UpsertCDConfigReq{
		Name:                 "api-server-gray",
		ReleaseRegion:        "新加坡",
		ReleaseEnv:           "灰度",
		DeploymentMode:       "金丝雀发布",
		TrafficBatchCount:    &trafficBatchCount,
		TrafficRatioList:     []float64{10, 30, 60},
		ManualAdjust:         &manualAdjust,
		AdjustTimeoutSeconds: &timeout,
	}, nil)
	require.NoError(t, err)
	require.NotNil(t, entity.ReleaseStrategy.CanaryTrafficRule)

	assert.Equal(t, "ap-singapore", entity.ReleaseRegion)
	assert.Equal(t, "gray", entity.ReleaseEnv)
	assert.Equal(t, model.DeploymentModeCanary, entity.ReleaseStrategy.DeploymentMode)
	assert.Equal(t, 3, entity.ReleaseStrategy.CanaryTrafficRule.TrafficBatchCount)
	assert.Equal(t, []float64{0.1, 0.3, 0.6}, entity.ReleaseStrategy.CanaryTrafficRule.TrafficRatioList)
	assert.True(t, entity.ReleaseStrategy.CanaryTrafficRule.ManualAdjust)
	assert.Equal(t, 600, entity.ReleaseStrategy.CanaryTrafficRule.AdjustTimeout)
}

func TestToCDConfigFrontendVOMapsChineseLabelsAndSummary(t *testing.T) {
	t.Parallel()

	res := toCDConfigFrontendVO(&model.CDConfig{
		BaseModel: model.BaseModel{ID: 1},
		Name:           "api-server-prod",
		BusinessUnitID: 9,
		ReleaseRegion:  "cn-north",
		ReleaseEnv:     "prod",
		ReleaseStrategy: model.ReleaseStrategy{
			DeploymentMode: model.DeploymentModeCanary,
			CanaryTrafficRule: &model.CanaryTrafficRule{
				TrafficBatchCount: 3,
				TrafficRatioList:  []float64{0.1, 0.3, 0.6},
				ManualAdjust:      true,
				AdjustTimeout:     300,
			},
		},
	})

	assert.Equal(t, "华北", res.ReleaseRegion)
	assert.Equal(t, "生产", res.ReleaseEnv)
	assert.Equal(t, "金丝雀发布", res.DeploymentMode)
	assert.Equal(t, "3 批次 / 10%,30%,60%", res.StrategySummary)
	require.NotNil(t, res.TrafficBatchCount)
	assert.Equal(t, 3, *res.TrafficBatchCount)
	assert.Equal(t, []float64{10, 30, 60}, res.TrafficRatioList)
	require.NotNil(t, res.ManualAdjust)
	assert.True(t, *res.ManualAdjust)
	require.NotNil(t, res.AdjustTimeoutSeconds)
	assert.Equal(t, 300, *res.AdjustTimeoutSeconds)
}

func TestFilterCDConfigsByDeploymentModeKeepsRequestedStrategy(t *testing.T) {
	t.Parallel()

	rows := []*model.CDConfig{
		{
			Name: "api-server-dev",
			ReleaseStrategy: model.ReleaseStrategy{
				DeploymentMode: model.DeploymentModeRolling,
			},
		},
		{
			Name: "api-server-gray",
			ReleaseStrategy: model.ReleaseStrategy{
				DeploymentMode: model.DeploymentModeCanary,
			},
		},
	}

	filtered := filterCDConfigsByDeploymentMode(rows, model.DeploymentModeCanary)
	require.Len(t, filtered, 1)
	assert.Equal(t, "api-server-gray", filtered[0].Name)
}
