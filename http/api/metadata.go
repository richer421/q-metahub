package api

import (
	"fmt"
	"strconv"
	"strings"

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
	v1.POST("/instance-oams", api.CreateInstanceOAM)
	v1.GET("/business-units/:id/instance-oams", api.ListBusinessUnitInstanceOAMs)
	v1.POST("/business-units/:id/instance-oams", api.CreateBusinessUnitInstanceOAM)
	v1.PUT("/instance-oams/:id", api.UpdateInstanceOAM)
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

func (a *MetadataAPI) CreateInstanceOAM(c *gin.Context) {
	var req vo.CreateInstanceOAMReq
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, req)
}

func (a *MetadataAPI) ListBusinessUnitInstanceOAMs(c *gin.Context) {
	businessUnitID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		common.Fail(c, fmt.Errorf("invalid business unit id"))
		return
	}

	env := strings.TrimSpace(c.Query("env"))
	keyword := strings.TrimSpace(c.Query("keyword"))

	res, err := a.svc.ListInstanceOAMs(c.Request.Context(), businessUnitID, env, keyword)
	if err != nil {
		common.Fail(c, err)
		return
	}

	common.OK(c, res)
}

func (a *MetadataAPI) CreateBusinessUnitInstanceOAM(c *gin.Context) {
	businessUnitID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		common.Fail(c, fmt.Errorf("invalid business unit id"))
		return
	}

	var req vo.CreateInstanceOAMReq
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, err)
		return
	}

	if strings.TrimSpace(req.Name) == "" || strings.TrimSpace(req.Env) == "" {
		common.Fail(c, fmt.Errorf("name and env are required"))
		return
	}

	res, err := a.svc.CreateInstanceOAM(c.Request.Context(), businessUnitID, &req)
	if err != nil {
		common.Fail(c, err)
		return
	}

	common.OK(c, res)
}

func (a *MetadataAPI) UpdateInstanceOAM(c *gin.Context) {
	instanceOAMID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		common.Fail(c, fmt.Errorf("invalid instance oam id"))
		return
	}

	var req vo.UpdateInstanceOAMReq
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, err)
		return
	}

	if strings.TrimSpace(req.Name) == "" || strings.TrimSpace(req.Env) == "" {
		common.Fail(c, fmt.Errorf("name and env are required"))
		return
	}

	res, err := a.svc.UpdateInstanceOAM(c.Request.Context(), instanceOAMID, &req)
	if err != nil {
		common.Fail(c, err)
		return
	}

	common.OK(c, res)
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
		"id":               entity.ID,
		"name":             entity.Name,
		"description":      entity.Description,
		"business_unit_id": entity.BusinessUnitID,
		"ci_config_id":     entity.CIConfigID,
		"cd_config_id":     entity.CDConfigID,
		"instance_oam_id":  entity.InstanceOAMID,
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
