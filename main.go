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

package main

import (
	"encoding/json"
	"errors"
	"flag"
	"os"
	"spring-boot-operator/global"
	"strconv"
	"strings"

	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	springbootv1alpha1 "spring-boot-operator/api/v1alpha1"
	"spring-boot-operator/controllers"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)

	_ = springbootv1alpha1.AddToScheme(scheme)
	// +kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "enable-leader-election", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))
	onStart()

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: metricsAddr,
		Port:               9443,
		LeaderElection:     enableLeaderElection,
		LeaderElectionID:   "91a3bf26.qingmu.io",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if err = (&controllers.SpringBootApplicationReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("SpringBootApplication"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "SpringBootApplication")
		os.Exit(1)
	}
	// +kubebuilder:scaffold:builder

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
func onStart() {
	config := global.GetGlobalConfig()
	ImageRepository := os.Getenv("IMAGE_REPOSITORY")
	if ImageRepository == "" {
		setupLog.Error(errors.New("Not set env IMAGE_REPOSITORY"), "")
		//os.Exit(1)
	} else {
		setupLog.Info("Get user set env value ", "IMAGE_REPOSITORY", ImageRepository)
	}
	config.ImageRepository = ImageRepository

	RequestCpu := os.Getenv("REQUEST_CPU")
	if RequestCpu == "" {
		setupLog.Info("Not set env REQUEST_CPU, Using default request CPU [50m]")
		RequestCpu = "50m"
	} else {
		setupLog.Info("Get user set env value", "REQUEST_CPU", RequestCpu)
	}
	config.RequestCpu = RequestCpu

	LimitCpu := os.Getenv("LIMIT_CPU")
	if LimitCpu == "" {
		setupLog.Info("Not set env LIMIT_CPU, Using default limit CPU [unlimited]")
	} else {
		setupLog.Info("Get user set env value ", "LIMIT_CPU", LimitCpu)
	}
	config.LimitCpu = LimitCpu

	RequestMemory := os.Getenv("REQUEST_MEMORY")
	if RequestMemory == "" {
		setupLog.Info("Not set env REQUEST_MEMORY, Using default request Memory [2Gi]")
		RequestMemory = "2Gi"
	} else {
		setupLog.Info("Get user set env value", "REQUEST_MEMORY", RequestMemory)
	}
	config.RequestMemory = RequestMemory

	LimitMemory := os.Getenv("LIMIT_MEMORY")
	if LimitMemory == "" {
		setupLog.Info("Not set env LIMIT_MEMORY, Using default limit Memory [2Gi]")
		LimitMemory = "2Gi"
	} else {
		setupLog.Info("Get user set env value ", "LIMIT_MEMORY", LimitMemory)
	}
	config.LimitMemory = LimitMemory

	ReadinessPath := os.Getenv("READINESS_PATH")
	if ReadinessPath == "" {
		setupLog.Info("Not set env READINESS_PATH, Using default spring boot 2 readiness path  [/actuator/health]")
		ReadinessPath = "/actuator/health"
	} else {
		setupLog.Info("Get user set env value ", "READINESS_PATH", ReadinessPath)
	}
	config.ReadinessPath = ReadinessPath

	ShutdownPath := os.Getenv("SHUTDOWN_PATH")
	if ShutdownPath == "" {
		ShutdownPath = "/spring/shutdown"
		setupLog.Info("Not set env SHUTDOWN_PATH, Using default spring boot 2 shutdown path [" + ShutdownPath + "]")
	} else {
		setupLog.Info("Get user set env value ", "SHUTDOWN_PATH", ShutdownPath)
	}
	config.ShutdownPath = ShutdownPath

	LivenessPath := os.Getenv("LIVENESS_PATH")
	if LivenessPath == "" {
		LivenessPath = "/actuator/health"
		setupLog.Info("Not set env LIVENESS_PATH, Using default spring boot 2 liveness path [" + LivenessPath + "]")
	} else {
		setupLog.Info("Get user set env value ", "LIVENESS_PATH", LivenessPath)
	}
	config.LivenessPath = LivenessPath

	Replicas := os.Getenv("REPLICAS")
	if Replicas == "" {
		Replicas = "3"
		setupLog.Info("Not set env REPLICAS, Using default replicas [3]")
	} else {
		setupLog.Info("Get user set env value ", "REPLICAS", Replicas)
	}
	atoi, err := strconv.Atoi(Replicas)
	if err != nil {
		setupLog.Error(err, "Replicas is not a number")
		os.Exit(1)
	}
	config.Replicas = int32(atoi)

	HostLogPath := os.Getenv("HOST_LOG_PATH")
	if HostLogPath == "" {
		HostLogPath = "/var/applog"
		setupLog.Info("Not set env LIVENESS_PATH, Using default spring boot 2 host log  path [" + HostLogPath + "]")
	} else {
		setupLog.Info("Get user set env value ", "HOST_LOG_PATH", HostLogPath)
	}
	config.HostLogPath = HostLogPath

	imagePullSecrets := os.Getenv("IMAGE_PULL_SECRETS")
	if imagePullSecrets == "" {
		setupLog.Info("Not set env IMAGE_PULL_SECRETS")
	} else {
		setupLog.Info("Get user set env value ", "IMAGE_PULL_SECRETS", imagePullSecrets)
		config.ImagePullSecrets = strings.Split(imagePullSecrets, ",")
	}

	Env := os.Getenv("Env")
	if Env == "" {
		setupLog.Info("Not set env Env")
	} else {
		setupLog.Info("Get user set env value ", "Env", Env)
		for _, kv := range strings.Split(imagePullSecrets, ",") {
			kyarray := strings.Split(kv, "=")
			if len(kyarray) == 2 {
				config.Env[kyarray[0]] = kyarray[1]
			} else {
				config.Env[kyarray[0]] = ""
			}
		}
	}

	port := os.Getenv("SPRING_BOOT_DEFAULT_PORT")
	if port == "" {
		setupLog.Info("Not set env SPRING_BOOT_DEFAULT_PORT,using 8080 by default")
		port = "8080"
	} else {
		setupLog.Info("Get user set env value ", "SPRING_BOOT_DEFAULT_PORT", port)
	}
	if i, err := strconv.Atoi(port); err == nil {
		config.Port = int32(i)
	} else {
		setupLog.Info("Not parse set env port [" + port + "],using 8080 by default")
		config.Port = int32(8080)
	}

	NodeAffinityKey := os.Getenv("NODE_AFFINITY_KEY")
	if NodeAffinityKey == "" {
		setupLog.Info("Not set env NODE_AFFINITY_KEY")
	} else {
		setupLog.Info("Get user set env value ", "NODE_AFFINITY_KEY", NodeAffinityKey)
	}
	config.NodeAffinityKey = NodeAffinityKey

	NodeAffinityValues := os.Getenv("NODE_AFFINITY_VALUES")
	if NodeAffinityValues == "" {
		setupLog.Info("Not set env NODE_AFFINITY_VALUES")
	} else {
		setupLog.Info("Get user set env value ", "NODE_AFFINITY_VALUES", NodeAffinityValues)
	}
	config.NodeAffinityKey = NodeAffinityValues

	NodeAffinityOperator := os.Getenv("NODE_AFFINITY_OPERATOR")
	if NodeAffinityOperator == "" {
		setupLog.Info("Not set env NODE_AFFINITY_OPERATOR")
	} else {
		setupLog.Info("Get user set env value ", "NODE_AFFINITY_OPERATOR", NodeAffinityOperator)
	}
	config.NodeAffinityOperator = NodeAffinityOperator

	if marshal, err := json.Marshal(config); err != nil {
		setupLog.Error(err, "Replicas is not a number")
		os.Exit(1)
	} else {
		setupLog.Info("Global config " + string(marshal))
	}

}
