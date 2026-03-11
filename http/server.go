package http

import (
	"github.com/richer421/q-metahub/conf"
	"github.com/richer421/q-metahub/http/middleware"
	"github.com/richer421/q-metahub/http/router"
	appotel "github.com/richer421/q-metahub/pkg/otel"

	"github.com/gin-gonic/gin"
)

func NewServer() *gin.Engine {
	r := gin.New()

	if conf.C.OTel.Enabled {
		r.Use(middleware.OTel())
	}
	r.Use(middleware.Logger(), middleware.Recovery())

	// Prometheus /metrics endpoint
	if conf.C.OTel.Enabled && conf.C.OTel.Prometheus.Enabled {
		if h := appotel.PrometheusHandler(); h != nil {
			r.GET(conf.C.OTel.Prometheus.Path, gin.WrapH(h))
		}
	}

	router.Register(r)

	return r
}
