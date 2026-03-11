package main

import (
	"github.com/richer421/q-metahub/infra/mysql/model"

	"gorm.io/gen"
)

func main() {
	g := gen.NewGenerator(gen.Config{
		OutPath: "./infra/mysql/dao",
		Mode:    gen.WithDefaultQuery,
	})

	g.ApplyBasic(
		model.Project{},
		model.BusinessUnit{},
		model.CIConfig{},
		model.CDConfig{},
		model.InstanceConfig{},
		model.DeployPlan{},
		model.Dependency{},
		model.DependencyBinding{},
	)

	g.Execute()
}
