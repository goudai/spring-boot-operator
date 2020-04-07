package global

import "sync"

var c *configureSpec

var once sync.Once

func GetGlobalConfig() *configureSpec {
	once.Do(func() {
		c = &configureSpec{}
	})
	return c
}

type configureSpec struct {
	ImageRepository  string
	RequestCpu       string
	LimitCpu         string
	RequestMemory    string
	LimitMemory      string
	LivenessPath     string
	ReadinessPath    string
	HostLogPath      string
	ShutdownPath     string
	Replicas         int32
	Port             int32
	Env              map[string]string
	ImagePullSecrets []string
	// failure-domain.beta.kubernetes.io/zone
	NodeAffinityKey string
	// In
	NodeAffinityValues []string
	// "cn-g", "cn-h", "cn-i"
	NodeAffinityOperator string
}
