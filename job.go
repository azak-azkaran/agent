package main

import (
	"bytes"
	"log"
	"os/exec"

	cmap "github.com/orcaman/concurrent-map"
)

var jobmap cmap.ConcurrentMap

type Job struct {
	Cmd         *exec.Cmd
	Function    func() error
	Stdout      *bytes.Buffer
	Stderr      *bytes.Buffer
	Name        string
	printOutput bool
}

func Log(toQueue string, p bool) {
	if p {
		log.Println("INFO: " + toQueue)
	}
}

func (job *Job) QueueStatus() {
	if job.Cmd != nil && job.Cmd.Process == nil {
		log.Println("Process not found")
		return
	}

	if job.Stdout.Len() > 0 {
		Log(job.Stdout.String(), job.printOutput)
	} else {
		Log("No Output in stdout", job.printOutput)
	}

	if job.Stderr.Len() > 0 {
		Log(job.Stderr.String(), job.printOutput)
	} else {
		Log("No Output in stderr", job.printOutput)
	}

}

func CreateJobFromFunction(f func() error, name string) Job {
	if jobmap == nil {
		jobmap = cmap.New()
	}

	job := Job{
		Stdout:   new(bytes.Buffer),
		Stderr:   new(bytes.Buffer),
		Function: f,
		Name:     name,
	}
	jobmap.Set(name, job)
	return job
}

func CreateJobFromCommand(cmd *exec.Cmd, name string) Job {
	if jobmap == nil {
		jobmap = cmap.New()
	}

	if jobmap.Has(name) {
		v, ok := jobmap.Get(name)
		if ok {
			oldCmd := v.(Job)
			if oldCmd.Cmd.Process != nil {
				log.Println("Found job:", name, "\tPID: ", oldCmd.Cmd.Process.Pid)
			}
		}
	}

	job := Job{
		Cmd:      cmd,
		Stdout:   new(bytes.Buffer),
		Stderr:   new(bytes.Buffer),
		Function: cmd.Run,
		Name:     name,
	}

	cmd.Stdout = job.Stdout
	cmd.Stderr = job.Stderr
	jobmap.Set(name, job)
	return job
}

func (job *Job) RunJob(printOutput bool) error {
	job.printOutput = printOutput
	log.Println("Starting job: ", job.Name)
	return job.doJob()
}

func (job *Job) doJob() error {
	err := job.Function()
	job.QueueStatus()
	return err
}

func (job *Job) RunJobBackground(printOutput bool) error {
	go func() {
		log.Println("Starting job in background: ", job.Name)
		job.printOutput = printOutput
		err := job.doJob()
		if err != nil {
			log.Println("ERROR: ", err)
			log.Println(job.Stderr.String())
		}
	}()
	return nil
}

func (job *Job) DontRun(printOutput bool) error {
	job.printOutput = printOutput

	if job.Cmd != nil {
		log.Println("Not Runing: ", job.Cmd)
	} else {
		log.Println("Not Runing: ", job.Name)
	}

	job.QueueStatus()
	return nil
}
