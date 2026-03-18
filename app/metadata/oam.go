package metadata

import (
	"github.com/richer421/q-metahub/app/metadata/vo"
	"github.com/richer421/q-metahub/infra/mysql/model"
)

const (
	defaultOAMAPIVersion  = "q.oam/v1alpha1"
	defaultOAMKind        = "InstanceApplication"
	defaultNetworkMode    = "k8s_service"
	defaultContainerImage = "IMAGE"
	defaultComponentName  = "app"
	defaultSchemaVersion  = "v1alpha1"
)

func convert2OAM(vo vo.CreateInstanceOAMReq) model.InstanceOAM {
	instanceName := vo.InstanceName
	if instanceName == "" {
		instanceName = vo.FrontendPayload.Basic.Container.Name
	}
	if instanceName == "" {
		instanceName = defaultComponentName
	}

	containerName := vo.FrontendPayload.Basic.Container.Name
	if containerName == "" {
		containerName = instanceName
	}

	containerImage := vo.FrontendPayload.Basic.Container.Image
	if containerImage == "" {
		containerImage = defaultContainerImage
	}

	mainContainerPorts := append([]int32(nil), vo.FrontendPayload.Basic.Container.Ports...)
	if len(mainContainerPorts) == 0 && vo.FrontendPayload.Extended != nil {
		mainContainerPorts = append(mainContainerPorts, vo.FrontendPayload.Extended.ServicePorts...)
	}

	networkMode := defaultNetworkMode
	if vo.FrontendPayload.Extended != nil && vo.FrontendPayload.Extended.NetworkMode != "" {
		networkMode = vo.FrontendPayload.Extended.NetworkMode
	}

	servicePorts := make([]int, 0)
	if vo.FrontendPayload.Extended != nil {
		for _, p := range vo.FrontendPayload.Extended.ServicePorts {
			servicePorts = append(servicePorts, int(p))
		}
	}
	if len(servicePorts) == 0 {
		for _, p := range mainContainerPorts {
			servicePorts = append(servicePorts, int(p))
		}
	}

	oamApp := model.OAMApplication{
		APIVersion: defaultOAMAPIVersion,
		Kind:       defaultOAMKind,
		Component: model.OAMPodComponent{
			Name: instanceName,
			Type: model.OAMComponentTypePod,
			Properties: model.OAMPodProperties{
				MainContainer: model.MainContainer{
					Container: model.Container{
						Name:  containerName,
						Image: containerImage,
						Ports: mainContainerPorts,
					},
				},
			},
		},
		Traits: &model.OAMTraits{
			Network: &model.NetworkTrait{
				Type: networkMode,
			},
		},
	}

	if len(servicePorts) > 0 {
		oamApp.Traits.Network.K8sServiceTrait = &model.K8sServiceTrait{Ports: servicePorts}
	}

	if vo.FrontendPayload.Basic.Replicas != nil && *vo.FrontendPayload.Basic.Replicas > 0 {
		oamApp.Traits.Scaling = &model.ScalingTrait{Replicas: *vo.FrontendPayload.Basic.Replicas}
	}

	return model.InstanceOAM{
		Name:           instanceName,
		BusinessUnitID: vo.BusinessUnitID,
		Env:            vo.Env,
		SchemaVersion:  defaultSchemaVersion,
		OAMApplication: oamApp,
	}
}

func convertToInstOAMVO(in model.InstanceOAM) vo.InstanceOAMVO {
	mainContainer := in.OAMApplication.Component.Properties.MainContainer
	container := convertMainContainerToVO(in.Name, mainContainer)
	replicas := extractReplicasFromTraits(in.OAMApplication.Traits)
	extended := buildExtendedFromTraits(in.OAMApplication.Traits, container.Ports)

	payload := vo.InstanceFrontendPayloadVO{
		Basic: vo.InstanceBasicVO{
			Replicas:  replicas,
			Container: container,
		},
		Extended: extended,
	}

	return vo.InstanceOAMVO{
		ID:              in.ID,
		CreatedAt:       in.CreatedAt,
		UpdatedAt:       in.UpdatedAt,
		Name:            in.Name,
		BusinessUnitID:  in.BusinessUnitID,
		Env:             in.Env,
		SchemaVersion:   in.SchemaVersion,
		OAMApplication:  in.OAMApplication,
		FrontendPayload: payload,
	}
}

func extractReplicasFromTraits(traits *model.OAMTraits) *int32 {
	if traits == nil || traits.Scaling == nil || traits.Scaling.Replicas <= 0 {
		return nil
	}
	value := traits.Scaling.Replicas
	return &value
}

func convertMainContainerToVO(instanceName string, mainContainer model.MainContainer) vo.InstanceContainerVO {
	containerName := mainContainer.Name
	if containerName == "" {
		containerName = instanceName
	}

	containerImage := mainContainer.Image
	if containerImage == "" {
		containerImage = defaultContainerImage
	}

	return vo.InstanceContainerVO{
		Name:  containerName,
		Image: containerImage,
		Ports: append([]int32(nil), mainContainer.Ports...),
	}
}

func buildExtendedFromTraits(traits *model.OAMTraits, fallbackPorts []int32) *vo.InstanceExtendedVO {
	networkMode := defaultNetworkMode
	servicePorts := make([]int32, 0)

	if traits != nil && traits.Network != nil {
		if traits.Network.Type != "" {
			networkMode = traits.Network.Type
		}
		servicePorts = append(servicePorts, convertServicePortsToInt32(traits.Network)...)
	}

	if len(servicePorts) == 0 {
		servicePorts = append(servicePorts, fallbackPorts...)
	}

	return &vo.InstanceExtendedVO{
		NetworkMode:  networkMode,
		ServicePorts: servicePorts,
	}
}

func convertServicePortsToInt32(network *model.NetworkTrait) []int32 {
	if network == nil || network.K8sServiceTrait == nil {
		return nil
	}

	ports := make([]int32, 0, len(network.K8sServiceTrait.Ports))
	for _, p := range network.K8sServiceTrait.Ports {
		ports = append(ports, int32(p))
	}
	return ports
}
