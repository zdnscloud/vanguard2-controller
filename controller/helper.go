package controller

import (
	corev1 "k8s.io/api/core/v1"
)

func isSubsetEqual(sa, sb corev1.EndpointSubset) bool {
	if len(sa.Addresses) != len(sb.Addresses) {
		return false
	}
	if len(sa.Ports) != len(sb.Ports) {
		return false
	}

	for i, addr := range sa.Addresses {
		baddr := sb.Addresses[i]
		if addr.IP != baddr.IP {
			return false
		}
		if addr.Hostname != baddr.Hostname {
			return false
		}
	}

	for i, port := range sa.Ports {
		bport := sb.Ports[i]
		if port.Name != bport.Name {
			return false
		}
		if port.Port != bport.Port {
			return false
		}
		if port.Protocol != bport.Protocol {
			return false
		}
	}
	return true
}

func isSubsetsEqual(a, b *corev1.Endpoints) bool {
	if len(a.Subsets) != len(b.Subsets) {
		return false
	}

	//subsets should be sorted
	for i, sa := range a.Subsets {
		sb := b.Subsets[i]
		if isSubsetEqual(sa, sb) == false {
			return false
		}
	}
	return true
}

func isHeaderlessService(svc *corev1.Service) bool {
	return svc.Spec.Type != corev1.ServiceTypeExternalName &&
		svc.Spec.ClusterIP == corev1.ClusterIPNone
}

func isNormalService(svc *corev1.Service) bool {
	return svc.Spec.Type != corev1.ServiceTypeExternalName &&
		svc.Spec.ClusterIP != corev1.ClusterIPNone
}

func isExternalService(svc *corev1.Service) bool {
	return svc.Spec.Type == corev1.ServiceTypeExternalName &&
		svc.Spec.ExternalName != ""
}
