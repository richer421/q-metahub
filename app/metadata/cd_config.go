package metadata

import (
	"context"
	"fmt"
	"math"
	"strings"

	"github.com/richer421/q-metahub/app/metadata/vo"
	"github.com/richer421/q-metahub/infra/mysql/dao"
	"github.com/richer421/q-metahub/infra/mysql/model"
)

const (
	defaultCDRenderEngine     = "helm"
	defaultGitOpsBranch       = "main"
	defaultGitOpsAppRoot      = "apps"
	defaultGitOpsManifestRoot = "manifests"
)

var supportedReleaseRegions = map[string]struct{}{
	"cn-east":      {},
	"cn-north":     {},
	"ap-singapore": {},
}

var supportedReleaseEnvs = map[string]struct{}{
	"dev":  {},
	"test": {},
	"gray": {},
	"prod": {},
}

var supportedDeploymentModes = map[model.DeploymentMode]struct{}{
	model.DeploymentModeRolling: {},
	model.DeploymentModeCanary:  {},
}

func listBusinessUnitCDConfigs(ctx context.Context, businessUnitID int64, req vo.CDConfigListReq) (*vo.CDConfigPageDTO, error) {
	page := req.Page
	if page < 1 {
		page = 1
	}

	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 10
	}

	query := dao.Q.WithContext(ctx).
		CDConfig.
		Where(dao.CDConfig.BusinessUnitID.Eq(businessUnitID))

	if keyword := strings.TrimSpace(req.Keyword); keyword != "" {
		query = query.Where(dao.CDConfig.Name.Like("%" + keyword + "%"))
	}

	if region, err := normalizeReleaseRegion(req.ReleaseRegion); err != nil {
		return nil, err
	} else if region != "" {
		query = query.Where(dao.CDConfig.ReleaseRegion.Eq(region))
	}

	if env, err := normalizeReleaseEnv(req.ReleaseEnv); err != nil {
		return nil, err
	} else if env != "" {
		query = query.Where(dao.CDConfig.ReleaseEnv.Eq(env))
	}

	var modeFilter model.DeploymentMode
	if mode, err := normalizeDeploymentMode(req.DeploymentMode); err != nil {
		return nil, err
	} else if mode != "" {
		modeFilter = mode
	}

	rows, err := query.Order(dao.CDConfig.UpdatedAt.Desc()).Find()
	if err != nil {
		return nil, fmt.Errorf("list cd configs by business_unit_id=%d: %w", businessUnitID, err)
	}

	rows = filterCDConfigsByDeploymentMode(rows, modeFilter)
	total := int64(len(rows))

	offset := (page - 1) * pageSize
	if offset > len(rows) {
		offset = len(rows)
	}
	end := offset + pageSize
	if end > len(rows) {
		end = len(rows)
	}
	rows = rows[offset:end]

	items := make([]vo.CDConfigFrontendVO, 0, len(rows))
	for _, row := range rows {
		items = append(items, toCDConfigFrontendVO(row))
	}

	return &vo.CDConfigPageDTO{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func getCDConfig(ctx context.Context, id int64) (*vo.CDConfigFrontendVO, error) {
	row, err := dao.Q.WithContext(ctx).CDConfig.Where(dao.CDConfig.ID.Eq(id)).First()
	if err != nil {
		return nil, fmt.Errorf("query cd_config id=%d: %w", id, err)
	}

	res := toCDConfigFrontendVO(row)
	return &res, nil
}

func createBusinessUnitCDConfig(ctx context.Context, businessUnitID int64, req vo.UpsertCDConfigReq) (*vo.CDConfigFrontendVO, error) {
	entity, err := buildCDConfigModel(businessUnitID, req, nil)
	if err != nil {
		return nil, err
	}

	if err := dao.Q.WithContext(ctx).CDConfig.Create(entity); err != nil {
		return nil, fmt.Errorf("create cd_config for business_unit_id=%d: %w", businessUnitID, err)
	}

	res := toCDConfigFrontendVO(entity)
	return &res, nil
}

func updateCDConfig(ctx context.Context, id int64, req vo.UpsertCDConfigReq) (*vo.CDConfigFrontendVO, error) {
	current, err := dao.Q.WithContext(ctx).CDConfig.Where(dao.CDConfig.ID.Eq(id)).First()
	if err != nil {
		return nil, fmt.Errorf("query cd_config id=%d: %w", id, err)
	}

	entity, err := buildCDConfigModel(current.BusinessUnitID, req, current)
	if err != nil {
		return nil, err
	}
	entity.ID = current.ID
	entity.CreatedAt = current.CreatedAt

	if err := dao.Q.WithContext(ctx).CDConfig.Save(entity); err != nil {
		return nil, fmt.Errorf("update cd_config id=%d: %w", id, err)
	}

	res := toCDConfigFrontendVO(entity)
	return &res, nil
}

func deleteCDConfig(ctx context.Context, id int64) error {
	count, err := dao.Q.WithContext(ctx).DeployPlan.Where(dao.DeployPlan.CDConfigID.Eq(id)).Count()
	if err != nil {
		return fmt.Errorf("count deploy plans by cd_config_id=%d: %w", id, err)
	}
	if count > 0 {
		return fmt.Errorf("该 CD 配置已被部署计划引用，禁止删除")
	}

	entity, err := dao.Q.WithContext(ctx).CDConfig.Where(dao.CDConfig.ID.Eq(id)).First()
	if err != nil {
		return fmt.Errorf("query cd_config id=%d: %w", id, err)
	}

	if _, err := dao.Q.WithContext(ctx).CDConfig.Delete(entity); err != nil {
		return fmt.Errorf("delete cd_config id=%d: %w", id, err)
	}

	return nil
}

func filterCDConfigsByDeploymentMode(rows []*model.CDConfig, mode model.DeploymentMode) []*model.CDConfig {
	if mode == "" {
		return rows
	}

	filtered := make([]*model.CDConfig, 0, len(rows))
	for _, row := range rows {
		if row == nil {
			continue
		}
		if row.ReleaseStrategy.DeploymentMode == mode {
			filtered = append(filtered, row)
		}
	}
	return filtered
}

func buildCDConfigModel(businessUnitID int64, req vo.UpsertCDConfigReq, current *model.CDConfig) (*model.CDConfig, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}

	releaseRegion, err := normalizeReleaseRegion(req.ReleaseRegion)
	if err != nil {
		return nil, err
	}
	if releaseRegion == "" {
		return nil, fmt.Errorf("release_region is required")
	}

	releaseEnv, err := normalizeReleaseEnv(req.ReleaseEnv)
	if err != nil {
		return nil, err
	}
	if releaseEnv == "" {
		return nil, fmt.Errorf("release_env is required")
	}

	deploymentMode, err := normalizeDeploymentMode(req.DeploymentMode)
	if err != nil {
		return nil, err
	}
	if deploymentMode == "" {
		return nil, fmt.Errorf("deployment_mode is required")
	}

	strategy, err := buildReleaseStrategy(deploymentMode, req)
	if err != nil {
		return nil, err
	}

	entity := &model.CDConfig{
		Name:            name,
		BusinessUnitID:  businessUnitID,
		ReleaseRegion:   releaseRegion,
		ReleaseEnv:      releaseEnv,
		RenderEngine:    defaultCDRenderEngine,
		ValuesYAML:      "",
		ReleaseStrategy: strategy,
		GitOps: &model.GitOpsConfig{
			Enabled:      true,
			Branch:       defaultGitOpsBranch,
			AppRoot:      defaultGitOpsAppRoot,
			ManifestRoot: defaultGitOpsManifestRoot,
		},
	}

	if current != nil {
		entity.ID = current.ID
		entity.CreatedAt = current.CreatedAt
		if strings.TrimSpace(current.RenderEngine) != "" {
			entity.RenderEngine = current.RenderEngine
		}
		entity.ValuesYAML = current.ValuesYAML
		if current.GitOps != nil {
			gitOpsCopy := *current.GitOps
			entity.GitOps = &gitOpsCopy
		}
	}

	return entity, nil
}

