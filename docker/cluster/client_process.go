package cluster

import(
	"chubaofs/chubaofs/chubaofs/util/log"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strings"
	"syscall"
	"time"
	"flag"
)

type ClientProcess struct{
	Name                  string
	Binary                string
	mountPoint			  string
	volName				  string
	owner				  string
	masterAddr			  string
	logDir				  string
	logLevel			  string
	profPort			  string
	exporterPort		  	  int
	consulAddr			  string
	lookupValid			  string
	attrValid			  string
	icacheTimeout		  string
	enSyncWrite			  string
	autoInvalData		  string

	ExtraArgs []string

	proc *exec.Cmd
	exit chan error
}

func (client *ClientProcess) Setup() (err error){

	client.proc = exec.Command(
		client.Binary,
		"-f",
		"-mount_point",client.mountPoint,
		"-vol_name",client.volName,
		"-owner",client.owner,
		"-master_addr",client.masterAddr,
		"-log_dir",client.logDir,
		"-log_level",client.logLevel,
		"-prof_port",client.profPort,
		"-exporter_port", fmt.Sprintf("%d", client.exporterPort),

		"-consul_addr",client.consulAddr,
		"-lookup_valid",client.lookupValid,
		"-attr_valid",client.attrValid,
		"-i_cache_timeout",client.icacheTimeout,
		"-ensync_write",client.enSyncWrite,
		"-auto_inval_data",client.autoInvalData,
		)
	client.mountPoint  =*flag.String(`-mount_point`,"/chubaofs/mnt","")
	client.volName     =*flag.String("-vol_name","ltptest","")
	client.owner	   =*flag.String("-owner","ltptest","")
	client.masterAddr  =*flag.String("-master_addr","192.168.0.11:17010,192.168.0.12:17010,192.168.0.13:17010","")
	client.logDir	   =*flag.String("-log_dir","chubaofs/log","")
	client.logLevel	   =*flag.String("-log_level","inifo","")
	client.profPort	   =*flag.String("-prof_port","17410","")
	client.consulAddr  =*flag.String("-consul_addr","http://192.168.0.100:8500","")
	client.exporterPort=*flag.Int("-exporter_port",9500,"")


	fmt.Println("masteraddr:  ",client.masterAddr)
	fmt.Println("logLevel:  ",client.logLevel)
	fmt.Println("logDir:  ",client.logDir)
	fmt.Println("consulAddr:  ",client.consulAddr)
	fmt.Println("exporterPort:  ",client.exporterPort)
	fmt.Println("mount_point:  ",client.mountPoint)
	fmt.Println("vol_name:  ",client.volName)
	fmt.Println("prof_port: ",client.profPort)
	fmt.Println("owner:  ",client.owner)


 
	client.proc.Args = append(client.proc.Args,client.ExtraArgs...)

	errFile,_ := os.Create(path.Join(client.logDir,"client-stderr.txt"))
	client.proc.Stderr = errFile
	client.proc.Env = append(client.proc.Env,os.Environ()...)
	log.LogInfof("%v %v", strings.Join(client.proc.Args, " "))
	/*cmd1 := exec.Command("client -f")
		if err = cmd1.Run(); err != nil {
		fmt.Println(err)
	}*/
	err = client.proc.Start()
	if err != nil {
		return
	}
	client.exit = make(chan error)
	go func() {
		if client.proc != nil {
			client.exit <- client.proc.Wait()
			fmt.Println("client.exit: ",client.exit)


		}
	}()

	timeout := time.Now().Add(60 * time.Second)
	for time.Now().Before(timeout) {
		if client.WaitForStatus() {
			return nil
		}
		select {
		case err := <-client.exit:
			return fmt.Errorf("process '%s' exited prematurely (err: %s)", client.Name, err)
		default:
			time.Sleep(300 * time.Millisecond)
		}
	}
	
	/*fmt.Println("masteraddr:  %s",client.masterAddr)
	fmt.Println("logLevel:  %s",client.logLevel)
	fmt.Println("logDir:  %s",client.logDir)
	fmt.Println("consulAddr:  %s",client.consulAddr)
	fmt.Println("exporterPort:  %d",client.exporterPort)
	fmt.Println("Binary:  %s",client.Binary)
	fmt.Println("autoInvalData:  %s",client.autoInvalData)
	fmt.Println("enSyncWrite:  %s",client.enSyncWrite)
	fmt.Println("icacheTimeout:  %s",client.icacheTimeout)
	fmt.Println("attrValid:  %s",client.attrValid)
	fmt.Println("lookupvid:  %s",client.lookupValid)
	fmt.Println("mout_point:  %s",client.mountPoint)
	fmt.Println("Name:  %s",client.Name)*/

	return fmt.Errorf("process '%s' timed out after 60s (err: %s)", client.Name, <-client.exit)

}


func (client *ClientProcess) WaitForStatus() bool {
	resp, err := http.Get("http://192.168.0.11:17010/admin/getCluster")
	if err != nil {
		return false
	}
	if resp.StatusCode == 200 {
		return true
	}
	return false
}

func (client *ClientProcess) TearDown() error {
	if client.proc == nil || client.exit == nil {
		return nil
	}

	// Attempt graceful shutdown with SIGTERM first
	client.proc.Process.Signal(syscall.SIGTERM)

	select {
	case err := <-client.exit:
		client.proc = nil
		fmt.Println("teardown err")
		return err

	case <-time.After(10 * time.Second):
		client.proc.Process.Kill()
		client.proc = nil
		fmt.Println("teardown exit")
		return <-client.exit
	}
}


func ClientProcessInstance() *ClientProcess {
	client := &ClientProcess{
		Name:                        "client",
		Binary:                      "client",
		mountPoint:                  "/chubaofs/mnt",
		volName: 					 "ltptest",
		owner: 						 "ltptest",
		masterAddr: 				 "192.168.0.11:17010,192.168.0.12:17010,192.168.0.13:17010",
		logDir: 					 "/chubaofs/log",
		logLevel: 					 "info",
		profPort:					 "17410",
		consulAddr: 				 "http://192.168.0.100:8500",
		exporterPort:			9500,

	}


	return client
}
