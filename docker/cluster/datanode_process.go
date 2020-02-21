package cluster

import (
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

type DataNodeProcess struct{
	Name                  string
	Binary                string
	role				  string
	listen				  string
	localIP				  string
	prof				  string
	logDir				  string
	logLevel			  string
	raftHeartbeat		  string
	raftReplica			  string
	raftDir				  string
	consulAddr			  string
	exporterPort		  int
	masterAddr	[]string
	disks	[]string
	ExtraArgs []string

	proc *exec.Cmd
	exit chan error
}



func (datanode *DataNodeProcess) Setup() (err error){

	datanode.proc = exec.Command(
		datanode.Binary,
		"-role",datanode.role,
		"-listen",datanode.listen,
		"-local_ip",datanode.localIP,
		"-prof",datanode.prof,
		"-log_dir",datanode.logDir,
		"-log_level",datanode.logLevel,
		"-raft_heart_beat",datanode.raftHeartbeat,
		"-raft_replica",datanode.raftReplica,
		"-raft_dir",datanode.raftDir,
		"-consul_addr",datanode.consulAddr,
		"-exporter_port", fmt.Sprintf("%d", datanode.exporterPort),

	)
	datanode.proc.Args = append(datanode.proc.Args,datanode.ExtraArgs...)

	errFile,_ := os.Create(path.Join(datanode.logDir,"datanode-stderr.txt"))
	datanode.proc.Stderr = errFile
	datanode.proc.Env = append(datanode.proc.Env,os.Environ()...)
	log.LogInfof("%v %v", strings.Join(datanode.proc.Args, " "))

	err = datanode.proc.Start()
	if err != nil {
		return
	}
	datanode.exit = make(chan error)
	go func() {
		if datanode.proc != nil {
			datanode.exit <- datanode.proc.Wait()
		}
	}()

	timeout := time.Now().Add(60 * time.Second)
	for time.Now().Before(timeout) {
		if datanode.WaitForStatus() {
			return nil
		}
		select {
		case err := <-datanode.exit:
			return fmt.Errorf("process '%s' exited prematurely (err: %s)", datanode.Name, err)
		default:
			time.Sleep(300 * time.Millisecond)
		}
	}

	return fmt.Errorf("process '%s' timed out after 60s (err: %s)", datanode.Name, <-datanode.exit)

}


func (datanode *DataNodeProcess) WaitForStatus() bool {
	resp, err := http.Get(datanode.localIP)
	if err != nil {
		return false
	}
	if resp.StatusCode == 200 {
		return true
	}
	return false
}

func (datanode *DataNodeProcess) TearDown() error {
	if datanode.proc == nil || datanode.exit == nil {
		return nil
	}

	// Attempt graceful shutdown with SIGTERM first
	datanode.proc.Process.Signal(syscall.SIGTERM)

	select {
	case err := <-datanode.exit:
		datanode.proc = nil
		return err

	case <-time.After(10 * time.Second):
		datanode.proc.Process.Kill()
		datanode.proc = nil
		return <-datanode.exit
	}
}


func DataNodeProcessInstance() *DataNodeProcess {
	datanode := &DataNodeProcess{
		Name:       "datanode",
		Binary:     "datanode",
		role: 		"datanode",
		listen: 	"6000",
		prof: 		"6001",
		localIP:	"192.168.31.174",
		logDir: 	"/export/Logs/datanode",
		logLevel: 	"info",
		raftHeartbeat: "9095",
		raftReplica: "9096",
		raftDir: 		"/export/Data/datanode/raft",
		consulAddr: 	"http://consul.prometheus-cfs.local",
		exporterPort:	 9512,
		masterAddr: 	[
			"192.168.31.173:80",
		"192.168.31.141:80",
		"192.168.30.200:80"
	],
		disks: [
		"/data0:21474836480",
		"/data1:21474836480"
	]
	}
	return datanode
}
