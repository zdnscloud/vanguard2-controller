package main

import (
	"flag"
	"log"
	"time"

	"github.com/zdnscloud/vanguard2-controller/controller"
)

func main() {
	var grpcServer, clusterDomain, serviceIPRange, podIPRange, serverAddress string
	flag.StringVar(&grpcServer, "grpc-server", "127.0.0.1:5555", "vanguard2 grpc server address")
	flag.StringVar(&clusterDomain, "cluster-domain", "", "k8s cluster domain")
	flag.StringVar(&serviceIPRange, "service-ip-range", "", "service ip range")
	flag.StringVar(&podIPRange, "pod-ip-range", "", "pod ip range")
	flag.StringVar(&serverAddress, "dns-server", "", "k8s dns service address")
	flag.Parse()

	log.Printf("start with: clusterDomain:%s, serviceIPRange:%s, podIPRange:%s, serverAddress:%s", clusterDomain, serviceIPRange, podIPRange, serverAddress)

	var client *controller.VgClient
	for {
		var err error
		client, err = controller.NewVgClient(grpcServer, clusterDomain, serviceIPRange, podIPRange, serverAddress)
		if err != nil {
			log.Printf("create vangaurd2 client failed:%s", err.Error())
		} else {
			break
		}
		<-time.After(time.Second)
	}
	log.Printf("finish initialize zone\n")

	ctl, err := controller.NewK8sController(client)
	if err != nil {
		log.Printf("create k8s controller failed:%s", err.Error())
		return
	}
	ctl.Run()
}
