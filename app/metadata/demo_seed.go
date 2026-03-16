package metadata

import (
	"context"
	"fmt"
	"strings"

	"github.com/richer421/q-metahub/app/metadata/vo"
	"github.com/richer421/q-metahub/infra/mysql/dao"
	"github.com/richer421/q-metahub/infra/mysql/model"
)

const defaultDemoImageRegistry = "localhost:30180"
const defaultDemoContainerPort = 8080

func (s *Service) SeedDemoSetup(ctx context.Context) (*vo.DeployPlanAggregateDTO, error) {
	project, err := dao.Q.WithContext(ctx).Project.Where(dao.Project.Name.Eq("q-demo-project")).First()
	if err == nil && project != nil {
		businessUnit, err := dao.Q.WithContext(ctx).BusinessUnit.Where(dao.BusinessUnit.ProjectID.Eq(project.ID)).First()
		if err != nil {
			return nil, err
		}
		fullSpec, err := s.GetBusinessUnitFullSpec(ctx, businessUnit.ID)
		if err != nil {
			return nil, err
		}
		if len(fullSpec.CIConfigs) == 0 || len(fullSpec.CDConfigs) == 0 || len(fullSpec.InstanceOAMs) == 0 || len(fullSpec.DeployPlans) == 0 {
			return nil, fmt.Errorf("demo setup is incomplete")
		}
		if err := s.reconcileDemoSetup(ctx, fullSpec); err != nil {
			return nil, err
		}
		return &vo.DeployPlanAggregateDTO{
			Project:      fullSpec.Project,
			BusinessUnit: fullSpec.BusinessUnit,
			CIConfig:     fullSpec.CIConfigs[0],
			CDConfig:     fullSpec.CDConfigs[0],
			InstanceOAM:  fullSpec.InstanceOAMs[0],
			DeployPlan:   fullSpec.DeployPlans[0],
		}, nil
	}

	req := &vo.CreateDeployPlanAggregateReq{
		Project: vo.CreateProjectReq{
			GitID:   1001,
			Name:    "q-demo-project",
			RepoURL: "https://github.com/richer421/q-demo.git",
		},
		BusinessUnit: vo.CreateBusinessUnitReq{
			Name:        "q-demo",
			Description: "demo business unit",
		},
		CIConfig: vo.CreateCIConfigReq{
			Name:          "q-demo-ci",
			ImageRegistry: defaultDemoImageRegistry,
			ImageRepo:     "q-demo/q-demo",
			ImageTagRule: map[string]any{
				"type": "commit",
			},
			BuildSpec: map[string]any{
				"branch":          "main",
				"dockerfile_path": "./Dockerfile",
				"docker_context":  ".",
				"make_command":    "build",
			},
		},
		CDConfig: vo.CreateCDConfigReq{
			Name:         "q-demo-cd",
			RenderEngine: "helm",
			ValuesYAML:   "replicaCount: 1\n",
			ReleaseStrategy: map[string]any{
				"deployment_mode": "rolling",
				"batch_rule": map[string]any{
					"batch_count":  1,
					"batch_ratio":  []float64{1},
					"trigger_type": "auto",
					"interval":     0,
				},
			},
			GitOps: map[string]any{
				"enabled":       true,
				"repo_url":      "https://github.com/richer421/q-demo-gitops.git",
				"branch":        "main",
				"app_root":      "apps",
				"manifest_root": "manifests",
			},
		},
		InstanceOAM: vo.CreateInstanceOAMReq{
			Name:            "q-demo-dev",
			Env:             "dev",
			SchemaVersion:   "v1alpha1",
			OAMApplication:  defaultDemoOAMApplicationMap(),
			FrontendPayload: defaultDemoFrontendPayloadMap(),
		},
		DeployPlan: vo.CreateDeployPlanReq{
			Name:        "q-demo-dev-plan",
			Description: "demo deploy plan",
		},
	}
	return s.CreateDeployPlanAggregate(ctx, req)
}

