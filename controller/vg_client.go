package controller

import (
	"context"
	"strings"
	"time"

	"github.com/zdnscloud/g53"
	"github.com/zdnscloud/vanguard2-controller/util"

	pb "github.com/zdnscloud/vanguard2-controller/proto"
	"google.golang.org/grpc"
	corev1 "k8s.io/api/core/v1"
)

const (
	GRPCConnTimeout    = 10 * time.Second
	DefaultTTL         = g53.RRTTL(5)
	DefaultSRVWeight   = 100
	DefaultSRVPriority = 10
	DNSSchemaVersion   = "1.0.1"
)

type VgClient struct {
	grpcClient pb.DynamicUpdateInterfaceClient
	conn       *grpc.ClientConn

	serviceZone        *g53.Name
	serviceReverseZone *g53.Name
	podReverseZone     *g53.Name
	serverAddress      string
}

func New(addr, clustDomain, serviceIPRange, podIPRange, serverAddress string) (*VgClient, error) {
	dialOptions := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithTimeout(GRPCConnTimeout),
	}

	conn, err := grpc.Dial(addr, dialOptions...)
	if err != nil {
		return nil, err
	}

	serviceZone, err := g53.NameFromString(clustDomain)
	if err != nil {
		return nil, err
	}
	serviceReverseZone, err := util.ReverseZoneName(serviceIPRange)
	if err != nil {
		return nil, err
	}
	podReverseZone, err := util.ReverseZoneName(podIPRange)
	if err != nil {
		return nil, err
	}

	cli := &VgClient{
		grpcClient:         pb.NewDynamicUpdateInterfaceClient(conn),
		conn:               conn,
		serviceZone:        serviceZone,
		serviceReverseZone: serviceReverseZone,
		podReverseZone:     podReverseZone,
		serverAddress:      serverAddress,
	}

	if err := cli.createZones(); err != nil {
		return nil, err
	}

	return cli, nil
}

func (c *VgClient) Close() error {
	return c.conn.Close()
}

func (c *VgClient) createZones() error {
	if err := c.createServiceZone(); err != nil {
		return err
	}
	if err := c.createServiceReverseZone(); err != nil {
		return err
	}
	return c.createPodReverseZone()
}

func (c *VgClient) createServiceZone() error {
	return c.doCreateZone(c.serviceZone, ServiceZoneTemplate, map[string]interface{}{
		"origin":            c.serviceZone.String(false),
		"ttl":               DefaultTTL,
		"clusterDnsService": c.serverAddress,
		"dnsSchemaVersion":  DNSSchemaVersion,
	})
}

func (c *VgClient) createServiceReverseZone() error {
	return c.doCreateZone(c.serviceReverseZone, ServiceReverseZoneTemplate, map[string]interface{}{
		"origin":            c.serviceReverseZone.String(false),
		"ttl":               DefaultTTL,
		"clusterDnsService": c.serverAddress,
	})
}

func (c *VgClient) createPodReverseZone() error {
	return c.doCreateZone(c.podReverseZone, PodReverseZoneTemplate, map[string]interface{}{
		"origin":            c.podReverseZone.String(false),
		"ttl":               DefaultTTL,
		"clusterDnsService": c.serverAddress,
	})
}

func (c *VgClient) doCreateZone(zoneName *g53.Name, template string, templateParameter map[string]interface{}) error {
	zoneContent, err := util.CompileTemplateFromMap(template, templateParameter)
	if err != nil {
		return err
	}
	_, err = c.grpcClient.AddZone(context.TODO(), &pb.AddZoneRequest{
		Zone:        zoneName.String(false),
		ZoneContent: zoneContent,
	})
	return err
}

func (c *VgClient) replaceServiceRRset(name *g53.Name, typ g53.RRType, rrset *g53.RRset) error {
	return c.replaceRRset(c.serviceZone, name, typ, rrset)
}

func (c *VgClient) replacePodReverseRRset(name *g53.Name, typ g53.RRType, rrset *g53.RRset) error {
	return c.replaceRRset(c.podReverseZone, name, typ, rrset)
}

func (c *VgClient) replaceServiceReverseRRset(name *g53.Name, typ g53.RRType, rrset *g53.RRset) error {
	return c.replaceRRset(c.serviceReverseZone, name, typ, rrset)
}

func (c *VgClient) replaceRRset(zone *g53.Name, name *g53.Name, typ g53.RRType, rrset *g53.RRset) error {
	if rrset != nil {
		/*
			up.DeleteRRset(tx, rrset)
			up.Add(tx, rrset)
		*/
	} else {
		/*
			up.DeleteRRset(tx, &g53.RRset{
				Name: name,
				Type: typ,
			})
		*/
	}
	return nil
}

func (c *VgClient) getServiceName(svc *corev1.Service) *g53.Name {
	n, _ := g53.NameFromStringUnsafe(strings.Join([]string{svc.Name, svc.Namespace, "svc"}, ".")).Concat(c.serviceZone)
	return n
}

func (c *VgClient) getEndpointsAddrName(addr *corev1.EndpointAddress, svc, namespace string) *g53.Name {
	podName := addr.Hostname
	if podName == "" {
		podName = strings.Replace(addr.IP, ".", "-", 3)
	}
	n, _ := g53.NameFromStringUnsafe(strings.Join([]string{podName, svc, namespace, "svc"}, ".")).Concat(c.serviceZone)
	return n
}

func (c *VgClient) getPortName(port, protocol, svc, namespace string) *g53.Name {
	n, _ := g53.NameFromStringUnsafe(strings.Join([]string{"_" + port, "_" + protocol, svc, namespace, "svc"}, ".")).Concat(c.serviceZone)
	return n
}
