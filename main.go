package main

import (
	"flag"
	"log"

	"github.com/zdnscloud/vanguard2-controller/controller"
)

func main() {
	var grpcServer, clustDomain, serviceIPRange, podIPRange, serverAddress string
	flag.StringVar(&grpcServer, "grpc-server", "", "vanguard2 grpc server address")
	flag.StringVar(&clustDomain, "cluster-domain", "", "k8s cluster domain")
	flag.StringVar(&serviceIPRange, "service-ip-range", "", "service ip range")
	flag.StringVar(&podIPRange, "pod-ip-range", "", "pod ip range")
	flag.StringVar(&serverAddress, "dns-server", "", "k8s dns service address")
	client, err := controller.NewVgClient(grpcServer, clustDomain, serviceIPRange, podIPRange, serverAddress)
	if err != nil {
		log.Fatalf("create vg client failed:%s", err.Error())
	}
	ctl, err := controller.NewK8sController(client)
	if err != nil {
		log.Fatalf("create k8s controller failed:%s", err.Error())
	}
	ctl.Run()
}
