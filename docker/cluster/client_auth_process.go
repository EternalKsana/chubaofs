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

type ClientAuthProcess struct{
	Name                  string
	Binary                string
	role				  string
	ip					  string
	port				  string
	prof			  	  string
	id					  string
	peers				  string
	logDir				  string
	logLevel			  string
	retainLogs			  string
	walDir				  string
	storeDir			  string
	clusterName			  string
	exporterPort		  int
	authServiceKey		  string
	authRootKey			  string
	enableHTTPS			  bool
	ExtraArgs []string

	proc *exec.Cmd
	exit chan error
}



func (clientauth *ClientAuthProcess) Setup() (err error){

	clientauth.proc = exec.Command(
		clientauth.Binary,
		"-role",clientauth.role,
		"-ip",clientauth.ip,
		"-port",clientauth.port,
		"-prof",clientauth.prof,
		"-id",clientauth.id,
		"-peers",clientauth.peers,
		"-log_dir",clientauth.logDir,
		"-log_level",clientauth.logLevel,
		"-retain_logs",clientauth.retainLogs,
		"-wal_dir",clientauth.walDir,
		"-store_dir",clientauth.storeDir,
		"-cluster_name",clientauth.clusterName,
		"-exporter_port", fmt.Sprintf("%d", clientauth.exporterPort),
		"-auth_service_key",clientauth.authServiceKey,
		"-auth_root_key",clientauth.authRootKey,
		"-enable_Https",fmt.Sprintf("%s	", clientauth.enableHTTPS),

	)
	clientauth.proc.Args = append(clientauth.proc.Args,clientauth.ExtraArgs...)

	errFile,_ := os.Create(path.Join(clientauth.logDir,"clientauth-stderr.txt"))
	clientauth.proc.Stderr = errFile
	clientauth.proc.Env = append(clientauth.proc.Env,os.Environ()...)
	log.LogInfof("%v %v", strings.Join(clientauth.proc.Args, " "))

	err = clientauth.proc.Start()
	if err != nil {
		return
	}
	clientauth.exit = make(chan error)
	go func() {
		if clientauth.proc != nil {
			clientauth.exit <- clientauth.proc.Wait()
		}
	}()

	timeout := time.Now().Add(60 * time.Second)
	for time.Now().Before(timeout) {
		if clientauth.WaitForStatus() {
			return nil
		}
		select {
		case err := <-clientauth.exit:
			return fmt.Errorf("process '%s' exited prematurely (err: %s)", clientauth.Name, err)
		default:
			time.Sleep(300 * time.Millisecond)
		}
	}

	return fmt.Errorf("process '%s' timed out after 60s (err: %s)", clientauth.Name, <-clientauth.exit)

}


func (clientauth *ClientAuthProcess) WaitForStatus() bool {
	resp, err := http.Get(clientauth.ip)
	if err != nil {
		return false
	}
	if resp.StatusCode == 200 {
		return true
	}
	return false
}

func (clientauth *ClientAuthProcess) TearDown() error {
	if clientauth.proc == nil || clientauth.exit == nil {
		return nil
	}

	// Attempt graceful shutdown with SIGTERM first
	clientauth.proc.Process.Signal(syscall.SIGTERM)

	select {
	case err := <-clientauth.exit:
		clientauth.proc = nil
		return err

	case <-time.After(10 * time.Second):
		clientauth.proc.Process.Kill()
		clientauth.proc = nil
		return <-clientauth.exit
	}
}


func ClientAuthProcessInstance() *ClientAuthProcess {
	clientauth := &ClientAuthProcess{
		Name:       "clientauth",
		Binary:     "clientauth",
		role: 		"authnode",
		ip: 		"192.168.0.14",
		port: 		"8080",
		prof:		"10088",
		id:			"1",
		peers: 		"1:192.168.0.14:8080,2:192.168.0.15:8081,3:192.168.0.16:8082",
		logDir: 	"/export/Logs/authnode",
		logLevel:	"info",
		retainLogs:	"100",
		walDir:		"/export/Data/authnode/raft",
		storeDir:	"/export/Data/authnode/rocksdbstore",
		exporterPort: 9510,
		clusterName:"test",
		authServiceKey:"9h/sNq4+5CUAyCnAZM927Y/gubgmSixh5hpsYQzZG20=",
		authRootKey:"wbpvIcHT/bLxLNZhfo5IhuNtdnw1n8kom+TimS2jpzs=",
		enableHTTPS:false,
	}
	return clientauth
}
