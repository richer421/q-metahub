package metadata

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/richer421/q-metahub/app/metadata/vo"
	"github.com/richer421/q-metahub/conf"
	"github.com/richer421/q-metahub/infra/mysql/dao"
	"github.com/richer421/q-metahub/infra/mysql/model"
)

const (
	defaultTemplateContainerPort = 8080
	defaultFrontendStatus        = "stopped"
	defaultFrontendType          = "Deployment"
	defaultFrontendInstanceType  = "deployment"
)

type frontendInstanceView struct {
	ID              string                   `json:"id,omitempty"`
	BusinessUnitID  string                   `json:"buId,omitempty"`
	Name            string                   `json:"name"`
	Env             string                   `json:"env"`
	Type            string                   `json:"type,omitempty"`
	InstanceType    string                   `json:"instanceType,omitempty"`
	Replicas        int32                    `json:"replicas,omitempty"`
	ReadyReplicas   int32                    `json:"readyReplicas,omitempty"`
	CPU             string                   `json:"cpu,omitempty"`
	Memory          string                   `json:"memory,omitempty"`
	YAML            string                   `json:"yaml,omitempty"`
	Spec            *frontendInstanceSpec    `json:"spec,omitempty"`
	AttachResources *frontendAttachResources `json:"attachResources,omitempty"`
	Pods            []any                    `json:"pods,omitempty"`
	Status          string                   `json:"status,omitempty"`
}

type frontendInstanceSpec struct {
	Deployment *frontendDeploymentSpec `json:"deployment,omitempty"`
}

type frontendDeploymentSpec struct {
	Replicas int32                 `json:"replicas,omitempty"`
	Template *frontendPodTemplate  `json:"template,omitempty"`
}

type frontendPodTemplate struct {
	Spec *frontendPodTemplateSpec `json:"spec,omitempty"`
}

type frontendPodTemplateSpec struct {
	Containers []frontendContainerView `json:"containers,omitempty"`
}

type frontendContainerView struct {
	Name      string                     `json:"name"`
	Image     string                     `json:"image,omitempty"`
	Command   []string                   `json:"command,omitempty"`
	Args      []string                   `json:"args,omitempty"`
	Ports     []frontendContainerPort    `json:"ports,omitempty"`
	Env       []frontendEnvVar           `json:"env,omitempty"`
	Resources *frontendContainerResource `json:"resources,omitempty"`
}

type frontendContainerPort struct {
	ContainerPort int32 `json:"containerPort"`
}

type frontendEnvVar struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type frontendContainerResource struct {
	Requests *frontendResourceValues `json:"requests,omitempty"`
	Limits   *frontendResourceValues `json:"limits,omitempty"`
}

type frontendResourceValues struct {
	CPU    string `json:"cpu,omitempty"`
	Memory string `json:"memory,omitempty"`
}

type frontendAttachResources struct {
	ConfigMaps map[string]frontendAttachResource `json:"configMaps,omitempty"`
	Secrets    map[string]frontendAttachResource `json:"secrets,omitempty"`
	Services   map[string]frontendAttachResource `json:"services,omitempty"`
}

type frontendAttachResource struct {
	Metadata *frontendAttachMetadata `json:"metadata,omitempty"`
	Spec     *frontendAttachSpec     `json:"spec,omitempty"`
}

type frontendAttachMetadata struct {
	Name string `json:"name,omitempty"`
}

type frontendAttachSpec struct {
	Ports []frontendServicePort `json:"ports,omitempty"`
}

type frontendServicePort struct {
	Port       int `json:"port"`
	TargetPort int `json:"targetPort"`
}

func (s *app) ListInstanceOAMTemplates(_ context.Context) []vo.InstanceOAMTemplateDTO {
	items := make([]vo.InstanceOAMTemplateDTO, 0, len(conf.C.InstanceOAMTemplates))
	for _, item := range conf.C.InstanceOAMTemplates {
		items = append(items, vo.InstanceOAMTemplateDTO{
			Key:           strings.TrimSpace(item.Key),
			Name:          strings.TrimSpace(item.Name),
			Description:   strings.TrimSpace(item.Description),
			Replicas:      item.Replicas,
			CPURequest:    strings.TrimSpace(item.CPURequest),
			CPULimit:      strings.TrimSpace(item.CPULimit),
			MemoryRequest: strings.TrimSpace(item.MemoryRequest),
			MemoryLimit:   strings.TrimSpace(item.MemoryLimit),
		})
	}
	return items
}

