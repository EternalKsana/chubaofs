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

type MasterProcess struct{
	Name                  string
	Binary                string
	role				  string
	ip					  string
	listen				  string
	prof				  string
	id					  string
	peers				  string
	logDir			 	  string
	logLevel			  string
	retainLogs			  string
	walDir				  string
	storeDir			  string
	clusterName			  string
	exporterPort		  int
	consulAddr			  string
	metaNodeReservedMem	  string

	ExtraArgs []string

	proc *exec.Cmd
	exit chan error
}

func (master *MasterProcess) Setup() (err error){

	master.proc = exec.Command(
		master.Binary,
		"-role",master.role,
		"-ip",master.ip,
		"-listen",master.listen,
		"-prof",master.prof,
		"-id",master.id,
		"-peers",master.peers,
		"-log_dir",master.logDir,
		"-log_level",master.logLevel,
		"-retain_logs",master.retainLogs,
		"-wal_dir",master.walDir,
		"-store_dir",master.storeDir,
		"-cluster_name",master.clusterName,
		"-exporter_port", fmt.Sprintf("%d", master.exporterPort),
		"-consul_addr",master.consulAddr,
	)
	master.proc.Args = append(master.proc.Args,master.ExtraArgs...)

	errFile,_ := os.Create(path.Join(master.logDir,"master-stderr.txt"))
	master.proc.Stderr = errFile
	master.proc.Env = append(master.proc.Env,os.Environ()...)
	log.LogInfof("%v %v", strings.Join(master.proc.Args, " "))

	err = master.proc.Start()
	if err != nil {
		return
	}
	master.exit = make(chan error)
	go func() {
		if master.proc != nil {
			master.exit <- master.proc.Wait()
		}
	}()

	timeout := time.Now().Add(60 * time.Second)
	for time.Now().Before(timeout) {
		if master.WaitForStatus() {
			return nil
		}
		select {
		case err := <-master.exit:
			return fmt.Errorf("process '%s' exited prematurely (err: %s)", master.Name, err)
		default:
			time.Sleep(300 * time.Millisecond)
		}
	}

	return fmt.Errorf("process '%s' timed out after 60s (err: %s)", master.Name, <-master.exit)

}


func (master *MasterProcess) WaitForStatus() bool {
	resp, err := http.Get(master.ip)
	if err != nil {
		return false
	}
	if resp.StatusCode == 200 {
		return true
	}
	return false
}

func (master *MasterProcess) TearDown() error {
	if master.proc == nil || master.exit == nil {
		return nil
	}

	// Attempt graceful shutdown with SIGTERM first
	master.proc.Process.Signal(syscall.SIGTERM)

	select {
	case err := <-master.exit:
		master.proc = nil
		return err

	case <-time.After(10 * time.Second):
		master.proc.Process.Kill()
		master.proc = nil
		return <-master.exit
	}
}


func MasterProcessInstance() *MasterProcess {
	master := &MasterProcess{
		Name:       "master",
		Binary:     "master",
		role: 		"master",
		id:			"1",
		ip: 		"192.168.31.173",
		listen: 	"80",
		prof:		"10088",
		peers: 		"1:192.168.31.173:80,2:192.168.31.141:80,3:192.168.30.200:80",
		retainLogs:	"20000",
		logDir: 	"/export/Logs/master",
		logLevel:	"info",
		walDir:		"/export/Data/master/raft",
		storeDir:	"/export/Data/master/rocksdbstore",
		exporterPort: 9510,
		consulAddr: "http://consul.prometheus-cfs.local",
		clusterName:"test",
		metaNodeReservedMem: "134217728",
	}
	return master
}
