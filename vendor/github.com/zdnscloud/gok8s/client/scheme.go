package client

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
)

func GetDefaultScheme() *runtime.Scheme {
	return scheme.Scheme
}
