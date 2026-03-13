package api

import (
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"

	appmetadata "github.com/richer421/q-metahub/app/metadata"
	"github.com/richer421/q-metahub/app/metadata/vo"
	"github.com/richer421/q-metahub/http/common"
	"github.com/richer421/q-metahub/infra/mysql/dao"
)

type MetadataAPI struct {
	svc *appmetadata.Service
}

func NewMetadataAPI() *MetadataAPI {
	return &MetadataAPI{svc: appmetadata.NewService()}
}

func RegisterMetadataRoutes(v1 *gin.RouterGroup) {
	api := NewMetadataAPI()

	v1.POST("/projects", api.CreateProject)
	v1.POST("/business-units", api.CreateBusinessUnit)
	v1.POST("/ci-configs", api.CreateCIConfig)
	v1.POST("/cd-configs", api.CreateCDConfig)
	v1.POST("/instance-configs", api.CreateInstanceConfig)
	v1.POST("/deploy-plans", api.CreateDeployPlanAggregate)
	v1.GET("/deploy-plans/:id", api.GetDeployPlan)
	v1.GET("/deploy-plans/:id/full-spec", api.GetDeployPlanFullSpec)
	v1.GET("/business-units/:id/full-spec", api.GetBusinessUnitFullSpec)
	v1.POST("/demo/seed", api.SeedDemoSetup)
}

func (a *MetadataAPI) CreateProject(c *gin.Context) {
	var req vo.CreateProjectReq
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, err)
		return
	}

	project := &dao.Project
	entity := &struct {
		GitID   int64  `json:"git_id"`
		Name    string `json:"name"`
		RepoURL string `json:"repo_url"`
	}{GitID: req.GitID, Name: req.Name, RepoURL: req.RepoURL}

	common.OK(c, entity)
	_ = project
}

func (a *MetadataAPI) CreateBusinessUnit(c *gin.Context) {
	var req vo.CreateBusinessUnitReq
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, req)
}

func (a *MetadataAPI) CreateCIConfig(c *gin.Context) {
	var req vo.CreateCIConfigReq
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, req)
}

func (a *MetadataAPI) CreateCDConfig(c *gin.Context) {
	var req vo.CreateCDConfigReq
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, req)
}

func (a *MetadataAPI) CreateInstanceConfig(c *gin.Context) {
	var req vo.CreateInstanceConfigReq
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, req)
}

func (a *MetadataAPI) CreateDeployPlanAggregate(c *gin.Context) {
	var req vo.CreateDeployPlanAggregateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, err)
		return
	}

	res, err := a.svc.CreateDeployPlanAggregate(c.Request.Context(), &req)
	if err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, res)
}

func (a *MetadataAPI) GetDeployPlan(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		common.Fail(c, fmt.Errorf("invalid deploy plan id"))
		return
	}

	entity, err := dao.Q.WithContext(c.Request.Context()).DeployPlan.Where(dao.DeployPlan.ID.Eq(id)).First()
	if err != nil {
		common.Fail(c, err)
		return
	}

	common.OK(c, gin.H{
		"id":                 entity.ID,
		"name":               entity.Name,
		"description":        entity.Description,
		"business_unit_id":   entity.BusinessUnitID,
		"ci_config_id":       entity.CIConfigID,
		"cd_config_id":       entity.CDConfigID,
		"instance_config_id": entity.InstanceConfigID,
	})
}

func (a *MetadataAPI) GetDeployPlanFullSpec(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		common.Fail(c, fmt.Errorf("invalid deploy plan id"))
		return
	}

	res, err := a.svc.GetDeployPlanFullSpec(c.Request.Context(), id)
	if err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, res)
}

func (a *MetadataAPI) GetBusinessUnitFullSpec(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		common.Fail(c, fmt.Errorf("invalid business unit id"))
		return
	}

	res, err := a.svc.GetBusinessUnitFullSpec(c.Request.Context(), id)
	if err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, res)
}

func (a *MetadataAPI) SeedDemoSetup(c *gin.Context) {
	res, err := a.svc.SeedDemoSetup(c.Request.Context())
	if err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, res)
}
