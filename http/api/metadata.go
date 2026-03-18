package api

import (
	"fmt"
	"strconv"
	"strings"

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

func ListInstanceOAMTemplates(c *gin.Context) {
	common.OK(c, metadata.App.ListInstanceOAMTemplates(c.Request.Context()))
}

func ListBusinessUnitInstanceOAMs(c *gin.Context) {
	businessUnitID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		common.Fail(c, fmt.Errorf("invalid business unit id"))
		return
	}

	page := 1
	if rawPage := strings.TrimSpace(c.DefaultQuery("page", "1")); rawPage != "" {
		page, err = strconv.Atoi(rawPage)
		if err != nil {
			common.Fail(c, fmt.Errorf("invalid page"))
			return
		}
	}

	pageSize := 10
	if rawPageSize := strings.TrimSpace(c.DefaultQuery("page_size", "10")); rawPageSize != "" {
		pageSize, err = strconv.Atoi(rawPageSize)
		if err != nil {
			common.Fail(c, fmt.Errorf("invalid page_size"))
			return
		}
	}

	res, err := metadata.App.ListBusinessUnitInstanceOAMs(
		c.Request.Context(),
		businessUnitID,
		page,
		pageSize,
		strings.TrimSpace(c.Query("env")),
		strings.TrimSpace(c.Query("keyword")),
	)
	if err != nil {
		common.Fail(c, err)
		return
	}

	common.OK(c, res)
}

func CreateBusinessUnitInstanceOAM(c *gin.Context) {
	businessUnitID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		common.Fail(c, fmt.Errorf("invalid business unit id"))
		return
	}

	var req vo.CreateInstanceOAMFromTemplateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, err)
		return
	}

	if strings.TrimSpace(req.Name) == "" || strings.TrimSpace(req.Env) == "" || strings.TrimSpace(req.TemplateKey) == "" {
		common.Fail(c, fmt.Errorf("name, env and template_key are required"))
		return
	}

	res, err := metadata.App.CreateBusinessUnitInstanceOAM(c.Request.Context(), businessUnitID, req)
	if err != nil {
		common.Fail(c, err)
		return
	}

	common.OK(c, res)
}

func UpdateInstanceOAM(c *gin.Context) {
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

	res, err := metadata.App.UpdateInstanceOAM(c.Request.Context(), instanceOAMID, req)
	if err != nil {
		common.Fail(c, err)
		return
	}

	common.OK(c, res)
}

func DeleteInstanceOAM(c *gin.Context) {
	instanceOAMID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		common.Fail(c, fmt.Errorf("invalid instance oam id"))
		return
	}

	if err := metadata.App.DeleteInstanceOAM(c.Request.Context(), instanceOAMID); err != nil {
		common.Fail(c, err)
		return
	}

	common.OK(c, gin.H{})
}
