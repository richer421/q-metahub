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

func ListBusinessUnits(c *gin.Context) {
	page := 1
	if rawPage := strings.TrimSpace(c.DefaultQuery("page", "1")); rawPage != "" {
		parsed, err := strconv.Atoi(rawPage)
		if err == nil && parsed > 0 {
			page = parsed
		}
	}

	pageSize := 10
	if rawPageSize := strings.TrimSpace(c.DefaultQuery("page_size", "10")); rawPageSize != "" {
		parsed, err := strconv.Atoi(rawPageSize)
		if err == nil && parsed > 0 {
			pageSize = parsed
		}
	}

	res, err := metadata.App.ListBusinessUnits(c.Request.Context(), page, pageSize, strings.TrimSpace(c.Query("keyword")))
	if err != nil {
		common.Fail(c, err)
		return
	}

	common.OK(c, res)
}

func CreateBusinessUnit(c *gin.Context) {
	var req vo.CreateBusinessUnitReq
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, err)
		return
	}
	if strings.TrimSpace(req.Name) == "" {
		common.Fail(c, fmt.Errorf("name is required"))
		return
	}
	if req.ProjectID <= 0 {
		common.Fail(c, fmt.Errorf("project_id is required"))
		return
	}

	res, err := metadata.App.CreateBusinessUnit(c.Request.Context(), req)
	if err != nil {
		common.Fail(c, err)
		return
	}

	common.OK(c, res)
}

func UpdateBusinessUnit(c *gin.Context) {
	businessUnitID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		common.Fail(c, fmt.Errorf("invalid business unit id"))
		return
	}

	var req vo.UpdateBusinessUnitReq
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, err)
		return
	}
	if strings.TrimSpace(req.Name) == "" {
		common.Fail(c, fmt.Errorf("name is required"))
		return
	}

	res, err := metadata.App.UpdateBusinessUnit(c.Request.Context(), businessUnitID, req)
	if err != nil {
		common.Fail(c, err)
		return
	}

	common.OK(c, res)
}

func DeleteBusinessUnit(c *gin.Context) {
	businessUnitID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		common.Fail(c, fmt.Errorf("invalid business unit id"))
		return
	}

	if err := metadata.App.DeleteBusinessUnit(c.Request.Context(), businessUnitID); err != nil {
		common.Fail(c, err)
		return
	}

	common.OK(c, gin.H{})
}

func ListInstanceOAMTemplates(c *gin.Context) {
	common.OK(c, metadata.App.ListInstanceOAMTemplates(c.Request.Context()))
}

func ListBusinessUnitCIConfigs(c *gin.Context) {
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

	res, err := metadata.App.ListBusinessUnitCIConfigs(
		c.Request.Context(),
		businessUnitID,
		page,
		pageSize,
		strings.TrimSpace(c.Query("keyword")),
	)
	if err != nil {
		common.Fail(c, err)
		return
	}

	common.OK(c, res)
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

func GetCIConfig(c *gin.Context) {
	ciConfigID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		common.Fail(c, fmt.Errorf("invalid ci config id"))
		return
	}

	res, err := metadata.App.GetCIConfig(c.Request.Context(), ciConfigID)
	if err != nil {
		common.Fail(c, err)
		return
	}

	common.OK(c, res)
}

func CreateBusinessUnitCIConfig(c *gin.Context) {
	businessUnitID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		common.Fail(c, fmt.Errorf("invalid business unit id"))
		return
	}

	var req vo.CreateCIConfigReq
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, err)
		return
	}

	res, err := metadata.App.CreateBusinessUnitCIConfig(c.Request.Context(), businessUnitID, req)
	if err != nil {
		common.Fail(c, err)
		return
	}

	common.OK(c, res)
}

func UpdateCIConfig(c *gin.Context) {
	ciConfigID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		common.Fail(c, fmt.Errorf("invalid ci config id"))
		return
	}

	var req vo.UpdateCIConfigReq
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, err)
		return
	}

	res, err := metadata.App.UpdateCIConfig(c.Request.Context(), ciConfigID, req)
	if err != nil {
		common.Fail(c, err)
		return
	}

	common.OK(c, res)
}

func DeleteCIConfig(c *gin.Context) {
	ciConfigID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		common.Fail(c, fmt.Errorf("invalid ci config id"))
		return
	}

	if err := metadata.App.DeleteCIConfig(c.Request.Context(), ciConfigID); err != nil {
		common.Fail(c, err)
		return
	}

	common.OK(c, gin.H{})
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

func ListBusinessUnitCDConfigs(c *gin.Context) {
	businessUnitID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		common.Fail(c, fmt.Errorf("invalid business unit id"))
		return
	}

	var req vo.CDConfigListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		common.Fail(c, err)
		return
	}

	res, err := metadata.App.ListBusinessUnitCDConfigs(c.Request.Context(), businessUnitID, req)
	if err != nil {
		common.Fail(c, err)
		return
	}

	common.OK(c, res)
}

func GetCDConfig(c *gin.Context) {
	cdConfigID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		common.Fail(c, fmt.Errorf("invalid cd config id"))
		return
	}

	res, err := metadata.App.GetCDConfig(c.Request.Context(), cdConfigID)
	if err != nil {
		common.Fail(c, err)
		return
	}

	common.OK(c, res)
}

func CreateBusinessUnitCDConfig(c *gin.Context) {
	businessUnitID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		common.Fail(c, fmt.Errorf("invalid business unit id"))
		return
	}

	var req vo.UpsertCDConfigReq
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, err)
		return
	}

	res, err := metadata.App.CreateBusinessUnitCDConfig(c.Request.Context(), businessUnitID, req)
	if err != nil {
		common.Fail(c, err)
		return
	}

	common.OK(c, res)
}

func UpdateCDConfig(c *gin.Context) {
	cdConfigID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		common.Fail(c, fmt.Errorf("invalid cd config id"))
		return
	}

	var req vo.UpsertCDConfigReq
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, err)
		return
	}

	res, err := metadata.App.UpdateCDConfig(c.Request.Context(), cdConfigID, req)
	if err != nil {
		common.Fail(c, err)
		return
	}

	common.OK(c, res)
}

func DeleteCDConfig(c *gin.Context) {
	cdConfigID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		common.Fail(c, fmt.Errorf("invalid cd config id"))
		return
	}

	if err := metadata.App.DeleteCDConfig(c.Request.Context(), cdConfigID); err != nil {
		common.Fail(c, err)
		return
	}

	common.OK(c, gin.H{})
}