func (s *Service) reconcileDemoSetup(ctx context.Context, fullSpec *vo.BusinessUnitFullSpecDTO) error {
	if len(fullSpec.CIConfigs) == 0 || len(fullSpec.InstanceOAMs) == 0 {
		return nil
	}

	ciCfg := fullSpec.CIConfigs[0]
	instanceCfg := fullSpec.InstanceOAMs[0]
	ciNeedsUpdate := ciCfg.ImageRegistry != defaultDemoImageRegistry
	instanceNeedsUpdate := !hasRunnableDemoPod(instanceCfg.OAMApplication)
	if !ciNeedsUpdate && !instanceNeedsUpdate {
		return nil
	}

	return dao.Q.Transaction(func(tx *dao.Query) error {
		q := tx.WithContext(ctx)
		if ciNeedsUpdate {
			ciModel, err := q.CIConfig.Where(dao.CIConfig.ID.Eq(ciCfg.ID)).First()
			if err != nil {
				return err
			}
			ciModel.ImageRegistry = defaultDemoImageRegistry
			if err := q.CIConfig.Save(ciModel); err != nil {
				return err
			}
			fullSpec.CIConfigs[0].ImageRegistry = defaultDemoImageRegistry
		}
		if instanceNeedsUpdate {
			instanceModel, err := q.InstanceOAM.Where(dao.InstanceOAM.ID.Eq(instanceCfg.ID)).First()
			if err != nil {
				return err
			}
			instanceModel.OAMApplication = defaultDemoOAMApplication()
			instanceModel.FrontendPayload = defaultDemoFrontendPayload()
			if err := q.InstanceOAM.Save(instanceModel); err != nil {
				return err
			}
			fullSpec.InstanceOAMs[0].OAMApplication = defaultDemoOAMApplicationMap()
			fullSpec.InstanceOAMs[0].FrontendPayload = defaultDemoFrontendPayloadMap()
		}
		return nil
	})
}

func hasRunnableDemoPod(oamApp map[string]any) bool {
	component, ok := oamApp["component"].(map[string]any)
	if !ok {
		return false
	}
	properties, ok := component["properties"].(map[string]any)
	if !ok {
		return false
	}
	mainContainer, ok := properties["mainContainer"].(map[string]any)
	if !ok {
		return false
	}

	name, _ := mainContainer["name"].(string)
	return strings.TrimSpace(name) != ""
}

func defaultDemoOAMApplicationMap() map[string]any {
	return map[string]any{
		"apiVersion": "q.oam/v1alpha1",
		"kind":       "InstanceApplication",
		"component": map[string]any{
			"name": "q-demo",
			"type": model.OAMComponentTypePod,
			"properties": map[string]any{
				"mainContainer": map[string]any{
					"name":  "q-demo",
					"image": fmt.Sprintf("%s/%s:latest", defaultDemoImageRegistry, "q-demo/q-demo"),
					"ports": []int{defaultDemoContainerPort},
				},
			},
		},
		"traits": map[string]any{
			"network": map[string]any{
				"type": "k8s_service",
				"k8sServiceTrait": map[string]any{
					"ports": []int{defaultDemoContainerPort},
				},
			},
			"config": map[string]any{
				"env": []map[string]any{
					{
						"key":   "ENV",
						"value": "dev",
					},
				},
			},
		},
	}
}

func defaultDemoOAMApplication() model.OAMApplication {
	var app model.OAMApplication
	_ = convertJSONMap(defaultDemoOAMApplicationMap(), &app)
	return app
}

func defaultDemoFrontendPayloadMap() map[string]any {
	return map[string]any{
		"basic": map[string]any{
			"name": "q-demo-dev",
			"env":  "dev",
		},
		"extended": map[string]any{
			"network_mode": "k8s_service",
			"ports":        []int{defaultDemoContainerPort},
		},
	}
}

func defaultDemoFrontendPayload() model.InstanceOAMPayload {
	var payload model.InstanceOAMPayload
	_ = convertJSONMap(defaultDemoFrontendPayloadMap(), &payload)
	return payload
}
