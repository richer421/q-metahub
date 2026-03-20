package metadata

import (
	"context"
	"fmt"
	"strings"

	"github.com/richer421/q-metahub/app/metadata/vo"
	"github.com/richer421/q-metahub/infra/mysql/dao"
	"github.com/richer421/q-metahub/infra/mysql/model"
)

func listBusinessUnits(ctx context.Context, page int, pageSize int, keyword string) (*vo.BusinessUnitPageDTO, error) {
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	query := dao.Q.WithContext(ctx).BusinessUnit
	if normalizedKeyword := strings.TrimSpace(keyword); normalizedKeyword != "" {
		like := "%" + normalizedKeyword + "%"
		query = query.Where(dao.BusinessUnit.Name.Like(like)).Or(dao.BusinessUnit.Description.Like(like))
	}

	offset := (page - 1) * pageSize
	rows, total, err := query.Order(dao.BusinessUnit.UpdatedAt.Desc()).FindByPage(offset, pageSize)
	if err != nil {
		return nil, err
	}

	items := make([]vo.BusinessUnitVO, 0, len(rows))
	for _, row := range rows {
		items = append(items, toBusinessUnitDTO(row))
	}

	return &vo.BusinessUnitPageDTO{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func createBusinessUnit(ctx context.Context, req vo.CreateBusinessUnitReq) (*vo.BusinessUnitVO, error) {
	entity := &model.BusinessUnit{
		Name:        strings.TrimSpace(req.Name),
		Description: strings.TrimSpace(req.Description),
		ProjectID:   req.ProjectID,
	}

	if err := dao.Q.WithContext(ctx).BusinessUnit.Create(entity); err != nil {
		return nil, err
	}

	dto := toBusinessUnitDTO(entity)
	return &dto, nil
}

func updateBusinessUnit(ctx context.Context, businessUnitID int64, req vo.UpdateBusinessUnitReq) (*vo.BusinessUnitVO, error) {
	entity, err := dao.Q.WithContext(ctx).BusinessUnit.Where(dao.BusinessUnit.ID.Eq(businessUnitID)).First()
	if err != nil {
		return nil, err
	}

	entity.Name = strings.TrimSpace(req.Name)
	entity.Description = strings.TrimSpace(req.Description)

	if err := dao.Q.WithContext(ctx).BusinessUnit.Save(entity); err != nil {
		return nil, err
	}

	dto := toBusinessUnitDTO(entity)
	return &dto, nil
}

func deleteBusinessUnit(ctx context.Context, businessUnitID int64) error {
	entity, err := dao.Q.WithContext(ctx).BusinessUnit.Where(dao.BusinessUnit.ID.Eq(businessUnitID)).First()
	if err != nil {
		return err
	}

	blocked, err := businessUnitDeleteBlocked(ctx, businessUnitID)
	if err != nil {
		return err
	}
	if blocked {
		return fmt.Errorf("business unit has related metadata and cannot be deleted")
	}

	_, err = dao.Q.WithContext(ctx).BusinessUnit.Delete(entity)
	return err
}

func businessUnitDeleteBlocked(ctx context.Context, businessUnitID int64) (bool, error) {
	type dependencyCount struct {
		name  string
		count func(context.Context) (int64, error)
	}

	checks := []dependencyCount{
		{
			name: "ci config",
			count: func(ctx context.Context) (int64, error) {
				return dao.Q.WithContext(ctx).CIConfig.Where(dao.CIConfig.BusinessUnitID.Eq(businessUnitID)).Count()
			},
		},
		{
			name: "cd config",
			count: func(ctx context.Context) (int64, error) {
				return dao.Q.WithContext(ctx).CDConfig.Where(dao.CDConfig.BusinessUnitID.Eq(businessUnitID)).Count()
			},
		},
		{
			name: "instance oam",
			count: func(ctx context.Context) (int64, error) {
				return dao.Q.WithContext(ctx).InstanceOAM.Where(dao.InstanceOAM.BusinessUnitID.Eq(businessUnitID)).Count()
			},
		},
		{
			name: "deploy plan",
			count: func(ctx context.Context) (int64, error) {
				return dao.Q.WithContext(ctx).DeployPlan.Where(dao.DeployPlan.BusinessUnitID.Eq(businessUnitID)).Count()
			},
		},
	}

	for _, check := range checks {
		total, err := check.count(ctx)
		if err != nil {
			return false, fmt.Errorf("count %s: %w", check.name, err)
		}
		if total > 0 {
			return true, nil
		}
	}

	return false, nil
}

func toBusinessUnitDTO(entity *model.BusinessUnit) vo.BusinessUnitVO {
	return vo.BusinessUnitVO{
		ID:          entity.ID,
		CreatedAt:   entity.CreatedAt,
		UpdatedAt:   entity.UpdatedAt,
		Name:        entity.Name,
		Description: entity.Description,
		ProjectID:   entity.ProjectID,
	}
}
