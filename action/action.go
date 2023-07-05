package action

import (
	"context"
	"dmfcta/apputils"
	"github.com/fsnotify/fsnotify"
	"github.com/go-cmd/cmd"
	"log"
	"strings"
	"time"
)

type JobInput struct {
	CMD       string
	SHELLTYPE string
	Timeout   time.Duration
	Endpoints []string
}

type action struct {
	watcher *fsnotify.Watcher
	jobs    []*job
}

type job struct {
	shellType *string
	command   *string
	timeout   *time.Duration
	endPoints []*string
	doSignal  chan struct{}
}

const (
	ShShell   = "sh"
	BashShell = "bash"
	ZshShell  = "zsh"
)

func New(watcher *fsnotify.Watcher) *action {
	return &action{
		watcher: watcher,
	}
}

func (a *action) Add(j JobInput) {
	a.jobs = append(a.jobs, &job{
		shellType: apputils.StrPointer(j.SHELLTYPE),
		command:   apputils.StrPointer(j.CMD),
		timeout:   &j.Timeout,
		endPoints: apputils.Map(j.Endpoints, func(v string) *string { return apputils.StrPointer(v) }),
		doSignal:  make(chan struct{}, 2),
	})
}

func (a *action) getEndpoints() []*string {
	var endpoints []*string
	// Define a map to take unique
	endpointMap := map[*string]struct{}{}
	for _, j := range a.jobs {
		for _, endpoint := range j.endPoints {
			if _, ok := endpointMap[endpoint]; !ok {
				// Add to endpoints
				endpoints = append(endpoints, endpoint)
				// Add to map for the next checking
				endpointMap[endpoint] = struct{}{}
			}
		}
	}
	return endpoints
}

func (a *action) notify(fullEndpoint string) {
	for _, j := range a.jobs {
		for _, endpoint := range j.endPoints {
			if strings.HasPrefix(fullEndpoint, *endpoint) {
				// Send a event to doSignal
				j.doSignal <- struct{}{}
			}
		}
	}
}

func (a *action) ListenToDoSignal() {
	// Start listening for events.
	go func() {
		for {
			select {
			case event, ok := <-a.watcher.Events:
				if !ok {
					return
				}
				// Detect event.Name belongs which action
				a.notify(event.Name)
			case err, ok := <-a.watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	// Add paths.
	for _, endpoint := range a.getEndpoints() {
		if err := a.watcher.Add(*endpoint); err != nil {
			log.Fatal(err)
		}
	}

	// Listen to doSignal
	for _, j := range a.jobs {
		go func(j *job) {
			for {
				// Take data from channel
				<-j.doSignal
				// Just do action when it's the last event.
				// That means the length is zero
				if len(j.doSignal) == 0 {
					// Sleep a bit and check again to be sure
					time.Sleep(500 * time.Millisecond)
					if len(j.doSignal) == 0 {
						go func() {
							log.Println("Do CMD:", *j.command)
							// Start a long-running process, capture stdout and stderr
							theActionCMD := cmd.NewCmd(*j.shellType, "-c", *j.command)
							theActionCMD.Start() // non-blocking
							// Use a goroutine to stop theActionCMD when the context timeout is reached
							// Take stdout, stderr in realtime
							currentStdOutIndex := 0
							currentStdErrIndex := 0
							// Stop theActionCMD when it reached the timeout
							ctx, cancel := context.WithTimeout(context.Background(), *j.timeout)
							defer cancel()
							go func() {
								select {
								case <-ctx.Done():
									err := theActionCMD.Stop()
									if err != nil {
										log.Println(err)
									}
								}
							}()

							steptime := 100 * time.Millisecond
							var i int64
							for i = 0; i < j.timeout.Milliseconds()/steptime.Milliseconds(); i++ {
								// The status will include all result from the beginning
								theActionStatus := theActionCMD.Status()
								// Sleep a bit before re-checking
								time.Sleep(steptime)
								// Print stdout
								stdoutSize := len(theActionStatus.Stdout)
								if stdoutSize > currentStdOutIndex {
									for i := currentStdOutIndex; i < stdoutSize; i++ {
										log.Println("StdOut:", theActionStatus.Stdout[i])
									}
									currentStdOutIndex = stdoutSize
								}
								// Print stderr
								stderrSize := len(theActionStatus.Stderr)
								if stderrSize > currentStdErrIndex {
									for i := currentStdErrIndex; i < stderrSize; i++ {
										log.Println("StdErr:", theActionStatus.Stderr[i])
									}
									currentStdErrIndex = stderrSize
								}
								// Break the loop when the action is completed
								if theActionStatus.Complete {
									log.Printf("Exit %d, command: %s", theActionStatus.Exit, *j.command)
									break
								}
							}
						}()

					}
				}

			}
		}(j)
	}
}