func (s *app) ListBusinessUnitInstanceOAMs(ctx context.Context, businessUnitID int64, page int, pageSize int, env string, keyword string) (*vo.InstanceOAMPageDTO, error) {
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	query := dao.Q.WithContext(ctx).
		InstanceOAM.
		Where(dao.InstanceOAM.BusinessUnitID.Eq(businessUnitID))

	if normalizedEnv := strings.TrimSpace(env); normalizedEnv != "" {
		query = query.Where(dao.InstanceOAM.Env.Eq(normalizedEnv))
	}
	if normalizedKeyword := strings.TrimSpace(keyword); normalizedKeyword != "" {
		query = query.Where(dao.InstanceOAM.Name.Like("%" + normalizedKeyword + "%"))
	}

	offset := (page - 1) * pageSize
	rows, total, err := query.Order(dao.InstanceOAM.UpdatedAt.Desc()).FindByPage(offset, pageSize)
	if err != nil {
		return nil, err
	}

	items := make([]vo.InstanceOAMDTO, 0, len(rows))
	for _, row := range rows {
		item, convErr := toInstanceOAMDTO(row)
		if convErr != nil {
			return nil, convErr
		}
		items = append(items, item)
	}

	return &vo.InstanceOAMPageDTO{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (s *app) CreateBusinessUnitInstanceOAM(ctx context.Context, businessUnitID int64, req vo.CreateInstanceOAMFromTemplateReq) (*vo.InstanceOAMDTO, error) {
	templateConfig, ok := findInstanceOAMTemplate(req.TemplateKey)
	if !ok {
		return nil, fmt.Errorf("instance template %q not found", strings.TrimSpace(req.TemplateKey))
	}

	view := buildTemplateInstanceView(businessUnitID, strings.TrimSpace(req.Name), strings.TrimSpace(req.Env), templateConfig)
	entity := &model.InstanceOAM{
		Name:           view.Name,
		BusinessUnitID: businessUnitID,
		Env:            view.Env,
		SchemaVersion:  defaultSchemaVersion,
		OAMApplication: buildOAMFromFrontendInstanceView(view),
	}

	if err := dao.Q.WithContext(ctx).InstanceOAM.Create(entity); err != nil {
		return nil, err
	}

	dto, err := toInstanceOAMDTO(entity)
	if err != nil {
		return nil, err
	}
	return &dto, nil
}

func (s *app) UpdateInstanceOAM(ctx context.Context, instanceOAMID int64, req vo.UpdateInstanceOAMReq) (*vo.InstanceOAMDTO, error) {
	entity, err := dao.Q.WithContext(ctx).InstanceOAM.Where(dao.InstanceOAM.ID.Eq(instanceOAMID)).First()
	if err != nil {
		return nil, err
	}

	nextOAM, nextName, nextEnv, err := buildUpdatedOAM(req, entity)
	if err != nil {
		return nil, err
	}

	schemaVersion := strings.TrimSpace(req.SchemaVersion)
	if schemaVersion == "" {
		schemaVersion = entity.SchemaVersion
	}
	if schemaVersion == "" {
		schemaVersion = defaultSchemaVersion
	}

	entity.Name = nextName
	entity.Env = nextEnv
	entity.SchemaVersion = schemaVersion
	entity.OAMApplication = nextOAM

	if err := dao.Q.WithContext(ctx).InstanceOAM.Save(entity); err != nil {
		return nil, err
	}

	dto, err := toInstanceOAMDTO(entity)
	if err != nil {
		return nil, err
	}
	return &dto, nil
}

func (s *app) DeleteInstanceOAM(ctx context.Context, instanceOAMID int64) error {
	entity, err := dao.Q.WithContext(ctx).InstanceOAM.Where(dao.InstanceOAM.ID.Eq(instanceOAMID)).First()
	if err != nil {
		return err
	}

	_, err = dao.Q.WithContext(ctx).InstanceOAM.Delete(entity)
	return err
}

func buildUpdatedOAM(req vo.UpdateInstanceOAMReq, current *model.InstanceOAM) (model.OAMApplication, string, string, error) {
	if view, ok, err := parseFrontendInstanceView(req.FrontendPayload); err != nil {
		return model.OAMApplication{}, "", "", err
	} else if ok {
		if strings.TrimSpace(view.Name) == "" {
			view.Name = strings.TrimSpace(req.Name)
		}
		if strings.TrimSpace(view.Name) == "" {
			view.Name = current.Name
		}
		if strings.TrimSpace(view.Env) == "" {
			view.Env = strings.TrimSpace(req.Env)
		}
		if strings.TrimSpace(view.Env) == "" {
			view.Env = current.Env
		}
		return buildOAMFromFrontendInstanceView(view), view.Name, view.Env, nil
	}

	var nextOAM model.OAMApplication
	if err := convertJSONMap(req.OAMApplication, &nextOAM); err != nil {
		return model.OAMApplication{}, "", "", err
	}

	nextName := strings.TrimSpace(req.Name)
	if nextName == "" {
		nextName = current.Name
	}
	nextEnv := strings.TrimSpace(req.Env)
	if nextEnv == "" {
		nextEnv = current.Env
	}
	return nextOAM, nextName, nextEnv, nil
}

func toInstanceOAMDTO(entity *model.InstanceOAM) (vo.InstanceOAMDTO, error) {
	oamApplication, err := toJSONMap(entity.OAMApplication)
	if err != nil {
		return vo.InstanceOAMDTO{}, err
	}

	instanceView := buildFrontendInstanceView(*entity)
	instanceMap, err := toJSONMap(instanceView)
	if err != nil {
		return vo.InstanceOAMDTO{}, err
	}

	return vo.InstanceOAMDTO{
		ID:             entity.ID,
		BusinessUnitID: entity.BusinessUnitID,
		Name:           entity.Name,
		Env:            entity.Env,
		SchemaVersion:  entity.SchemaVersion,
		OAMApplication: oamApplication,
		FrontendPayload: map[string]any{
			"basic": map[string]any{
				"instance": instanceMap,
			},
		},
	}, nil
}

func buildTemplateInstanceView(businessUnitID int64, name string, env string, templateConfig conf.InstanceOAMTemplateConfig) frontendInstanceView {
	return frontendInstanceView{
		BusinessUnitID: fmt.Sprintf("%d", businessUnitID),
		Name:           name,
		Env:            env,
		Type:           defaultFrontendType,
		InstanceType:   defaultFrontendInstanceType,
		Replicas:       templateConfig.Replicas,
		ReadyReplicas:  0,
		CPU:            strings.TrimSpace(templateConfig.CPURequest),
		Memory:         strings.TrimSpace(templateConfig.MemoryRequest),
		Status:         defaultFrontendStatus,
		Pods:           []any{},
		Spec: &frontendInstanceSpec{
			Deployment: &frontendDeploymentSpec{
				Replicas: templateConfig.Replicas,
				Template: &frontendPodTemplate{
					Spec: &frontendPodTemplateSpec{
						Containers: []frontendContainerView{
							{
								Name:  name,
								Image: defaultContainerImage,
								Ports: []frontendContainerPort{
									{ContainerPort: defaultTemplateContainerPort},
								},
								Env: []frontendEnvVar{
									{Name: "APP_ENV", Value: env},
								},
								Resources: &frontendContainerResource{
									Requests: &frontendResourceValues{
										CPU:    strings.TrimSpace(templateConfig.CPURequest),
										Memory: strings.TrimSpace(templateConfig.MemoryRequest),
									},
									Limits: &frontendResourceValues{
										CPU:    strings.TrimSpace(templateConfig.CPULimit),
										Memory: strings.TrimSpace(templateConfig.MemoryLimit),
									},
								},
							},
						},
					},
				},
			},
		},
		AttachResources: &frontendAttachResources{
			ConfigMaps: map[string]frontendAttachResource{},
			Secrets:    map[string]frontendAttachResource{},
			Services: map[string]frontendAttachResource{
				name: {
					Metadata: &frontendAttachMetadata{Name: name},
					Spec: &frontendAttachSpec{
						Ports: []frontendServicePort{
							{Port: defaultTemplateContainerPort, TargetPort: defaultTemplateContainerPort},
						},
					},
				},
			},
		},
	}
}

func buildFrontendInstanceView(entity model.InstanceOAM) frontendInstanceView {
	mainContainer := entity.OAMApplication.Component.Properties.MainContainer
	replicas := extractReplicasFromTraits(entity.OAMApplication.Traits)
	replicaCount := int32(1)
	if replicas != nil && *replicas > 0 {
		replicaCount = *replicas
	}

	servicePorts := extractServicePorts(entity.OAMApplication.Traits)
	attachResources := &frontendAttachResources{
		ConfigMaps: map[string]frontendAttachResource{},
		Secrets:    map[string]frontendAttachResource{},
		Services:   map[string]frontendAttachResource{},
	}
	if len(servicePorts) > 0 {
		ports := make([]frontendServicePort, 0, len(servicePorts))
		for _, port := range servicePorts {
			ports = append(ports, frontendServicePort{
				Port:       port,
				TargetPort: port,
			})
		}
		attachResources.Services[entity.Name] = frontendAttachResource{
			Metadata: &frontendAttachMetadata{Name: entity.Name},
			Spec:     &frontendAttachSpec{Ports: ports},
		}
	}

	return frontendInstanceView{
		ID:             fmt.Sprintf("%d", entity.ID),
		BusinessUnitID: fmt.Sprintf("%d", entity.BusinessUnitID),
		Name:           entity.Name,
		Env:            entity.Env,
		Type:           defaultFrontendType,
		InstanceType:   defaultFrontendInstanceType,
		Replicas:       replicaCount,
		ReadyReplicas:  0,
		CPU:            getCPURequest(mainContainer.Resources),
		Memory:         getMemoryRequest(mainContainer.Resources),
		YAML:           buildFrontendInstanceYAML(entity.Name, entity.Env, replicaCount, mainContainer, attachResources),
		Spec: &frontendInstanceSpec{
			Deployment: &frontendDeploymentSpec{
				Replicas: replicaCount,
				Template: &frontendPodTemplate{
					Spec: &frontendPodTemplateSpec{
						Containers: []frontendContainerView{
							{
								Name:      mainContainer.Name,
								Image:     mainContainer.Image,
								Command:   commandToSlice(mainContainer.Command),
								Args:      append([]string(nil), mainContainer.Args...),
								Ports:     toFrontendContainerPorts(mainContainer.Ports),
								Env:       toFrontendEnvVars(mainContainer.Env),
								Resources: toFrontendResources(mainContainer.Resources),
							},
						},
					},
				},
			},
		},
		AttachResources: attachResources,
		Pods:            []any{},
		Status:          defaultFrontendStatus,
	}
}

func parseFrontendInstanceView(payload map[string]any) (frontendInstanceView, bool, error) {
	basicRaw, ok := payload["basic"]
	if !ok {
		return frontendInstanceView{}, false, nil
	}

	basic, ok := basicRaw.(map[string]any)
	if !ok {
		return frontendInstanceView{}, false, fmt.Errorf("frontend_payload.basic is invalid")
	}

	instanceRaw, ok := basic["instance"]
	if !ok {
		return frontendInstanceView{}, false, nil
	}

	data, err := json.Marshal(instanceRaw)
	if err != nil {
		return frontendInstanceView{}, false, err
	}

	var view frontendInstanceView
	if err := json.Unmarshal(data, &view); err != nil {
		return frontendInstanceView{}, false, err
	}
	return view, true, nil
}

func buildOAMFromFrontendInstanceView(view frontendInstanceView) model.OAMApplication {
	container := frontendContainerView{
		Name:  view.Name,
		Image: defaultContainerImage,
	}
	replicas := view.Replicas

	if view.Spec != nil && view.Spec.Deployment != nil {
		if view.Spec.Deployment.Replicas > 0 {
			replicas = view.Spec.Deployment.Replicas
		}
		if view.Spec.Deployment.Template != nil && view.Spec.Deployment.Template.Spec != nil && len(view.Spec.Deployment.Template.Spec.Containers) > 0 {
			container = view.Spec.Deployment.Template.Spec.Containers[0]
		}
	}
	if replicas <= 0 {
		replicas = 1
	}
	if strings.TrimSpace(container.Name) == "" {
		container.Name = view.Name
	}
	if strings.TrimSpace(container.Image) == "" {
		container.Image = defaultContainerImage
	}

	servicePorts := extractServicePortsFromView(view, container)

	return model.OAMApplication{
		APIVersion: defaultOAMAPIVersion,
		Kind:       defaultOAMKind,
		Metadata: &model.OAMObjectMeta{
			Name: view.Name,
		},
		Component: model.OAMPodComponent{
			Name: view.Name,
			Type: model.OAMComponentTypePod,
			Properties: model.OAMPodProperties{
				MainContainer: model.MainContainer{
					Container: model.Container{
						Name:    container.Name,
						Image:   container.Image,
						Command: joinCommand(container.Command),
						Args:    append([]string(nil), container.Args...),
						Env:     toModelEnvVars(container.Env),
						Ports:   toModelPorts(container.Ports),
						Resources: &model.ResourceQuota{
							Cpu: &model.CpuQuota{
								Request: firstNonEmpty(getRequestCPUFromView(container.Resources), view.CPU),
								Limit:   getLimitCPUFromView(container.Resources),
							},
							Memory: &model.MemoryQuota{
								Request: firstNonEmpty(getRequestMemoryFromView(container.Resources), view.Memory),
								Limit:   getLimitMemoryFromView(container.Resources),
							},
						},
					},
				},
			},
		},
		Traits: &model.OAMTraits{
			Scaling: &model.ScalingTrait{
				Replicas: replicas,
			},
			Network: &model.NetworkTrait{
				Type: "k8s_service",
				K8sServiceTrait: &model.K8sServiceTrait{
					Ports: servicePorts,
				},
			},
		},
	}
}

func extractServicePortsFromView(view frontendInstanceView, container frontendContainerView) []int {
	if view.AttachResources != nil {
		for _, service := range view.AttachResources.Services {
			if service.Spec == nil {
				continue
			}
			ports := make([]int, 0, len(service.Spec.Ports))
			for _, port := range service.Spec.Ports {
				if port.Port > 0 {
					ports = append(ports, port.Port)
				}
			}
			if len(ports) > 0 {
				return ports
			}
		}
	}

	ports := make([]int, 0, len(container.Ports))
	for _, port := range container.Ports {
		if port.ContainerPort > 0 {
			ports = append(ports, int(port.ContainerPort))
		}
	}
	return ports
}

func extractServicePorts(traits *model.OAMTraits) []int {
	if traits == nil || traits.Network == nil || traits.Network.K8sServiceTrait == nil {
		return nil
	}
	return append([]int(nil), traits.Network.K8sServiceTrait.Ports...)
}

func toFrontendContainerPorts(ports []int32) []frontendContainerPort {
	items := make([]frontendContainerPort, 0, len(ports))
	for _, port := range ports {
		items = append(items, frontendContainerPort{ContainerPort: port})
	}
	return items
}

func toFrontendEnvVars(envVars []model.OAMEnvVar) []frontendEnvVar {
	items := make([]frontendEnvVar, 0, len(envVars))
	for _, item := range envVars {
		items = append(items, frontendEnvVar{Name: item.Key, Value: item.Value})
	}
	return items
}

func toFrontendResources(resources *model.ResourceQuota) *frontendContainerResource {
	if resources == nil {
		return nil
	}
	return &frontendContainerResource{
		Requests: &frontendResourceValues{
			CPU:    getCPURequest(resources),
			Memory: getMemoryRequest(resources),
		},
		Limits: &frontendResourceValues{
			CPU:    getCPULimit(resources),
			Memory: getMemoryLimit(resources),
		},
	}
}

func toModelEnvVars(envVars []frontendEnvVar) []model.OAMEnvVar {
	items := make([]model.OAMEnvVar, 0, len(envVars))
	for _, item := range envVars {
		if strings.TrimSpace(item.Name) == "" {
			continue
		}
		items = append(items, model.OAMEnvVar{Key: item.Name, Value: item.Value})
	}
	return items
}

func toModelPorts(ports []frontendContainerPort) []int32 {
	items := make([]int32, 0, len(ports))
	for _, item := range ports {
		if item.ContainerPort > 0 {
			items = append(items, item.ContainerPort)
		}
	}
	return items
}

func getCPURequest(resources *model.ResourceQuota) string {
	if resources == nil || resources.Cpu == nil {
		return ""
	}
	return resources.Cpu.Request
}

func getCPULimit(resources *model.ResourceQuota) string {
	if resources == nil || resources.Cpu == nil {
		return ""
	}
	return resources.Cpu.Limit
}

func getMemoryRequest(resources *model.ResourceQuota) string {
	if resources == nil || resources.Memory == nil {
		return ""
	}
	return resources.Memory.Request
}

func getMemoryLimit(resources *model.ResourceQuota) string {
	if resources == nil || resources.Memory == nil {
		return ""
	}
	return resources.Memory.Limit
}

func getRequestCPUFromView(resources *frontendContainerResource) string {
	if resources == nil || resources.Requests == nil {
		return ""
	}
	return resources.Requests.CPU
}

func getLimitCPUFromView(resources *frontendContainerResource) string {
	if resources == nil || resources.Limits == nil {
		return ""
	}
	return resources.Limits.CPU
}

func getRequestMemoryFromView(resources *frontendContainerResource) string {
	if resources == nil || resources.Requests == nil {
		return ""
	}
	return resources.Requests.Memory
}

func getLimitMemoryFromView(resources *frontendContainerResource) string {
	if resources == nil || resources.Limits == nil {
		return ""
	}
	return resources.Limits.Memory
}

func commandToSlice(command string) []string {
	trimmed := strings.TrimSpace(command)
	if trimmed == "" {
		return nil
	}
	return []string{trimmed}
}

func joinCommand(parts []string) string {
	return strings.TrimSpace(strings.Join(parts, " "))
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func toJSONMap(value any) (map[string]any, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}

	var out map[string]any
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func findInstanceOAMTemplate(key string) (conf.InstanceOAMTemplateConfig, bool) {
	normalizedKey := strings.TrimSpace(key)
	for _, item := range conf.C.InstanceOAMTemplates {
		if strings.EqualFold(strings.TrimSpace(item.Key), normalizedKey) {
			return item, true
		}
	}
	return conf.InstanceOAMTemplateConfig{}, false
}

func buildFrontendInstanceYAML(name string, env string, replicas int32, mainContainer model.MainContainer, attachResources *frontendAttachResources) string {
	doc := map[string]any{
		"name":          name,
		"env":           env,
		"instance_type": defaultFrontendInstanceType,
		"spec": map[string]any{
			"deployment": map[string]any{
				"replicas": replicas,
				"template": map[string]any{
					"spec": map[string]any{
						"containers": []map[string]any{
							{
								"name":  mainContainer.Name,
								"image": mainContainer.Image,
								"ports": buildContainerPortDocs(mainContainer.Ports),
								"env":   buildEnvDocs(mainContainer.Env),
								"resources": map[string]any{
									"requests": map[string]any{
										"cpu":    getCPURequest(mainContainer.Resources),
										"memory": getMemoryRequest(mainContainer.Resources),
									},
									"limits": map[string]any{
										"cpu":    getCPULimit(mainContainer.Resources),
										"memory": getMemoryLimit(mainContainer.Resources),
									},
								},
							},
						},
					},
				},
			},
		},
		"attach_resources": map[string]any{
			"configMaps": map[string]any{},
			"secrets":    map[string]any{},
			"services":   buildServiceDocs(attachResources),
		},
	}

	data, err := yaml.Marshal(doc)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

func buildContainerPortDocs(ports []int32) []map[string]any {
	items := make([]map[string]any, 0, len(ports))
	for _, port := range ports {
		items = append(items, map[string]any{"containerPort": port})
	}
	return items
}

func buildEnvDocs(envVars []model.OAMEnvVar) []map[string]any {
	items := make([]map[string]any, 0, len(envVars))
	for _, item := range envVars {
		items = append(items, map[string]any{
			"name":  item.Key,
			"value": item.Value,
		})
	}
	return items
}

func buildServiceDocs(attachResources *frontendAttachResources) map[string]any {
	result := map[string]any{}
	if attachResources == nil {
		return result
	}

	for name, service := range attachResources.Services {
		serviceDoc := map[string]any{
			"metadata": map[string]any{
				"name": name,
			},
		}
		if service.Metadata != nil && strings.TrimSpace(service.Metadata.Name) != "" {
			serviceDoc["metadata"] = map[string]any{
				"name": service.Metadata.Name,
			}
		}
		if service.Spec != nil && len(service.Spec.Ports) > 0 {
			ports := make([]map[string]any, 0, len(service.Spec.Ports))
			for _, port := range service.Spec.Ports {
				ports = append(ports, map[string]any{
					"port":       port.Port,
					"targetPort": port.TargetPort,
				})
			}
			serviceDoc["spec"] = map[string]any{"ports": ports}
		}
		result[name] = serviceDoc
	}
	return result
}

func convertJSONMap(input map[string]any, target any) error {
	if input == nil {
		return nil
	}

	data, err := json.Marshal(input)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, target)
}
