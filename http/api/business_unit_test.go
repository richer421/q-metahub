package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/richer421/q-metahub/infra/mysql/dao"
	"github.com/richer421/q-metahub/pkg/testutil"
)

type testResponse struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

type businessUnitPageResponse struct {
	Items    []businessUnitResponse `json:"items"`
	Total    int64                  `json:"total"`
	Page     int                    `json:"page"`
	PageSize int                    `json:"page_size"`
}

type businessUnitResponse struct {
	ID          int64     `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	ProjectID   int64     `json:"project_id"`
}

func TestListBusinessUnits(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, mock, err := testutil.NewMockDB()
	if err != nil {
		t.Fatalf("new mock db: %v", err)
	}
	dao.SetDefault(db)

	rows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name", "description", "project_id"}).
		AddRow(7, time.Date(2026, 3, 18, 10, 0, 0, 0, time.UTC), time.Date(2026, 3, 18, 11, 0, 0, 0, time.UTC), "api-server", "核心 REST API 服务", 101)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `business_units`")).
		WithArgs("%api%", "%api%", 2).
		WillReturnRows(rows)

	router := newBusinessUnitTestRouter()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/business-units?page=1&page_size=2&keyword=api", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", recorder.Code)
	}

	var response testResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if response.Code != 0 {
		t.Fatalf("unexpected response code: %d, body=%s", response.Code, recorder.Body.String())
	}

	var page businessUnitPageResponse
	if err := json.Unmarshal(response.Data, &page); err != nil {
		t.Fatalf("unmarshal data: %v", err)
	}

	if page.Total != 1 || page.Page != 1 || page.PageSize != 2 {
		t.Fatalf("unexpected page metadata: %+v", page)
	}
	if len(page.Items) != 1 || page.Items[0].ProjectID != 101 || page.Items[0].Name != "api-server" {
		t.Fatalf("unexpected items: %+v", page.Items)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestCreateBusinessUnit(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, mock, err := testutil.NewMockDB()
	if err != nil {
		t.Fatalf("new mock db: %v", err)
	}
	dao.SetDefault(db)

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `business_units`")).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), "member-center", "负责会员账户与权益", 101).
		WillReturnResult(sqlmock.NewResult(9, 1))

	router := newBusinessUnitTestRouter()
	body := bytes.NewBufferString(`{"name":"member-center","description":"负责会员账户与权益","project_id":101}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/business-units", body)
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	var response testResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if response.Code != 0 {
		t.Fatalf("unexpected response code: %d, body=%s", response.Code, recorder.Body.String())
	}

	var item businessUnitResponse
	if err := json.Unmarshal(response.Data, &item); err != nil {
		t.Fatalf("unmarshal data: %v", err)
	}
	if item.ID != 9 || item.ProjectID != 101 || item.Name != "member-center" {
		t.Fatalf("unexpected business unit: %+v", item)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestUpdateBusinessUnit(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, mock, err := testutil.NewMockDB()
	if err != nil {
		t.Fatalf("new mock db: %v", err)
	}
	dao.SetDefault(db)

	selectRows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name", "description", "project_id"}).
		AddRow(7, time.Date(2026, 3, 18, 9, 0, 0, 0, time.UTC), time.Date(2026, 3, 18, 9, 30, 0, 0, time.UTC), "old-name", "old-desc", 101)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `business_units` WHERE `business_units`.`id` = ? ORDER BY `business_units`.`id` LIMIT ?")).
		WithArgs(7, 1).
		WillReturnRows(selectRows)
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `business_units`")).
		WithArgs(
			time.Date(2026, 3, 18, 9, 0, 0, 0, time.UTC),
			time.Date(2026, 3, 18, 9, 30, 0, 0, time.UTC),
			"new-name",
			"new-desc",
			int64(101),
			int64(7),
			sqlmock.AnyArg(),
		).
		WillReturnResult(sqlmock.NewResult(0, 1))

	router := newBusinessUnitTestRouter()
	body := bytes.NewBufferString(`{"name":"new-name","description":"new-desc","project_id":999}`)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/business-units/7", body)
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	var response testResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if response.Code != 0 {
		t.Fatalf("unexpected response code: %d, body=%s", response.Code, recorder.Body.String())
	}

	var item businessUnitResponse
	if err := json.Unmarshal(response.Data, &item); err != nil {
		t.Fatalf("unmarshal data: %v", err)
	}
	if item.ProjectID != 101 || item.Name != "new-name" || item.Description != "new-desc" {
		t.Fatalf("unexpected updated business unit: %+v", item)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestDeleteBusinessUnitBlockedByDependencies(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, mock, err := testutil.NewMockDB()
	if err != nil {
		t.Fatalf("new mock db: %v", err)
	}
	dao.SetDefault(db)

	selectRows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name", "description", "project_id"}).
		AddRow(7, time.Date(2026, 3, 18, 9, 0, 0, 0, time.UTC), time.Date(2026, 3, 18, 9, 30, 0, 0, time.UTC), "api-server", "desc", 101)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `business_units` WHERE `business_units`.`id` = ? ORDER BY `business_units`.`id` LIMIT ?")).
		WithArgs(7, 1).
		WillReturnRows(selectRows)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT count(*) FROM `ci_configs` WHERE `ci_configs`.`business_unit_id` = ?")).
		WithArgs(7).
		WillReturnRows(sqlmock.NewRows([]string{"count(*)"}).AddRow(1))

	router := newBusinessUnitTestRouter()
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/business-units/7", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	var response testResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if response.Code != -1 {
		t.Fatalf("expected failure response, got: %s", recorder.Body.String())
	}
	if response.Message == "" {
		t.Fatalf("expected dependency error message")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func newBusinessUnitTestRouter() *gin.Engine {
	router := gin.New()
	v1 := router.Group("/api/v1")
	v1.GET("/business-units", ListBusinessUnits)
	v1.POST("/business-units", CreateBusinessUnit)
	v1.PUT("/business-units/:id", UpdateBusinessUnit)
	v1.DELETE("/business-units/:id", DeleteBusinessUnit)
	return router
}
