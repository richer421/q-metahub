package api

import (
	"encoding/json"
	"strconv"

	"github.com/gin-gonic/gin"

	appmetadata "github.com/richer421/q-metahub/app/metadata"
	"github.com/richer421/q-metahub/app/metadata/vo"
	"github.com/richer421/q-metahub/http/common"
	openmodeloam "github.com/richer421/q-metahub/pkg/openModel/oam"
)

type MetadataAPI struct {
	svc *appmetadata.Service
}

func NewMetadataAPI() *MetadataAPI {
	return &MetadataAPI{svc: appmetadata.NewService()}
}

func RegisterMetadataRoutes(v1 *gin.RouterGroup) {
	api := NewMetadataAPI()

	v1.POST("/instance-oams", api.CreateInstanceOAM)
	v1.GET("/deploy-plans/:id", api.GetDeployPlan)
}

func (a *MetadataAPI) CreateInstanceOAM(c *gin.Context) {
	var req vo.CreateInstanceOAMReq
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, req)
}

func (a *MetadataAPI) GetDeployPlan(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		common.Fail(c, err)
		return
	}

	res, err := a.svc.GetDeployPlanFullSpec(c.Request.Context(), id)
	if err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, toOpenModelDeployPlanSpec(res))
}

func toOpenModelDeployPlanSpec(in *vo.DeployPlanAggregateDTO) *openmodeloam.DeployPlanSpecDTO {
	if in == nil {
		return nil
	}
	return &openmodeloam.DeployPlanSpecDTO{
		Project: openmodeloam.ProjectDTO{
			ID:      in.Project.ID,
			Name:    in.Project.Name,
			RepoURL: in.Project.RepoURL,
		},
		BusinessUnit: openmodeloam.BusinessUnitDTO{
			ID:   in.BusinessUnit.ID,
			Name: in.BusinessUnit.Name,
		},
		CDConfig: openmodeloam.CDConfigDTO{
			ID:              in.CDConfig.ID,
			GitOps:          in.CDConfig.GitOps,
			ReleaseStrategy: in.CDConfig.ReleaseStrategy,
		},
		InstanceOAM: openmodeloam.InstanceOAMDTO{
			ID:             in.InstanceOAM.ID,
			Env:            in.InstanceOAM.Env,
			SchemaVersion:  in.InstanceOAM.SchemaVersion,
			OAMApplication: toOpenModelOAMApplication(in.InstanceOAM.OAMApplication),
		},
		DeployPlan: openmodeloam.DeployPlanDTO{
			ID:            in.DeployPlan.ID,
			CDConfigID:    in.DeployPlan.CDConfigID,
			InstanceOAMID: in.DeployPlan.InstanceOAMID,
		},
	}
}

func toOpenModelOAMApplication(in any) openmodeloam.OAMApplication {
	data, err := json.Marshal(in)
	if err != nil {
		return openmodeloam.OAMApplication{}
	}
	var out openmodeloam.OAMApplication
	if err := json.Unmarshal(data, &out); err != nil {
		return openmodeloam.OAMApplication{}
	}
	return out
}
