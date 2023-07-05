package main

import (
	"dmfcta/action"
	"dmfcta/apputils"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"time"
)

func main() {
	// Create new watcher.
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()
	// User define
	myAction := action.New(watcher)
	// Add job
	actionMap := map[string]string{}
	criteriaMap := map[string]string{}
	shellMap := map[string]string{}
	timeoutMap := map[string]string{}

	actionKey := apputils.GetEnv("DMFCTA_ACTION_KEY", "ACTION")
	criteriaKey := apputils.GetEnv("DMFCTA_CRITERIA_KEY", "CRITERIA")
	shellKey := apputils.GetEnv("DMFCTA_SHELL_KEY", "SHELL")
	timeoutKey := apputils.GetEnv("DMFCTA_TIMEOUT_KEY", "TIMEOUT")
	pattern := fmt.Sprintf(`(%s|%s|%s|%s)_(\d+)=(.+)`, actionKey, criteriaKey, shellKey, timeoutKey)
	rg, _ := regexp.Compile(pattern)
	for _, kv := range os.Environ() {
		res := rg.FindAllStringSubmatch(kv, -1)
		for i := range res {
			//like Java: match.group(1), match.group(2), etc
			//fmt.Printf("key: %s, number: %s, value: %s\n", res[i][1], res[i][2], res[i][3])
			switch res[i][1] {
			case actionKey:
				actionMap[fmt.Sprintf("number_%s", res[i][2])] = res[i][3]
			case criteriaKey:
				criteriaMap[fmt.Sprintf("number_%s", res[i][2])] = res[i][3]
			case shellKey:
				shellMap[fmt.Sprintf("number_%s", res[i][2])] = res[i][3]
			case timeoutKey:
				timeoutMap[fmt.Sprintf("number_%s", res[i][2])] = res[i][3]
			}
		}
	}
	// Merge actions
	var jobs []action.JobInput
	for k, a := range actionMap {
		jobs = append(jobs, action.JobInput{
			CMD: a,
			SHELLTYPE: func() string {
				finalShell := ""
				if shell, ok := shellMap[k]; ok {
					finalShell = shell
				}
				if finalShell == "" {
					for _, s := range []string{action.ShShell, action.BashShell, action.ZshShell} {
						if _, err := exec.LookPath(s); err == nil {
							finalShell = s
							break
						}
					}
				}

				// Set default shell
				// Loop to check the existent shell, take one
				if _, err := exec.LookPath(finalShell); err != nil {
					return ""
				}
				return finalShell
			}(),
			Timeout: func() time.Duration {
				if timeout, ok := timeoutMap[k]; ok {
					// timeout is like: 3600, in second
					intTimeout, err := strconv.ParseInt(timeout, 10, 0)
					if err == nil {
						return time.Duration(intTimeout) * time.Second
					}
				}
				// Set default timeout
				return time.Minute
			}(),
			Endpoints: func() []string {
				if criteria, ok := criteriaMap[k]; ok {
					return []string{criteria}
				}
				return []string{}
			}(),
		})
	}

	// Check for jobs with the same action
	// Create a map to store unique jobs
	uniqueJobs := make(map[string]action.JobInput)
	// Filter and retrieve unique jobs
	for _, job := range jobs {
		key := fmt.Sprintf("%s-%s-%s", job.CMD, job.SHELLTYPE, job.Timeout)
		// Add the job to the map if it doesn't exist
		if _, ok := uniqueJobs[key]; !ok {
			uniqueJobs[key] = job
		} else {
			job.Endpoints = append(uniqueJobs[key].Endpoints, job.Endpoints...)
			uniqueJobs[key] = job
		}
	}
	for _, job := range uniqueJobs {
		myAction.Add(job)
		log.Printf("Action: %s", job.CMD)
		log.Printf("Shell: %s", job.SHELLTYPE)
		log.Printf("Timeout: %s", job.Timeout)
		log.Println("Endpoints:")
		for _, endpoint := range job.Endpoints {
			log.Printf("- %s", endpoint)
		}
		log.Println("-------------------------------------")
	}

	myAction.ListenToDoSignal()

	// Block main goroutine forever.
	<-make(chan struct{})
}
