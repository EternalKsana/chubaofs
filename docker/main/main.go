package main

import (
	."chubaofs/chubaofs/chubaofs/docker/cluster"
	//"fmt"
)

func main() {
	client := ClientProcessInstance()
	cluster.ClientProcess.Setup(*client)
	//fmt.Println("%s",err)
	//fmt.Println("%s",client.masterAddr)
	cluster.ClientProcess.TearDown(*client)

}
