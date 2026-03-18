package router

import (
	"github.com/gin-gonic/gin"
	"github.com/richer421/q-metahub/http/api"
)

func RegisterV1(apiGroup *gin.RouterGroup) {
	v1 := apiGroup.Group("/v1")

	RegisterMetadataRoutes(v1)
	RegisterOpenModel(v1)
}

func RegisterMetadataRoutes(v1 *gin.RouterGroup) {
	legacyGroup := v1.Group("/metadata")
	legacyGroup.POST("/instance-oams", api.CreateInstanceOAM)

	v1.GET("/instance-oam-templates", api.ListInstanceOAMTemplates)
	v1.GET("/business-units/:id/instance-oams", api.ListBusinessUnitInstanceOAMs)
	v1.POST("/business-units/:id/instance-oams", api.CreateBusinessUnitInstanceOAM)
	v1.GET("/business-units/:id/cd-configs", api.ListBusinessUnitCDConfigs)
	v1.POST("/business-units/:id/cd-configs", api.CreateBusinessUnitCDConfig)
	v1.PUT("/instance-oams/:id", api.UpdateInstanceOAM)
	v1.DELETE("/instance-oams/:id", api.DeleteInstanceOAM)
	v1.GET("/cd-configs/:id", api.GetCDConfig)
	v1.PUT("/cd-configs/:id", api.UpdateCDConfig)
	v1.DELETE("/cd-configs/:id", api.DeleteCDConfig)
}

func RegisterOpenModel(v1 *gin.RouterGroup) {
	apiGroup := v1.Group("/open-model")
	apiGroup.GET("/deploy-plans/:id", api.GetOpenModelDeployPlan)
}
