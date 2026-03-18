package api

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/richer421/q-metahub/app/metadata"
	"github.com/richer421/q-metahub/http/common"
	openmodeloam "github.com/richer421/q-metahub/pkg/openModel/oam"
)

func GetOpenModelDeployPlan(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		common.Fail(c, err)
		return
	}

	res, err := metadata.App.GetDeployPlan(c.Request.Context(), id)
	if err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, openmodeloam.ToOpenModelDeployPlan(res))
}
