package main

import (
	"chubaofs/chubaofs/chubaofs/docker/cluster"


	"fmt"
)

func main() {
	client := cluster.ClientProcessInstance()
	err :=(*cluster.ClientProcess).Setup(client)
	fmt.Println("this is err:",err)
	//fmt.Println("%s",client.masterAddr)
	t :=(*cluster.ClientProcess).TearDown(client)
	fmt.Println("this is teardown bool:",t)

	/*client := cluster.MasterProcessInstance()
	err :=(*cluster.MasterProcess).Setup(client)
	fmt.Println("this is err:",err)
	//fmt.Println("%s",client.masterAddr)
	t :=(*cluster.MasterProcess).TearDown(client)
	fmt.Println("this is teardown bool:",t)


	client := cluster.ClientAuthProcessInstance()
	err :=(*cluster.ClientAuthProcess).Setup(client)
	fmt.Println("this is err:",err)
	//fmt.Println("%s",client.masterAddr)
	t :=(*cluster.ClientAuthProcess).TearDown(client)
	fmt.Println("this is teardown bool:",t)


	client := cluster.DataNodeProcessInstance()
	err :=(*cluster.DataNodeProcess).Setup(client)
	fmt.Println("this is err:",err)
	//fmt.Println("%s",client.masterAddr)
	t :=(*cluster.DataNodeProcess).TearDown(client)
	fmt.Println("this is teardown bool:",t)


	client := cluster.MasterProcessInstance()
	err :=(*cluster.MetaProcess).Setup(client)
	fmt.Println("this is err:",err)
	//fmt.Println("%s",client.masterAddr)
	t :=(*cluster.MetaProcess).TearDown(client)
	fmt.Println("this is teardown bool:",t)


	client := cluster.ObjectNodeProcessInstance()
	err :=(*cluster.ObjectNodeProcess).Setup(client)
	fmt.Println("this is err:",err)
	//fmt.Println("%s",client.masterAddr)
	t :=(*cluster.ObjectNodeProcess).TearDown(client)
	fmt.Println("this is teardown bool:",t)*/
}

