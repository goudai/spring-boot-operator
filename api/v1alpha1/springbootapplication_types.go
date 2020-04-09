/*
Copyright 2020 qingmu.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	"fmt"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"spring-boot-operator/global"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// SpringBootApplicationSpec defines the desired state of SpringBootApplication
type SpringBootApplicationSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// The spring boot body
	SpringBoot SpringBoot `json:"springBoot"`
}

type SpringBoot struct {
	// The spring boot application Port
	Port int32 `json:"port,omitempty"`
	// The spring boot application Image
	// If the value is empty,using fmt.Sprintf("%s/%s:%s", config.ImageRepository, Name, s.Version) by default
	Image string `json:"image,omitempty"`
	// The spring boot application image version. this is required
	Version string `json:"version,omitempty"`
	// The spring boot application service ip (kube-proxy cluster ip). "" by default
	ClusterIp string `json:"clusterIp,omitempty"`
	// The spring boot application replicas. 3 by default
	Replicas int32 `json:"replicas,omitempty"`
	// The spring boot application Resource(Cpu,Memory)
	// 2Gi Request Memory by default.
	// 2Gi Limit Memory by default.
	// 100m Request Cpu by default.
	// Un limit Cpu by default.
	Resource ResourceSpec `json:"resource,omitempty"`
	// The spring boot application path
	// Liveness and Readiness  is '/actuator/health' by  default
	// HostLog is '/var/applog' by default
	// Shutdown is '/spring/shutdown' by default
	Path PathSpec `json:"path,omitempty"`
	// The pull image secrets.
	ImagePullSecrets []string `json:"imagePullSecrets,omitempty"`
	// The spring boot application env.
	Env []v1.EnvVar `json:"env,omitempty"`

	NodeAffinity NodeAffinitySpec `json:"nodeAffinity,omitempty"`
}

type NodeAffinitySpec struct {
	Key      string   `json:"key"`
	Operator string   `json:"operator"`
	Values   []string `json:"values"`
}

type ResourceSpec struct {
	// Cpu resource
	Cpu CpuSpec `json:"cpu,omitempty"`
	// Memory resource
	Memory MemorySpec `json:"memory,omitempty"`
}

type CpuSpec struct {
	// 100m Request Cpu by default.
	Request string `json:"request,omitempty"`
	// Un limit Cpu by default.
	Limit string `json:"limit,omitempty"`
}

type MemorySpec struct {
	// 2Gi Request Memory by default.
	Request string `json:"request,omitempty"`
	// 2Gi Limit Memory by default.
	Limit string `json:"limit,omitempty"`
}

type PathSpec struct {
	// Liveness  is '/actuator/health' by  default
	Liveness string `json:"liveness,omitempty"`
	//  Readiness  is '/actuator/health' by  default
	Readiness string `json:"readiness,omitempty"`
	// HostLog is '/var/applog' by default
	HostLog string `json:"hostLog,omitempty"`
	// Shutdown is '/spring/shutdown' by default
	Shutdown string `json:"shutdown,omitempty"`
}

// SpringBootApplicationStatus defines the observed state of SpringBootApplication
type SpringBootApplicationStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true

// SpringBootApplication is the Schema for the springbootapplications API
type SpringBootApplication struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SpringBootApplicationSpec   `json:"spec,omitempty"`
	Status SpringBootApplicationStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// SpringBootApplicationList contains a list of SpringBootApplication
type SpringBootApplicationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SpringBootApplication `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SpringBootApplication{}, &SpringBootApplicationList{})
}

func (s *SpringBoot) Check(Name string) (*SpringBoot, error) {
	config := global.GetGlobalConfig()
	if s.Image == "" {
		image := fmt.Sprintf("%s/%s:%s", config.ImageRepository, Name, s.Version)
		s.Image = image
	}
	if s.Path.Shutdown == "" {
		s.Path.Shutdown = config.ShutdownPath
	}

	if s.Path.HostLog == "" {
		s.Path.HostLog = config.HostLogPath
	}
	if s.Path.Liveness == "" {
		s.Path.Liveness = config.LivenessPath
	}
	if s.Path.Readiness == "" {
		s.Path.Readiness = config.ReadinessPath
	}
	if s.Replicas == 0 {
		s.Replicas = config.Replicas
	}
	if s.Resource.Cpu.Limit == "" {
		s.Resource.Cpu.Limit = config.LimitCpu
	}
	if s.Resource.Cpu.Request == "" {
		s.Resource.Cpu.Request = config.RequestCpu
	}

	if s.Resource.Memory.Limit == "" {
		s.Resource.Memory.Limit = config.LimitMemory
	}
	if s.Resource.Memory.Request == "" {
		s.Resource.Memory.Request = config.RequestMemory
	}
	if s.ClusterIp == "" {
		//s.ClusterIp = "None"
	}
	if s.Env == nil {
		s.Env = []v1.EnvVar{}
	}
	envStringMap := make(map[string]string)
	for _, v := range s.Env {
		envStringMap[v.Name] = v.Value
	}
	if len(config.Env) > 0 {
		for k, v := range config.Env {
			if _, ok := envStringMap[k]; !ok {
				s.Env = append(s.Env, v1.EnvVar{
					Name:  k,
					Value: v,
				})
			}
		}
	}

	if s.Port == 0 {
		s.Port = config.Port
	}

	if len(config.ImagePullSecrets) > 0 {
		for _, secret := range config.ImagePullSecrets {
			s.ImagePullSecrets = append(s.ImagePullSecrets, secret)
		}
	}

	return s, nil
}