func buildReleaseStrategy(mode model.DeploymentMode, req vo.UpsertCDConfigReq) (model.ReleaseStrategy, error) {
	switch mode {
	case model.DeploymentModeRolling:
		return model.ReleaseStrategy{
			DeploymentMode: model.DeploymentModeRolling,
			BatchRule: model.BatchRule{
				BatchCount:  1,
				BatchRatio:  []float64{1},
				TriggerType: model.TriggerTypeAuto,
				Interval:    0,
			},
		}, nil
	case model.DeploymentModeCanary:
		if req.TrafficBatchCount == nil || *req.TrafficBatchCount <= 0 {
			return model.ReleaseStrategy{}, fmt.Errorf("traffic_batch_count is required for canary deployment")
		}
		if len(req.TrafficRatioList) == 0 {
			return model.ReleaseStrategy{}, fmt.Errorf("traffic_ratio_list is required for canary deployment")
		}

		trafficRatios := make([]float64, 0, len(req.TrafficRatioList))
		for _, ratio := range req.TrafficRatioList {
			if ratio <= 0 {
				return model.ReleaseStrategy{}, fmt.Errorf("traffic_ratio_list must contain positive values")
			}
			trafficRatios = append(trafficRatios, ratio/100)
		}

		var manualAdjust bool
		if req.ManualAdjust != nil {
			manualAdjust = *req.ManualAdjust
		}

		adjustTimeout := 0
		if req.AdjustTimeoutSeconds != nil {
			adjustTimeout = *req.AdjustTimeoutSeconds
		}
		if manualAdjust && adjustTimeout <= 0 {
			return model.ReleaseStrategy{}, fmt.Errorf("adjust_timeout_seconds is required when manual_adjust is enabled")
		}

		return model.ReleaseStrategy{
			DeploymentMode: model.DeploymentModeCanary,
			BatchRule: model.BatchRule{
				BatchCount:  1,
				BatchRatio:  []float64{1},
				TriggerType: model.TriggerTypeAuto,
				Interval:    0,
			},
			CanaryTrafficRule: &model.CanaryTrafficRule{
				TrafficBatchCount: *req.TrafficBatchCount,
				TrafficRatioList:  trafficRatios,
				ManualAdjust:      manualAdjust,
				AdjustTimeout:     adjustTimeout,
			},
		}, nil
	default:
		return model.ReleaseStrategy{}, fmt.Errorf("invalid deployment_mode")
	}
}

