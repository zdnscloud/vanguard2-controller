package util

import (
	"fmt"
	"net"
	"strings"

	"github.com/zdnscloud/g53"
)

const (
	ReverseBaseZone = "in-addr.arpa"
)

func ReverseZoneName(network string) (*g53.Name, error) {
	_, net, err := net.ParseCIDR(network)
	if err != nil {
		return nil, err
	}

	ipLabels := strings.Split(network, ".")
	ones, _ := net.Mask.Size()
	var zone string
	switch ones {
	case 8:
		zone = strings.Join([]string{ipLabels[0], ReverseBaseZone}, ".")
	case 16:
		zone = strings.Join([]string{ipLabels[1], ipLabels[0], ReverseBaseZone}, ".")
	case 24:
		zone = strings.Join([]string{ipLabels[2], ipLabels[1], ipLabels[0], ReverseBaseZone}, ".")
	default:
		return nil, fmt.Errorf("only support 8, 16, 24 bits network mask")
	}

	return g53.NameFromStringUnsafe(zone), nil
}

func ReverseIPName(ip string) (*g53.Name, error) {
	labels := strings.Split(ip, ".")
	if len(labels) != 4 {
		return nil, fmt.Errorf("ipv4 address %s isn't valid", ip)
	} else {
		return g53.NameFromStringUnsafe(strings.Join([]string{labels[3], labels[2], labels[1], labels[0], ReverseBaseZone}, ".")), nil
	}
}
