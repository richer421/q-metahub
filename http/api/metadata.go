package api

import (
	"github.com/gin-gonic/gin"
	"github.com/richer421/q-metahub/app/metadata"

	"github.com/richer421/q-metahub/app/metadata/vo"
	"github.com/richer421/q-metahub/http/common"
)

func CreateInstanceOAM(c *gin.Context) {
	var req vo.CreateInstanceOAMReq
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, err)
		return
	}
	res, err := metadata.App.CreateInstanceOAM(c.Request.Context(), req)
	if err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, res)
}