func toCDConfigFrontendVO(row *model.CDConfig) vo.CDConfigFrontendVO {
	res := vo.CDConfigFrontendVO{
		ID:              row.ID,
		CreatedAt:       row.CreatedAt,
		UpdatedAt:       row.UpdatedAt,
		Name:            row.Name,
		BusinessUnitID:  row.BusinessUnitID,
		ReleaseRegion:   row.ReleaseRegion,
		ReleaseEnv:      row.ReleaseEnv,
		DeploymentMode:  string(row.ReleaseStrategy.DeploymentMode),
		StrategySummary: buildStrategySummary(row.ReleaseStrategy),
	}

	if row.ReleaseStrategy.CanaryTrafficRule != nil {
		rule := row.ReleaseStrategy.CanaryTrafficRule
		trafficBatchCount := rule.TrafficBatchCount
		manualAdjust := rule.ManualAdjust
		adjustTimeout := rule.AdjustTimeout
		res.TrafficBatchCount = &trafficBatchCount
		res.ManualAdjust = &manualAdjust
		res.AdjustTimeoutSeconds = &adjustTimeout
		res.TrafficRatioList = denormalizeTrafficRatios(rule.TrafficRatioList)
	}

	return res
}

func normalizeReleaseRegion(value string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", nil
	}
	if _, ok := supportedReleaseRegions[trimmed]; ok {
		return trimmed, nil
	}
	return "", fmt.Errorf("invalid release_region")
}

func normalizeReleaseEnv(value string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", nil
	}
	if _, ok := supportedReleaseEnvs[trimmed]; ok {
		return trimmed, nil
	}
	return "", fmt.Errorf("invalid release_env")
}

func normalizeDeploymentMode(value string) (model.DeploymentMode, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", nil
	}
	mode := model.DeploymentMode(trimmed)
	if _, ok := supportedDeploymentModes[mode]; ok {
		return mode, nil
	}
	return "", fmt.Errorf("invalid deployment_mode")
}

func denormalizeTrafficRatios(values []float64) []float64 {
	res := make([]float64, 0, len(values))
	for _, value := range values {
		res = append(res, math.Round(value*10000)/100)
	}
	return res
}

func buildStrategySummary(strategy model.ReleaseStrategy) string {
	if strategy.DeploymentMode == model.DeploymentModeCanary && strategy.CanaryTrafficRule != nil {
		parts := make([]string, 0, len(strategy.CanaryTrafficRule.TrafficRatioList))
		for _, ratio := range strategy.CanaryTrafficRule.TrafficRatioList {
			parts = append(parts, fmt.Sprintf("%g%%", math.Round(ratio*10000)/100))
		}
		return fmt.Sprintf("%d batches / %s", strategy.CanaryTrafficRule.TrafficBatchCount, strings.Join(parts, ","))
	}

	return "rolling (default strategy)"
}
