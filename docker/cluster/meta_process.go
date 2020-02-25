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

type MetaProcess struct{
	Name                  string
	Binary                string
	role				  string
	listen				  string
	prof				  string
	localIP				  string
	logLevel			  string
	metadataDir			  string
	logDir				  string
	raftDir				  string
	raftHeartbeatPort	  string
	raftReplicaPort		  string
	consulAddr			  string
	exporterPort		  int
	masterAddr		   	  string
	totalMem			  string
	ExtraArgs []string

	proc *exec.Cmd
	exit chan error
}


func (meta *MetaProcess) Setup() (err error){

	meta.proc = exec.Command(
		meta.Binary,
		"-role",meta.role,
		"-listen",meta.listen,
		"-prof",meta.prof,
		"-ip",meta.localIP,
		"-meta_data_dir",meta.metadataDir,
		"-log_dir",meta.logDir,
		"-log_level",meta.logLevel,
		"-raft_dir",meta.raftDir,
		"-raft_heart_beat_port",meta.raftHeartbeatPort,
		"-raft_replica_port",meta.raftReplicaPort,
		"-consul_addr",meta.consulAddr,
		"-exporter_port", fmt.Sprintf("%d", meta.exporterPort),
		"-master_addr",meta.masterAddr,
		"-total_mem",meta.totalMem,
	)
	meta.proc.Args = append(meta.proc.Args,meta.ExtraArgs...)

	errFile,_ := os.Create(path.Join(meta.logDir,"meta-stderr.txt"))
	meta.proc.Stderr = errFile
	meta.proc.Env = append(meta.proc.Env,os.Environ()...)
	log.LogInfof("%v %v", strings.Join(meta.proc.Args, " "))

	err = meta.proc.Start()
	if err != nil {
		return
	}
	meta.exit = make(chan error)
	go func() {
		if meta.proc != nil {
			meta.exit <- meta.proc.Wait()
		}
	}()

	timeout := time.Now().Add(60 * time.Second)
	for time.Now().Before(timeout) {
		if meta.WaitForStatus() {
			return nil
		}
		select {
		case err := <-meta.exit:
			return fmt.Errorf("process '%s' exited prematurely (err: %s)", meta.Name, err)
		default:
			time.Sleep(300 * time.Millisecond)
		}
	}

	return fmt.Errorf("process '%s' timed out after 60s (err: %s)", meta.Name, <-meta.exit)

}


func (meta *MetaProcess) WaitForStatus() bool {
	resp, err := http.Get(meta.localIP)
	if err != nil {
		return false
	}
	if resp.StatusCode == 200 {
		return true
	}
	return false
}

func (meta *MetaProcess) TearDown() error {
	if meta.proc == nil || meta.exit == nil {
		return nil
	}

	// Attempt graceful shutdown with SIGTERM first
	meta.proc.Process.Signal(syscall.SIGTERM)

	select {
	case err := <-meta.exit:
		meta.proc = nil
		return err

	case <-time.After(10 * time.Second):
		meta.proc.Process.Kill()
		meta.proc = nil
		return <-meta.exit
	}
}


func MetaProcessInstance() *MetaProcess {
	meta := &MetaProcess{
		Name:       "meta",
		Binary:     "meta",
		role: 		"metanode",
		listen: 	"9021",
		prof: 		"9092",
		localIP:	"192.168.31.173",
		logLevel: 	"debug",
		metadataDir: "/export/Data/metanode",
		logDir: 	"/export/Logs/metanode",
		raftDir: 	"/export/Data/metanode/raft",
		raftHeartbeatPort: "9093",
		raftReplicaPort: "9094",
		consulAddr: 	"http://consul.prometheus-cfs.local",
		exporterPort: 	9511,
		totalMem:  	"17179869184",
		/*masterAddr: [
			"192.168.31.173:80",
		"192.168.31.141:80",
		"192.168.30.200:80"
	]*/
	}
	return meta
}
