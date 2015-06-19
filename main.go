package main

import (
	"log"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"syscall"
)

func init() {
	runtime.GOMAXPROCS(1)
	runtime.LockOSThread()
}

const Version = "0.1.0"

func main() {
	log.SetFlags(0)

	if len(os.Args) < 2 {
		log.Println("!! no command to run")
		log.Printf("usage: %s [command]", os.Args[0])
		log.Println()
		log.Printf("%s version: %s (%s on %s/%s; %s)", os.Args[0], Version, runtime.Version(), runtime.GOOS, runtime.GOARCH, runtime.Compiler)
		os.Exit(1)
	}

	// first, test to make sure the nginx config is valid, if not error early so we can see the error
	test := exec.Command("nginx", "-t")
	test.Stdout = os.Stdout
	test.Stderr = os.Stderr
	if err := test.Run(); err != nil {
		log.Fatal(err)
	}

	// next, spawn our child process and follow it's std{err,out}
	child := exec.Command(os.Args[1], os.Args[2:]...)
	child.Stdout = os.Stdout
	child.Stderr = os.Stderr
	if err := child.Start(); err != nil {
		log.Fatal(err)
	}

	// lastly spawn nginx, but we don't care about it's std{err,out}
	nginx := exec.Command("nginx", "-g", "daemon off;")
	if err := nginx.Start(); err != nil {
		log.Fatal(err)
	}

	// wait for either of the processes to exit
	done := make(chan struct{})
	go func() {
		child.Wait()
		done <- struct{}{}
	}()
	go func() {
		nginx.Wait()
		done <- struct{}{}
	}()

	// intercept our signals
	signals := make(chan os.Signal, 1)
	signal.Notify(signals)
	go func() {
		for sig := range signals {
			if sig == syscall.SIGTERM {
				// `docker stop` sends a SIGTERM, but we want to incercept and convert to SIGQUIT
				// because nginx will gracefully exit with SIGQUIT.
				nginx.Process.Signal(syscall.SIGQUIT)
			} else {
				child.Process.Signal(sig)
			}
		}
	}()
	<-done
	// shut down
	child.Process.Signal(syscall.SIGTERM)
	nginx.Process.Signal(syscall.SIGTERM)
	child.Wait()
	nginx.Wait()
}
