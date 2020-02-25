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
type ObjectNodeProcess struct{
	Name                  string
	Binary                string
	role				  string
	listen				  string
	region				  string
	domains				  []string
	logDir				  string
	logLevel			  string
	masterAddr			  []string
	exporterPort		          int
	prof				  string

	ExtraArgs []string

	proc *exec.Cmd
	exit chan error
}



func (objectnode *ObjectNodeProcess) Setup() (err error){

	objectnode.proc = exec.Command(
		objectnode.Binary,
		"-role",objectnode.role,
		"-listen",objectnode.listen,
		"-prof",objectnode.prof,
		"-region",objectnode.region,
		"-log_dir",objectnode.logDir,
		"-log_level",objectnode.logLevel,
		"-exporter_port", fmt.Sprintf("%d", objectnode.exporterPort), )
	objectnode.proc.Args = append(objectnode.proc.Args,objectnode.ExtraArgs...)

	errFile,_ := os.Create(path.Join(objectnode.logDir,"objectnode-stderr.txt"))
	objectnode.proc.Stderr = errFile
	objectnode.proc.Env = append(objectnode.proc.Env,os.Environ()...)
	log.LogInfof("%v %v", strings.Join(objectnode.proc.Args, " "))

	err = objectnode.proc.Start()
	if err != nil {
		return
	}
	objectnode.exit = make(chan error)
	go func() {
		if objectnode.proc != nil {
			objectnode.exit <- objectnode.proc.Wait()
		}
	}()

	timeout := time.Now().Add(60 * time.Second)
	for time.Now().Before(timeout) {
		if objectnode.WaitForStatus() {
			return nil
		}
		select {
		case err := <-objectnode.exit:
			return fmt.Errorf("process '%s' exited prematurely (err: %s)", objectnode.Name, err)
		default:
			time.Sleep(300 * time.Millisecond)
		}
	}

	return fmt.Errorf("process '%s' timed out after 60s (err: %s)", objectnode.Name, <-objectnode.exit)

}


func (objectnode *ObjectNodeProcess) WaitForStatus() bool {
	resp, err := http.Get(objectnode.masterAddr[1])
	if err != nil {
		return false
	}
	if resp.StatusCode == 200 {
		return true
	}
	return false
}

func (objectnode *ObjectNodeProcess) TearDown() error {
	if objectnode.proc == nil || objectnode.exit == nil {
		return nil
	}

	// Attempt graceful shutdown with SIGTERM first
	objectnode.proc.Process.Signal(syscall.SIGTERM)

	select {
	case err := <-objectnode.exit:
		objectnode.proc = nil
		return err

	case <-time.After(10 * time.Second):
		objectnode.proc.Process.Kill()
		objectnode.proc = nil
		return <-objectnode.exit
	}
}


func ObjectNodeProcessInstance() *ObjectNodeProcess {
	objectnode := &ObjectNodeProcess{
		Name:       "objectnode",
		Binary:     "objectnode",
		role: 		"objectnode",
		listen: 	":80",
		region: 	"cfs_default",
		/*domains: 	[
			"object.cfs.local"
	],*/
		logDir: 	"/opt/cfs/objectnode/logs",
		logLevel: 	"debug",
		/*masterAddr: 	[
		"172.20.240.95:7002",
		"172.20.240.94:7002",
		"172.20.240.67:7002"
	],*/
		exporterPort: 9512,
		prof: 		"7013",
	}
	return objectnode
}
