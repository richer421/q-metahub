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

// instanceFormModel is backend view model bridging UI form and persisted OAM.
type instanceFormModel struct {
	Name        string
	Env         string
	NetworkMode string
	Ports       []int
	Replicas    *int32
}

func frontendVOToForm(payload vo.InstanceFrontendPayloadVO, fallbackName, fallbackEnv string) instanceFormModel {
	name := payload.Basic.Name
	if name == "" {
		name = fallbackName
	}
	if name == "" {
		name = defaultComponentName
	}

	env := payload.Basic.Env
	if env == "" {
		env = fallbackEnv
	}

	networkMode := payload.Extended.NetworkMode
	if networkMode == "" {
		networkMode = defaultNetworkMode
	}

	return instanceFormModel{
		Name:        name,
		Env:         env,
		NetworkMode: networkMode,
		Ports:       append([]int(nil), payload.Extended.Ports...),
		Replicas:    payload.Advanced.Replicas,
	}
}

func formToFrontendVO(form instanceFormModel) vo.InstanceFrontendPayloadVO {
	return vo.InstanceFrontendPayloadVO{
		Basic: vo.InstanceBasicVO{
			Name: form.Name,
			Env:  form.Env,
		},
		Extended: vo.InstanceExtendedVO{
			NetworkMode: form.NetworkMode,
			Ports:       append([]int(nil), form.Ports...),
		},
		Advanced: vo.InstanceAdvancedVO{
			Replicas: form.Replicas,
		},
	}
}

func formToOAM(form instanceFormModel) model.OAMApplication {
	mainPorts := make([]int32, 0, len(form.Ports))
	for _, port := range form.Ports {
		mainPorts = append(mainPorts, int32(port))
	}

	oam := model.OAMApplication{
		APIVersion: defaultOAMAPIVersion,
		Kind:       defaultOAMKind,
		Component: model.OAMPodComponent{
			Name: form.Name,
			Type: model.OAMComponentTypePod,
			Properties: model.OAMPodProperties{
				MainContainer: model.MainContainer{
					Container: model.Container{
						Name:  form.Name,
						Image: defaultContainerImage,
						Ports: mainPorts,
					},
				},
			},
		},
		Traits: &model.OAMTraits{
			Network: &model.NetworkTrait{
				Type:            form.NetworkMode,
				K8sServiceTrait: &model.K8sServiceTrait{Ports: append([]int(nil), form.Ports...)},
			},
		},
	}

	if len(form.Ports) == 0 {
		oam.Traits.Network.K8sServiceTrait = nil
	}
	if form.Replicas != nil && *form.Replicas > 0 {
		oam.Traits.Scaling = &model.ScalingTrait{Replicas: *form.Replicas}
	}
	return oam
}

func oamToForm(oam model.OAMApplication, fallbackName, fallbackEnv string) instanceFormModel {
	name := oam.Component.Name
	if name == "" {
		name = oam.Component.Properties.MainContainer.Name
	}
	if name == "" {
		name = fallbackName
	}
	if name == "" {
		name = defaultComponentName
	}

	networkMode := defaultNetworkMode
	if oam.Traits != nil && oam.Traits.Network != nil && oam.Traits.Network.Type != "" {
		networkMode = oam.Traits.Network.Type
	}

	form := instanceFormModel{
		Name:        name,
		Env:         fallbackEnv,
		NetworkMode: networkMode,
		Ports:       oamPorts(oam),
	}
	if oam.Traits != nil && oam.Traits.Scaling != nil && oam.Traits.Scaling.Replicas > 0 {
		replicas := oam.Traits.Scaling.Replicas
		form.Replicas = &replicas
	}
	return form
}

func normalizeOAMApplication(oam model.OAMApplication, fallbackName string) model.OAMApplication {
	if oam.APIVersion == "" {
		oam.APIVersion = defaultOAMAPIVersion
	}
	if oam.Kind == "" {
		oam.Kind = defaultOAMKind
	}
	if oam.Component.Type == "" {
		oam.Component.Type = model.OAMComponentTypePod
	}
	if oam.Component.Name == "" {
		oam.Component.Name = oam.Component.Properties.MainContainer.Name
	}
	if oam.Component.Name == "" {
		oam.Component.Name = fallbackName
	}
	if oam.Component.Name == "" {
		oam.Component.Name = defaultComponentName
	}
	if oam.Component.Properties.MainContainer.Name == "" {
		oam.Component.Properties.MainContainer.Name = oam.Component.Name
	}
	if oam.Component.Properties.MainContainer.Image == "" {
		oam.Component.Properties.MainContainer.Image = defaultContainerImage
	}
	if len(oam.Component.Properties.MainContainer.Ports) == 0 {
		for _, port := range servicePortsFromOAM(oam) {
			oam.Component.Properties.MainContainer.Ports = append(oam.Component.Properties.MainContainer.Ports, int32(port))
		}
	}
	if oam.Traits != nil && oam.Traits.Network != nil && oam.Traits.Network.Type == "" {
		oam.Traits.Network.Type = defaultNetworkMode
	}
	return oam
}

func oamPorts(oam model.OAMApplication) []int {
	if len(oam.Component.Properties.MainContainer.Ports) > 0 {
		out := make([]int, 0, len(oam.Component.Properties.MainContainer.Ports))
		for _, port := range oam.Component.Properties.MainContainer.Ports {
			out = append(out, int(port))
		}
		return out
	}
	return servicePortsFromOAM(oam)
}

func servicePortsFromOAM(oam model.OAMApplication) []int {
	if oam.Traits == nil || oam.Traits.Network == nil || oam.Traits.Network.K8sServiceTrait == nil {
		return nil
	}
	return append([]int(nil), oam.Traits.Network.K8sServiceTrait.Ports...)
}
