package api_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/richer421/q-metahub/conf"
	apphttp "github.com/richer421/q-metahub/http"
	"github.com/richer421/q-metahub/infra/mysql/dao"
	"github.com/richer421/q-metahub/pkg/logger"
	"github.com/richer421/q-metahub/pkg/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	gin.SetMode(gin.TestMode)
	logger.Init(conf.LogConfig{})
}

func TestListBusinessUnitCDConfigsRejectsInvalidBusinessUnitID(t *testing.T) {
	server := apphttp.NewServer()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/business-units/not-a-number/cd-configs", nil)
	resp := httptest.NewRecorder()

	server.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Contains(t, resp.Body.String(), "\"code\":-1")
	assert.Contains(t, resp.Body.String(), "invalid business unit id")
}

func TestDeleteCDConfigRejectsInvalidID(t *testing.T) {
	server := apphttp.NewServer()
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/cd-configs/not-a-number", nil)
	resp := httptest.NewRecorder()

	server.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Contains(t, resp.Body.String(), "\"code\":-1")
	assert.Contains(t, resp.Body.String(), "invalid cd config id")
}

func TestCreateBusinessUnitCDConfigCreatesRollingConfig(t *testing.T) {
	db, mock, err := testutil.NewMockDB()
	require.NoError(t, err)
	dao.SetDefault(db)

	mock.ExpectExec("INSERT INTO `cd_configs`").
		WithArgs(
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			"api-server-dev",
			int64(1),
			"cn-east",
			"dev",
			"helm",
			"",
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	server := apphttp.NewServer()
	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/business-units/1/cd-configs",
		bytes.NewBufferString(`{"name":"api-server-dev","release_region":"cn-east","release_env":"dev","deployment_mode":"rolling"}`),
	)
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	server.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Contains(t, resp.Body.String(), "\"name\":\"api-server-dev\"")
	assert.Contains(t, resp.Body.String(), "\"release_region\":\"cn-east\"")
	assert.Contains(t, resp.Body.String(), "\"release_env\":\"dev\"")
	assert.Contains(t, resp.Body.String(), "\"deployment_mode\":\"rolling\"")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteCDConfigRejectsReferencedConfig(t *testing.T) {
	db, mock, err := testutil.NewMockDB()
	require.NoError(t, err)
	dao.SetDefault(db)

	mock.ExpectQuery("SELECT count\\(\\*\\) FROM `deploy_plans`").
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	server := apphttp.NewServer()
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/cd-configs/1", nil)
	resp := httptest.NewRecorder()

	server.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Contains(t, resp.Body.String(), "\"code\":-1")
	assert.Contains(t, resp.Body.String(), "已被部署计划引用")
	assert.NoError(t, mock.ExpectationsWereMet())
}
