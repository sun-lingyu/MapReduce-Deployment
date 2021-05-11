package mr

//
// RPC definitions.
//
// remember to capitalize all names.
//

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"
	"path"
	"path/filepath"

	"net"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

// Add your RPC definitions here.
//!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
//very very important:
//Must capitalize the fileds in RPC related structs!
//(for example: Status in FileRequestReply)
//otherwise it will not be set to coordinator assigned value,
//and will be regarded as private.
//!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!

type RequestArgs struct {
	Workerid string //use ip address to represent worker
	Rtype    int    //0: ask for work; 1: ask for intermediate filename(send by reduce worker)
}

type WorkRequestReply struct {
	Wtype    int    //0: map work; 1: reduce work; 2: exit work; 3: wait for others to finish
	NReduce  int    //number of reduce tasks
	NMap     int    //number of map tasks
	Filename string //input file for map work. only valid when wtype==0
	Rnumber  int    //reduce partition number. only valid when wtype==1
}

type FileRequestReply struct {
	Status    int //0: normal; 1: exception
	Filenames []string
	Hostnames []string
	Ips       []string
}

//
// example to show how to declare the arguments
// and reply for an RPC.
//
type ExampleArgs struct {
	X int
}

type ExampleReply struct {
	Y int
}

// Cook up a unique-ish UNIX-domain socket name
// in /var/tmp, for the coordinator.
// Can't use the current directory since
// Athena AFS doesn't support UNIX-domain sockets.
func coordinatorSock() string {
	s := "/var/tmp/824-mr-"
	s += strconv.Itoa(os.Getuid())
	return s
}

func connect_session(user, password, host string, port int) (*ssh.Session, error) {
	var (
		auth         []ssh.AuthMethod
		addr         string
		clientConfig *ssh.ClientConfig
		sshClient    *ssh.Client
		session      *ssh.Session
		err          error
	)
	// get auth method
	auth = make([]ssh.AuthMethod, 0)
	auth = append(auth, ssh.Password(password))

	clientConfig = &ssh.ClientConfig{
		User:    user,
		Auth:    auth,
		Timeout: 30 * time.Second,
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	}

	// connet to ssh
	addr = fmt.Sprintf("%s:%d", host, port)

	if sshClient, err = ssh.Dial("tcp", addr, clientConfig); err != nil {
		return nil, err
	}

	// create ssh session
	if session, err = sshClient.NewSession(); err != nil {
		return nil, err
	}

	return session, nil
}

func connect(user, password, host string, port int) (*sftp.Client, error) {
	var (
		auth         []ssh.AuthMethod
		addr         string
		clientConfig *ssh.ClientConfig
		sshClient    *ssh.Client
		sftpClient   *sftp.Client
		err          error
	)
	// get auth method
	auth = make([]ssh.AuthMethod, 0)
	auth = append(auth, ssh.Password(password))

	clientConfig = &ssh.ClientConfig{
		User:    user,
		Auth:    auth,
		Timeout: 30 * time.Second,
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
            return nil
        },
	}

	// connet to ssh
	addr = fmt.Sprintf("%s:%d", host, port)

	if sshClient, err = ssh.Dial("tcp", addr, clientConfig); err != nil {
		return nil, err
	}

	// create sftp client
	if sftpClient, err = sftp.NewClient(sshClient); err != nil {
		return nil, err
	}

	return sftpClient, nil
}

func readRemote(user, password, host, filename string) (*sftp.Client, *sftp.File, error) {
	var (
		err        error
		sftpClient *sftp.Client
	)
	//Here, change to the actual SSH connection user name, password, host name or IP, SSH port
	sftpClient, err = connect(user, password, host, 22)
	if err != nil {
		log.Fatal(err)
	}
	//defer sftpClient.Close()

	filename=filepath.Join("/root/mapreduce/src/main",filename)
	fmt.Println(filename)

	srcFile, err := sftpClient.Open(filename)
	return sftpClient, srcFile, err
}

func sendRemote(user, password, host, oname string) (error){
	sftpClient, err := connect(user, password, host, 22)
	if err != nil {
		log.Fatal(err)
	}
	
	var remoteDir = "/root/mapreduce/src/main"
	var remoteFileName = oname

	dstFile, err := sftpClient.Create(path.Join(remoteDir, remoteFileName))
	if err != nil {
	  log.Fatal(err)
	}
	defer dstFile.Close()

	file, err := os.Open(oname) // For read access.
	if err != nil {
		log.Fatal(err)
	}
  
	buf := make([]byte, 1024)
	for {
		n, err := file.Read(buf)
		if n == 0 {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
	  dstFile.Write(buf[:n])
	}
  
	fmt.Println("copy file to remote server finished!")
	return err
}

// session.Run 会阻塞，所以需要多线程
func AwakenWorker(idx int, session *ssh.Session, command string) {
	fmt.Println("Send command to worker", idx)
	session.Stdout = os.Stdout
	session.Stderr = os.Stderr
	// session.Run("cd /root/mapreduce/src/main;" + "rm mr-* ") 不能有
	err := session.Run("cd /root/mapreduce/src/main;" + command)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Worker", idx, " job finished.")
	session.Close()
}

// This function raise all workers(hosts) up, using ssh session.
func AwakenWorkers(user, password string, hosts []string, command string) {
	var (
		session *ssh.Session
		err     error
	)
	for i, host := range hosts {
		fmt.Println(i, host)
		session, err = connect_session(user, password, host, 22)
		if err != nil {
			log.Fatal(err)
		}
		go AwakenWorker(i, session, command)

	}

}

func GetOutboundIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80") //actually can be any address...
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP.String()
}
