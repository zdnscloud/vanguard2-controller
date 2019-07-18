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

func NewVgClient(grpcServer, clustDomain, serviceIPRange, podIPRange, serverAddress string) (*VgClient, error) {
	dialOptions := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithTimeout(GRPCConnTimeout),
	}

	conn, err := grpc.Dial(grpcServer, dialOptions...)
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

	if err := cli.initZones(); err != nil {
		return nil, err
	}

	return cli, nil
}

func (c *VgClient) Close() error {
	return c.conn.Close()
}

func (c *VgClient) initZones() error {
	c.doDeleteZone([]*g53.Name{c.serviceZone, c.serviceReverseZone, c.podReverseZone})
	return c.createZones()
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

func (c *VgClient) doDeleteZone(zones []*g53.Name) {
	zoneNames := make([]string, len(zones))
	for i, z := range zones {
		zoneNames[i] = z.String(false)
	}

	c.grpcClient.DeleteZone(context.TODO(), &pb.DeleteZoneRequest{
		Zones: zoneNames,
	})
}

func (c *VgClient) replaceServiceRRset(rrset *g53.RRset) error {
	return c.replaceRRset(c.serviceZone, rrset)
}

func (c *VgClient) deleteServiceRRset(name *g53.Name, typ g53.RRType) error {
	return c.deleteRRset(c.serviceZone, name, typ)
}

func (c *VgClient) replacePodReverseRRset(rrset *g53.RRset) error {
	return c.replaceRRset(c.podReverseZone, rrset)
}

func (c *VgClient) deletePodReverseRRset(name *g53.Name) error {
	return c.deleteRRset(c.podReverseZone, name, g53.RR_PTR)
}

func (c *VgClient) replaceServiceReverseRRset(rrset *g53.RRset) error {
	return c.replaceRRset(c.serviceReverseZone, rrset)
}

func (c *VgClient) deleteServiceReverseRRset(name *g53.Name) error {
	return c.deleteRRset(c.serviceReverseZone, name, g53.RR_PTR)
}

func (c *VgClient) deleteRRset(zone *g53.Name, name *g53.Name, typ g53.RRType) error {
	_, err := c.grpcClient.DeleteRRset(context.TODO(), &pb.DeleteRRsetRequest{
		Zone: zone.String(false),
		Rrsets: []*pb.RRsetHeader{
			&pb.RRsetHeader{
				Name: name.String(false),
				Type: g53RRTypeToPB(typ),
			},
		},
	})
	return err
}

func (c *VgClient) replaceRRset(zone *g53.Name, rrset *g53.RRset) error {
	if err := c.deleteRRset(zone, rrset.Name, rrset.Type); err != nil {
		return err
	}

	var rdatas []string
	for _, rdata := range rrset.Rdatas {
		rdatas = append(rdatas, rdata.String())
	}

	_, err := c.grpcClient.AddRRset(context.TODO(), &pb.AddRRsetRequest{
		Zone: zone.String(false),
		Rrsets: []*pb.RRset{
			&pb.RRset{
				Name:   rrset.Name.String(false),
				Type:   g53RRTypeToPB(rrset.Type),
				Ttl:    uint32(rrset.Ttl),
				Rdatas: rdatas,
			},
		},
	})
	return err
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

func g53RRTypeToPB(typ g53.RRType) pb.RRType {
	switch typ {
	case g53.RR_A:
		return pb.RRType_A
	case g53.RR_AAAA:
		return pb.RRType_AAAA
	case g53.RR_NS:
		return pb.RRType_NS
	case g53.RR_SOA:
		return pb.RRType_SOA
	case g53.RR_CNAME:
		return pb.RRType_CNAME
	case g53.RR_TXT:
		return pb.RRType_TXT
	case g53.RR_SRV:
		return pb.RRType_SRV
	case g53.RR_PTR:
		return pb.RRType_PTR
	default:
		panic("unsupported type:" + typ.String())
	}
}
