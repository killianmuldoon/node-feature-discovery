package kubeconf

import (
	"context"
	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
	"io/ioutil"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	kubeletconfigv1beta1 "k8s.io/kubelet/config/v1beta1"
)

// configzEndpoint is a struct to allow unmarshalling the TopologyManagerPolicy from the kubelet configz endpoint
// The endpoint does not return kubeletConfig API object. It contains all the configuration fields without the headers.
// This is a limitation of the current (1.20) kubelet endpoint. This struct is a minimum wrapper for Topology Policy.
type configzEndpoint struct {
	Kubeletconfig struct {
		TopologyManagerPolicy          string   `json:"topologyManagerPolicy"`
	} `json:"kubeletconfig"`
}

// GetKubeletConfigFromLocalFile returns KubeletConfiguration loaded from the node local config
func GetKubeletConfigFromLocalFile(kubeletConfigPath string) (*kubeletconfigv1beta1.KubeletConfiguration, error) {
	kubeletBytes, err := ioutil.ReadFile(kubeletConfigPath)
	if err != nil {
		return nil, err
	}

	kubeletConfig := &kubeletconfigv1beta1.KubeletConfiguration{}
	if err := yaml.Unmarshal(kubeletBytes, kubeletConfig); err != nil {
		return nil, err
	}
	return kubeletConfig, nil
}
// GetKubeletConfigFromKubeletAPI creates a kubernetes client to pull Topology Manager information on configz
func GetKubeletConfigFromKubeletAPI(kubeletConfigPath string, nodename string) (*kubeletconfigv1beta1.KubeletConfiguration, error ){
	config, err := rest.InClusterConfig()
	if err != nil {
		config, err = clientcmd.BuildConfigFromFlags("", kubeletConfigPath)
	}
	if err != nil {
		return nil, err
	}
	cl, err  := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	//request here is the same as: curl https://$APISERVER:6443/api/v1/nodes/$NODE_NAME/proxy/configz
	request := cl.CoreV1().RESTClient().Get().Resource("nodes").Name(nodename).SubResource("proxy").Suffix("configz")
	kubeletBytes, err := request.DoRaw(context.TODO())
	if err != nil {
		return nil, errors.Wrap(err, "failed to get info from node")
	}
	configz := &configzEndpoint{}
	if err := yaml.Unmarshal(kubeletBytes, configz); err != nil {
		return nil, err
	}
	return &kubeletconfigv1beta1.KubeletConfiguration{TopologyManagerPolicy: configz.Kubeletconfig.TopologyManagerPolicy}, nil
}
