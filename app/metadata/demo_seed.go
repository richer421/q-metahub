package metadata

import (
	"context"
	"fmt"

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
		if len(fullSpec.CIConfigs) == 0 || len(fullSpec.CDConfigs) == 0 || len(fullSpec.InstanceConfigs) == 0 || len(fullSpec.DeployPlans) == 0 {
			return nil, fmt.Errorf("demo setup is incomplete")
		}
		if err := s.reconcileDemoSetup(ctx, fullSpec); err != nil {
			return nil, err
		}
		return &vo.DeployPlanAggregateDTO{
			Project:        fullSpec.Project,
			BusinessUnit:   fullSpec.BusinessUnit,
			CIConfig:       fullSpec.CIConfigs[0],
			CDConfig:       fullSpec.CDConfigs[0],
			InstanceConfig: fullSpec.InstanceConfigs[0],
			DeployPlan:     fullSpec.DeployPlans[0],
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
		InstanceConfig: vo.CreateInstanceConfigReq{
			Name:         "q-demo-dev",
			Env:          "dev",
			InstanceType: "deployment",
			Spec:         defaultDemoInstanceSpecMap(),
			AttachResources: map[string]any{
				"services": map[string]any{
					"q-demo": map[string]any{
						"metadata": map[string]any{
							"name": "q-demo",
						},
						"spec": map[string]any{
							"selector": map[string]any{
								"app": "q-demo",
							},
							"ports": []map[string]any{
								{
									"port":       defaultDemoContainerPort,
									"targetPort": defaultDemoContainerPort,
								},
							},
						},
					},
				},
			},
		},
		DeployPlan: vo.CreateDeployPlanReq{
			Name:        "q-demo-dev-plan",
			Description: "demo deploy plan",
		},
	}
	return s.CreateDeployPlanAggregate(ctx, req)
}

func (s *Service) reconcileDemoSetup(ctx context.Context, fullSpec *vo.BusinessUnitFullSpecDTO) error {
	if len(fullSpec.CIConfigs) == 0 || len(fullSpec.InstanceConfigs) == 0 {
		return nil
	}

	ciCfg := fullSpec.CIConfigs[0]
	instanceCfg := fullSpec.InstanceConfigs[0]
	ciNeedsUpdate := ciCfg.ImageRegistry != defaultDemoImageRegistry
	instanceNeedsUpdate := !hasRunnableDemoDeployment(instanceCfg.Spec)
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
			instanceModel, err := q.InstanceConfig.Where(dao.InstanceConfig.ID.Eq(instanceCfg.ID)).First()
			if err != nil {
				return err
			}
			instanceModel.Spec = defaultDemoInstanceSpec()
			if err := q.InstanceConfig.Save(instanceModel); err != nil {
				return err
			}
			fullSpec.InstanceConfigs[0].Spec = defaultDemoInstanceSpecMap()
		}
		return nil
	})
}

func hasRunnableDemoDeployment(spec map[string]any) bool {
	deployment, ok := spec["deployment"].(map[string]any)
	if !ok {
		return false
	}
	template, ok := deployment["template"].(map[string]any)
	if !ok {
		return false
	}
	podSpec, ok := template["spec"].(map[string]any)
	if !ok {
		return false
	}
	containers, ok := podSpec["containers"].([]any)
	return ok && len(containers) > 0
}

func defaultDemoInstanceSpecMap() map[string]any {
	return map[string]any{
		"deployment": map[string]any{
			"selector": map[string]any{
				"matchLabels": map[string]any{
					"app": "q-demo",
				},
			},
			"template": map[string]any{
				"metadata": map[string]any{
					"labels": map[string]any{
						"app": "q-demo",
					},
				},
				"spec": map[string]any{
					"containers": []map[string]any{
						{
							"name":  "q-demo",
							"image": fmt.Sprintf("%s/%s:latest", defaultDemoImageRegistry, "q-demo/q-demo"),
							"ports": []map[string]any{
								{
									"containerPort": defaultDemoContainerPort,
								},
							},
						},
					},
				},
			},
		},
	}
}

func defaultDemoInstanceSpec() model.InstanceSpec {
	var spec model.InstanceSpec
	_ = convertJSONMap(defaultDemoInstanceSpecMap(), &spec)
	return spec
}
