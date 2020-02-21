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
	exporterPort		  string
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
		"-mount_point",client.mountPoint,
		"-vol_name",client.volName,
		"-owner",client.owner,
		"-master_addr",client.masterAddr,
		"-log_dir",client.logDir,
		"-log_level",client.logLevel,
		"-prof_port",client.profPort,
		"-exporter_port",client.exporterPort,
		"-consul_addr",client.consulAddr,
		"-lookup_valid",client.lookupValid,
		"-attr_valid",client.attrValid,
		"-i_cache_timeout",client.icacheTimeout,
		"-ensync_write",client.enSyncWrite,
		"-auto_inval_data",client.autoInvalData,
		)
	client.proc.Args = append(client.proc.Args,client.ExtraArgs...)

	errFile,_ := os.Create(path.Join(client.logDir,"client-stderr.txt"))
	client.proc.Stderr = errFile
	client.proc.Env = append(client.proc.Env,os.Environ()...)
	log.LogInfof("%v %v", strings.Join(client.proc.Args, " "))

	err = client.proc.Start()
	if err != nil {
		return
	}
	client.exit = make(chan error)
	go func() {
		if client.proc != nil {
			client.exit <- client.proc.Wait()
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

	return fmt.Errorf("process '%s' timed out after 60s (err: %s)", client.Name, <-client.exit)

}


func (client *ClientProcess) WaitForStatus() bool {
	resp, err := http.Get(client.masterAddr)
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
		return err

	case <-time.After(10 * time.Second):
		client.proc.Process.Kill()
		client.proc = nil
		return <-client.exit
	}
}


func ClientProcessInstance() *ClientProcess {
	client := &ClientProcess{
		Name:                        "client",
		Binary:                      "client",
		mountPoint:                  "/mnt/fuse",
		volName: 					 "test",
		owner: 						 "cfs",
		masterAddr: 				 "192.168.31.173:80,192.168.31.141:80,192.168.30.200:80",
		logDir: 					 "/export/Logs/client",
		logLevel: 					 "info",
		profPort:					 "10094",
		consulAddr: 				 "http://consul.prometheus-cfs.local",

	}


	return client
}