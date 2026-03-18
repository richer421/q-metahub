package metadata

import (
	"context"
	"strings"

	"github.com/richer421/q-metahub/app/metadata/vo"
	"github.com/richer421/q-metahub/infra/mysql/dao"
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
