package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestListBusinessUnitCIConfigsRejectsInvalidBusinessUnitID(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "bad-id"}}
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/business-units/bad-id/ci-configs", nil)

	ListBusinessUnitCIConfigs(c)

	assertErrorMessage(t, w, "invalid business unit id")
}

func TestListBusinessUnitCIConfigsRejectsInvalidPage(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/business-units/1/ci-configs?page=x", nil)

	ListBusinessUnitCIConfigs(c)

	assertErrorMessage(t, w, "invalid page")
}

func TestListBusinessUnitCIConfigsRejectsInvalidPageSize(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/business-units/1/ci-configs?page_size=x", nil)

	ListBusinessUnitCIConfigs(c)

	assertErrorMessage(t, w, "invalid page_size")
}

func TestGetCIConfigRejectsInvalidID(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "bad-id"}}
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/ci-configs/bad-id", nil)

	GetCIConfig(c)

	assertErrorMessage(t, w, "invalid ci config id")
}

func TestDeleteCIConfigRejectsInvalidID(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "bad-id"}}
	c.Request = httptest.NewRequest(http.MethodDelete, "/api/v1/ci-configs/bad-id", nil)

	DeleteCIConfig(c)

	assertErrorMessage(t, w, "invalid ci config id")
}

func assertErrorMessage(t *testing.T, w *httptest.ResponseRecorder, expected string) {
	t.Helper()

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var body struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if body.Code != -1 {
		t.Fatalf("expected code -1, got %d", body.Code)
	}
	if body.Message != expected {
		t.Fatalf("expected message %q, got %q", expected, body.Message)
	}
}
