package metadata

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/richer421/q-metahub/app/metadata/vo"
	"github.com/richer421/q-metahub/infra/mysql/model"
)

func TestBuildInstanceOAMDeriveOAMFromFrontendPayload(t *testing.T) {
	replicas := int32(3)
	req := vo.CreateInstanceOAMReq{
		Name:          "demo-dev",
		Env:           "dev",
		SchemaVersion: "v1alpha1",
		FrontendPayload: vo.InstanceFrontendPayloadVO{
			Basic: vo.InstanceBasicVO{
				Name: "demo",
				Env:  "dev",
			},
			Extended: vo.InstanceExtendedVO{
				NetworkMode: "k8s_service",
				Ports:       []int{8080},
			},
			Advanced: vo.InstanceAdvancedVO{
				Replicas: &replicas,
			},
		},
	}

	instance := buildInstanceOAM(req, 100)
	require.NotNil(t, instance)

	assert.Equal(t, "demo", instance.OAMApplication.Component.Name)
	assert.Equal(t, int32(8080), instance.OAMApplication.Component.Properties.MainContainer.Ports[0])
	require.NotNil(t, instance.OAMApplication.Traits)
	require.NotNil(t, instance.OAMApplication.Traits.Network)
	assert.Equal(t, []int{8080}, instance.OAMApplication.Traits.Network.K8sServiceTrait.Ports)
	require.NotNil(t, instance.OAMApplication.Traits.Scaling)
	assert.Equal(t, int32(3), instance.OAMApplication.Traits.Scaling.Replicas)
}

func TestToInstanceOAMDTODeriveFrontendPayloadFromOAM(t *testing.T) {
	in := &model.InstanceOAM{
		BaseModel: model.BaseModel{
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Name:           "demo-dev",
		BusinessUnitID: 100,
		Env:            "dev",
		SchemaVersion:  "v1alpha1",
		OAMApplication: model.OAMApplication{
			APIVersion: "q.oam/v1alpha1",
			Kind:       "InstanceApplication",
			Component: model.OAMPodComponent{
				Name: "demo",
				Type: "pod",
				Properties: model.OAMPodProperties{
					MainContainer: model.MainContainer{
						Container: model.Container{
							Name:  "demo",
							Image: "IMAGE",
							Ports: []int32{8080},
						},
					},
				},
			},
			Traits: &model.OAMTraits{
				Network: &model.NetworkTrait{
					Type: "k8s_service",
					K8sServiceTrait: &model.K8sServiceTrait{
						Ports: []int{8080},
					},
				},
				Scaling: &model.ScalingTrait{Replicas: 2},
			},
		},
	}

	dto := toInstanceOAMDTO(in)
	assert.Equal(t, "demo", dto.FrontendPayload.Basic.Name)
	assert.Equal(t, "dev", dto.FrontendPayload.Basic.Env)
	assert.Equal(t, "k8s_service", dto.FrontendPayload.Extended.NetworkMode)
	assert.Equal(t, []int{8080}, dto.FrontendPayload.Extended.Ports)
	require.NotNil(t, dto.FrontendPayload.Advanced.Replicas)
	assert.Equal(t, int32(2), *dto.FrontendPayload.Advanced.Replicas)
}
