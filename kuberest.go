package podexecutor

import (
	"encoding/base64"
	"fmt"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

type k8sRest struct {
	client rest.Interface
	config *rest.Config
}

func newK8SRestFromExistingConfig(restCfg *rest.Config) (*k8sRest, error) {
	k8sClientSet, err := kubernetes.NewForConfig(restCfg)
	if err != nil {
		return nil, fmt.Errorf("could not create kubernetes client set for rest config: %w", err)
	}

	return &k8sRest{
		client: k8sClientSet.CoreV1().RESTClient(),
		config: restCfg,
	}, nil
}

func newK8SRest(masterURL, config string) (*k8sRest, error) {
	restCfg, err := newRestConfig(masterURL, config)
	if err != nil {
		return nil, err
	}

	k8sClientSet, err := kubernetes.NewForConfig(restCfg)
	if err != nil {
		return nil, fmt.Errorf("could not create kubernetes client set for rest config: %w", err)
	}

	return &k8sRest{
		client: k8sClientSet.CoreV1().RESTClient(),
		config: restCfg,
	}, nil
}

func newRestConfig(masterURL, config string) (*rest.Config, error) {
	k8sCfg, err := clientcmd.BuildConfigFromKubeconfigGetter(
		masterURL,
		func() (*clientcmdapi.Config, error) {
			b, err := base64.StdEncoding.DecodeString(config)
			if err != nil {
				return nil, fmt.Errorf(
					"failed to base64 decode config for masterURL %s: %w",
					masterURL,
					err,
				)
			}

			return clientcmd.Load(b)
		},
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes config for masterURL %s: %w", masterURL, err)
	}

	return k8sCfg, nil
}
