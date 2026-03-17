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
	apiGroup := v1.Group("/metadata")
	apiGroup.POST("/instance-oams", api.CreateInstanceOAM)
}

func RegisterOpenModel(v1 *gin.RouterGroup) {
	apiGroup := v1.Group("/open-model")
	apiGroup.GET("/deploy-plans/:id", api.GetOpenModelDeployPlan)
}
