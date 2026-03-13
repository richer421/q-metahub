package router

import (
	"github.com/gin-gonic/gin"
	"github.com/richer421/q-metahub/http/api"
)

func RegisterV1(apiGroup *gin.RouterGroup) {
	v1 := apiGroup.Group("/v1")

	api.RegisterMetadataRoutes(v1)
}
