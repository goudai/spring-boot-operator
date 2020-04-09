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

package controllers

import (
	"context"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"strconv"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	springbootv1alpha1 "spring-boot-operator/api/v1alpha1"
)

// SpringBootApplicationReconciler reconciles a SpringBootApplication object
type SpringBootApplicationReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=springboot.qingmu.io,resources=springbootapplications,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=springboot.qingmu.io,resources=springbootapplications/status,verbs=get;update;patch

func (r *SpringBootApplicationReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("springbootapplication", req.NamespacedName)
	c := r.Client

	app := &springbootv1alpha1.SpringBootApplication{}
	err := r.Get(ctx, req.NamespacedName, app)
	if err != nil {
		log.Info(req.NamespacedName.Name + " is deleted .")
		return ctrl.Result{}, nil
	}
	name := app.GetObjectMeta().GetName()
	springBoot, err := app.Spec.SpringBoot.Check(name)
	if err != nil {
		log.Error(err, "check err ")
		return ctrl.Result{}, nil
	}
	log.Info("Received spring boot app,service is [" + name + ":" + strconv.Itoa(int(springBoot.Port)) + "], image is [" + springBoot.Image + "] ")
	// mate data
	labels := map[string]string{
		"k8s-app": name,
	}
	meta := metav1.ObjectMeta{
		Namespace: req.Namespace,
		Name:      name,
		Labels:    labels,
	}

	service := &v1.Service{ObjectMeta: meta}

	// Create or Update the Service
	//r.Get(ctx,req.NamespacedName,service)

	if err = controllerutil.SetControllerReference(app, service, r.Scheme); err != nil {
		return ctrl.Result{}, err
	}
	if op, err := controllerutil.CreateOrUpdate(ctx, c, service, func() error {
		// Deployment selector is immutable so we set this value only if
		// a new object is going to be created
		if service.ObjectMeta.CreationTimestamp.IsZero() {
			service.Spec.Selector = labels
		}

		//service.Spec = v1.ServiceSpec{
		//	Ports: []v1.ServicePort{
		//		{
		//			Name: name,
		//			Port: springBoot.Port,
		//		},
		//	},
		//	Selector: labels,
		//}
		service.Spec.Selector = labels
		service.Spec.Ports = []v1.ServicePort{
			{
				Name: name,
				Port: springBoot.Port,
			},
		}
		if springBoot.ClusterIp != "" {
			service.Spec.ClusterIP = springBoot.ClusterIp
		}
		return nil
	}); err != nil {
		log.Error(err, "Deployment reconcile failed")
		return ctrl.Result{}, nil
	} else {
		log.Info(string(op) + "  service success " + name)
	}

	deploy := &appsv1.Deployment{ObjectMeta: meta}
	// Create or Update the deployment
	if err := controllerutil.SetControllerReference(app, deploy, r.Scheme); err != nil {
		return ctrl.Result{}, err
	}
	if op, err := controllerutil.CreateOrUpdate(ctx, c, deploy, func() error {

		// Deployment selector is immutable so we set this value only if
		// a new object is going to be created
		selector := &metav1.LabelSelector{
			MatchLabels: labels,
		}
		if deploy.ObjectMeta.CreationTimestamp.IsZero() {
			deploy.Spec.Selector = selector
		}

		// update the Deployment pod template
		// pod
		port := intstr.IntOrString{IntVal: springBoot.Port}

		limitRlist := v1.ResourceList{
			"memory": resource.MustParse(springBoot.Resource.Memory.Limit),
		}
		if springBoot.Resource.Cpu.Limit != "" {
			limitRlist["cpu"] = resource.MustParse(springBoot.Resource.Cpu.Limit)
		}
		podSpec := &v1.PodSpec{
			Affinity: &v1.Affinity{
				PodAntiAffinity: &v1.PodAntiAffinity{
					PreferredDuringSchedulingIgnoredDuringExecution: []v1.WeightedPodAffinityTerm{
						{Weight: 1,
							PodAffinityTerm: v1.PodAffinityTerm{
								TopologyKey: "kubernetes.io/hostname",
								LabelSelector: &metav1.LabelSelector{
									MatchExpressions: []metav1.LabelSelectorRequirement{
										{
											Key:      "k8s-app",
											Operator: "In",
											Values:   []string{name},
										},
									},
								},
							},
						},
					},
				},
			},
			Containers: []v1.Container{
				{
					Name:            name,
					Image:           springBoot.Image,
					ImagePullPolicy: "IfNotPresent",
					Ports:           []v1.ContainerPort{{ContainerPort: springBoot.Port}},
					Env:             springBoot.Env,
					Resources: v1.ResourceRequirements{
						Requests: v1.ResourceList{
							"cpu":    resource.MustParse(springBoot.Resource.Cpu.Request),
							"memory": resource.MustParse(springBoot.Resource.Memory.Request),
						},
						Limits: limitRlist,
					},
					Lifecycle: &v1.Lifecycle{
						PreStop: &v1.Handler{
							HTTPGet: &v1.HTTPGetAction{
								Path: springBoot.Path.Shutdown,
								Port: port,
							},
						},
					},
					LivenessProbe: &v1.Probe{
						Handler: v1.Handler{
							HTTPGet: &v1.HTTPGetAction{
								Path: springBoot.Path.Liveness,
								Port: port,
							},
						},
					},
					ReadinessProbe: &v1.Probe{
						Handler: v1.Handler{
							HTTPGet: &v1.HTTPGetAction{
								Path: springBoot.Path.Readiness,
								Port: port,
							},
						},
					},
				},
			},
		}

		if len(springBoot.ImagePullSecrets) > 0 {
			references := []v1.LocalObjectReference{}
			for _, secret := range springBoot.ImagePullSecrets {
				references = append(references, v1.LocalObjectReference{Name: secret})
			}
			podSpec.ImagePullSecrets = references
		}

		nodeAffinity := springBoot.NodeAffinity
		if nodeAffinity.Key != "" {
			podSpec.Affinity.NodeAffinity = &v1.NodeAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: &v1.NodeSelector{
					NodeSelectorTerms: []v1.NodeSelectorTerm{
						{
							MatchExpressions: []v1.NodeSelectorRequirement{
								{
									Key:      nodeAffinity.Key,
									Operator: v1.NodeSelectorOperator(nodeAffinity.Operator),
									Values:   nodeAffinity.Values,
								},
							},
						},
					},
				},
			}
		}

		hostLog := springBoot.Path.HostLog
		if hostLog != "" {
			volumeName := "applogpath"
			hostPathType := new(v1.HostPathType)
			*hostPathType = v1.HostPathDirectoryOrCreate
			podSpec.Volumes = []v1.Volume{
				{
					Name: volumeName,
					VolumeSource: v1.VolumeSource{
						HostPath: &v1.HostPathVolumeSource{
							Path: hostLog,
							Type: hostPathType,
						},
					},
				},
			}
			container := &podSpec.Containers[0]
			container.VolumeMounts = []v1.VolumeMount{{
				Name:      volumeName,
				ReadOnly:  false,
				MountPath: hostLog,
			}}

		}

		revisionHistoryLimit := int32(10)
		deploy.Spec = appsv1.DeploymentSpec{
			Replicas:             &springBoot.Replicas,
			RevisionHistoryLimit: &revisionHistoryLimit,
			Template: v1.PodTemplateSpec{
				ObjectMeta: meta,
				Spec:       *podSpec,
			},
			Strategy: appsv1.DeploymentStrategy{
				Type:          "RollingUpdate",
				RollingUpdate: &appsv1.RollingUpdateDeployment{},
			},
			Selector: selector,
		}
		return nil
	}); err != nil {
		log.Error(err, "Deployment reconcile failed")
		return ctrl.Result{}, nil
	} else {
		log.Info(string(op) + " " + name + " deployment ")
	}

	return ctrl.Result{}, nil
}

func (r *SpringBootApplicationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&springbootv1alpha1.SpringBootApplication{}).
		Complete(r)
}
