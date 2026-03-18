package metadata

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/richer421/q-metahub/app/metadata/vo"
	"github.com/richer421/q-metahub/infra/mysql/dao"
	"github.com/richer421/q-metahub/infra/mysql/model"
	"gorm.io/gorm"
)

var (
	validImageTagRuleTypes = map[string]struct{}{
		"branch":    {},
		"tag":       {},
		"commit":    {},
		"timestamp": {},
		"custom":    {},
	}
	validCustomTemplatePattern = regexp.MustCompile(`^([A-Za-z0-9]+|[-_.]|\$\{(branch|tag|commit|timestamp)\})+$`)
)

func (s *app) ListBusinessUnitCIConfigs(ctx context.Context, businessUnitID int64, page int, pageSize int, keyword string) (*vo.CIConfigPageVO, error) {
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	query := dao.Q.WithContext(ctx).
		CIConfig.
		Where(dao.CIConfig.BusinessUnitID.Eq(businessUnitID))

	if normalizedKeyword := strings.TrimSpace(keyword); normalizedKeyword != "" {
		query = query.Where(dao.CIConfig.Name.Like("%" + normalizedKeyword + "%"))
	}

	offset := (page - 1) * pageSize
	rows, total, err := query.Order(dao.CIConfig.UpdatedAt.Desc()).FindByPage(offset, pageSize)
	if err != nil {
		return nil, err
	}

	items := make([]vo.CIConfigVO, 0, len(rows))
	for _, row := range rows {
		items = append(items, toCIConfigVO(row))
	}

	return &vo.CIConfigPageVO{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (s *app) GetCIConfig(ctx context.Context, ciConfigID int64) (*vo.CIConfigVO, error) {
	row, err := dao.Q.WithContext(ctx).CIConfig.Where(dao.CIConfig.ID.Eq(ciConfigID)).First()
	if err != nil {
		return nil, err
	}

	item := toCIConfigVO(row)
	return &item, nil
}

func (s *app) CreateBusinessUnitCIConfig(ctx context.Context, businessUnitID int64, req vo.CreateCIConfigReq) (*vo.CIConfigVO, error) {
	businessUnit, err := dao.Q.WithContext(ctx).BusinessUnit.Where(dao.BusinessUnit.ID.Eq(businessUnitID)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("business unit not found")
		}
		return nil, err
	}

	name := strings.TrimSpace(req.Name)
	if _, err := dao.Q.WithContext(ctx).CIConfig.
		Where(dao.CIConfig.BusinessUnitID.Eq(businessUnitID)).
		Where(dao.CIConfig.Name.Eq(name)).
		First(); err == nil {
		return nil, fmt.Errorf("ci config name already exists in current business unit")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	entity, err := normalizeCreateCIConfig(businessUnitID, req, businessUnit.Name)
	if err != nil {
		return nil, err
	}
	if err := dao.Q.WithContext(ctx).CIConfig.Create(entity); err != nil {
		return nil, err
	}

	item := toCIConfigVO(entity)
	return &item, nil
}

func (s *app) UpdateCIConfig(ctx context.Context, ciConfigID int64, req vo.UpdateCIConfigReq) (*vo.CIConfigVO, error) {
	current, err := dao.Q.WithContext(ctx).CIConfig.Where(dao.CIConfig.ID.Eq(ciConfigID)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("ci config not found")
		}
		return nil, err
	}

	next, err := mergeCIConfigUpdate(current, req)
	if err != nil {
		return nil, err
	}
	if next.Name != current.Name {
		if _, err := dao.Q.WithContext(ctx).CIConfig.
			Where(dao.CIConfig.BusinessUnitID.Eq(current.BusinessUnitID)).
			Where(dao.CIConfig.Name.Eq(next.Name)).
			Not(dao.CIConfig.ID.Eq(ciConfigID)).
			First(); err == nil {
			return nil, fmt.Errorf("ci config name already exists in current business unit")
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
	}
	if err := dao.Q.WithContext(ctx).CIConfig.Save(next); err != nil {
		return nil, err
	}

	item := toCIConfigVO(next)
	return &item, nil
}

func (s *app) DeleteCIConfig(ctx context.Context, ciConfigID int64) error {
	current, err := dao.Q.WithContext(ctx).CIConfig.Where(dao.CIConfig.ID.Eq(ciConfigID)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("ci config not found")
		}
		return err
	}

	refCount, err := dao.Q.WithContext(ctx).DeployPlan.Where(dao.DeployPlan.CIConfigID.Eq(ciConfigID)).Count()
	if err != nil {
		return err
	}
	if refCount > 0 {
		return fmt.Errorf("ci config is referenced by %d deploy plans and cannot be deleted", refCount)
	}

	_, err = dao.Q.WithContext(ctx).CIConfig.Delete(current)
	return err
}

func normalizeCreateCIConfig(businessUnitID int64, req vo.CreateCIConfigReq, businessUnitName string) (*model.CIConfig, error) {
	name := strings.TrimSpace(req.Name)
	imageRegistry := normalizeImageRegistry(req.ImageRegistry)
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if imageRegistry == "" {
		return nil, fmt.Errorf("image_registry is required")
	}
	imageTagRule, err := normalizeImageTagRule(req.ImageTagRule)
	if err != nil {
		return nil, err
	}
	buildSpec, err := normalizeBuildSpec(req.BuildSpec, model.BuildSpec{})
	if err != nil {
		return nil, err
	}

	return &model.CIConfig{
		Name:           name,
		BusinessUnitID: businessUnitID,
		ImageRegistry:  imageRegistry,
		ImageRepo:      normalizeImageRepo(businessUnitName),
		ImageTagRule:   imageTagRule,
		BuildSpec:      buildSpec,
	}, nil
}

func mergeCIConfigUpdate(current *model.CIConfig, req vo.UpdateCIConfigReq) (*model.CIConfig, error) {
	next := *current

	if req.Name != nil {
		next.Name = strings.TrimSpace(*req.Name)
		if next.Name == "" {
			return nil, fmt.Errorf("name is required")
		}
	}
	if req.ImageRegistry != nil {
		next.ImageRegistry = normalizeImageRegistry(*req.ImageRegistry)
		if next.ImageRegistry == "" {
			return nil, fmt.Errorf("image_registry is required")
		}
	}
	if req.ImageTagRule != nil {
		imageTagRule, err := normalizeImageTagRule(*req.ImageTagRule)
		if err != nil {
			return nil, err
		}
		next.ImageTagRule = imageTagRule
	}
	if req.BuildSpec != nil {
		buildSpec, err := normalizeBuildSpec(*req.BuildSpec, current.BuildSpec)
		if err != nil {
			return nil, err
		}
		next.BuildSpec = buildSpec
	}

	return &next, nil
}

func normalizeImageRegistry(value string) string {
	return strings.TrimRight(strings.TrimSpace(value), "/")
}

func normalizeImageRepo(name string) string {
	normalized := strings.ToLower(strings.TrimSpace(name))
	var builder strings.Builder
	lastHyphen := false

	for _, r := range normalized {
		switch {
		case r >= 'a' && r <= 'z':
			builder.WriteRune(r)
			lastHyphen = false
		case r >= '0' && r <= '9':
			builder.WriteRune(r)
			lastHyphen = false
		case r == ' ' || r == '-' || r == '_' || r == '.':
			if builder.Len() > 0 && !lastHyphen {
				builder.WriteByte('-')
				lastHyphen = true
			}
		}
	}

	return strings.Trim(builder.String(), "-")
}

func normalizeImageTagRule(in vo.CIConfigImageTagRuleVO) (model.ImageTagRule, error) {
	if _, ok := validImageTagRuleTypes[in.Type]; !ok {
		return model.ImageTagRule{}, fmt.Errorf("invalid image tag rule type")
	}
	if in.Type == "custom" {
		if !validCustomTemplatePattern.MatchString(in.Template) {
			return model.ImageTagRule{}, fmt.Errorf("invalid image tag rule template")
		}
		return model.ImageTagRule{
			Type:     in.Type,
			Template: in.Template,
		}, nil
	}

	return model.ImageTagRule{
		Type:          in.Type,
		WithTimestamp: in.Type == "branch" && in.WithTimestamp,
		WithCommit:    in.Type == "branch" && in.WithCommit,
	}, nil
}

func normalizeBuildSpec(in vo.CIConfigBuildSpecVO, current model.BuildSpec) (model.BuildSpec, error) {
	buildSpec := current

	makefileFallback := current.MakefilePath
	if makefileFallback == "" {
		makefileFallback = "./Makefile"
	}
	makefilePath, err := normalizeRelativePath(in.MakefilePath, makefileFallback)
	if err != nil {
		return model.BuildSpec{}, fmt.Errorf("invalid makefile path")
	}
	dockerfileFallback := current.DockerfilePath
	if dockerfileFallback == "" {
		dockerfileFallback = "./Dockerfile"
	}
	dockerfilePath, err := normalizeRelativePath(in.DockerfilePath, dockerfileFallback)
	if err != nil {
		return model.BuildSpec{}, fmt.Errorf("invalid dockerfile path")
	}

	buildSpec.MakefilePath = makefilePath
	if buildSpec.MakeCommand == "" {
		buildSpec.MakeCommand = "build"
	}
	buildSpec.DockerfilePath = dockerfilePath
	if buildSpec.DockerContext == "" {
		buildSpec.DockerContext = "."
	}
	if buildSpec.BuildArgs == nil {
		buildSpec.BuildArgs = map[string]string{}
	}

	return buildSpec, nil
}

func normalizeRelativePath(value string, fallback string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return fallback, nil
	}
	trimmed = strings.ReplaceAll(trimmed, "\\", "/")
	if strings.HasPrefix(trimmed, "/") || strings.Contains(trimmed, "..") || strings.Contains(trimmed, "//") {
		return "", fmt.Errorf("invalid path")
	}

	parts := strings.Split(strings.TrimPrefix(trimmed, "./"), "/")
	for _, part := range parts {
		if part == "" {
			return "", fmt.Errorf("invalid path")
		}
		for _, r := range part {
			if (r < 'a' || r > 'z') && (r < 'A' || r > 'Z') && (r < '0' || r > '9') && r != '.' && r != '_' && r != '-' {
				return "", fmt.Errorf("invalid path")
			}
		}
	}

	return "./" + strings.TrimPrefix(trimmed, "./"), nil
}
