package main

import (
	"log"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

//func getNetInfo() (string, net.HardwareAddr) {
//
//	addrs, err := net.InterfaceAddrs()
//	if err != nil {
//		fmt.Println(err)
//	}
//
//	var hostIP, currentNetworkHardwareName string
//
//	for _, address := range addrs {
//
//		// check the address type and if it is not a loopback the display it
//		// = GET LOCAL IP ADDRESS
//		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
//			if ipnet.IP.To4() != nil {
//				hostIP = ipnet.IP.String()
//			}
//		}
//	}
//
//	interfaces, _ := net.Interfaces()
//	for _, interf := range interfaces {
//
//		if addr, err := interf.Addrs(); err == nil {
//			for _, addr := range addr {
//				// only interested in the name with current IP address
//				if strings.Contains(addr.String(), hostIP) {
//					currentNetworkHardwareName = interf.Name
//				}
//			}
//		}
//	}
//
//	// extract the hardware information base on the interface name
//	// capture above
//	netInterface, err := net.InterfaceByName(currentNetworkHardwareName)
//
//	if err != nil {
//		fmt.Println(err)
//	}
//
//	macAddress := netInterface.HardwareAddr
//
//	// verify if the MAC address can be parsed properly
//	_, err = net.ParseMAC(macAddress.String())
//
//	if err != nil {
//		log.Println("No able to parse MAC address : ", err)
//	}
//
//	return hostIP, macAddress
//}
