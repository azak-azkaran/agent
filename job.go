package main

import (
	"bytes"
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
	finished    bool
}

func LogJobStatus(job *Job) {
	if !job.printOutput {
		return
	}

	if job.Stdout.Len() > 0 {
		Sugar.Info(job.Stdout.String())
	} else {
		Sugar.Info("No Output in stdout")
	}

	if job.Stderr.Len() > 0 {
		Sugar.Info(job.Stderr.String())
	} else {
		Sugar.Info("No Output in stderr")
	}
}

func (job *Job) IsFinished() bool {
	return job.finished
}

func (job *Job) QueueStatus() {
	job.finished = true
	if job.Cmd != nil && job.Cmd.Process == nil {
		Sugar.Info("Process not found")
		return
	}
	LogJobStatus(job)
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
	jobmap.Set(name, &job)
	return job
}

func CreateJobFromCommand(cmd *exec.Cmd, name string) Job {
	if jobmap == nil {
		jobmap = cmap.New()
	}

	if jobmap.Has(name) {
		v, ok := jobmap.Get(name)
		if ok {
			oldCmd := v.(*Job)
			if oldCmd.Cmd.Process != nil {
				Sugar.Info("Found job:", name, "\tPID: ", oldCmd.Cmd.Process.Pid)
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
	jobmap.Set(name, &job)
	return job
}

func (job *Job) RunJob(printOutput bool) error {
	job.printOutput = printOutput
	Sugar.Info("Starting job: ", job.Name)
	return job.doJob()
}

func (job *Job) doJob() error {
	err := job.Function()
	job.QueueStatus()
	jobmap.Set(job.Name, job)
	return err
}

func (job *Job) RunJobBackground(printOutput bool) error {
	go func() {
		Sugar.Info("Starting job in background: ", job.Name)
		job.printOutput = printOutput
		err := job.doJob()
		if err != nil {
			Sugar.Error("ERROR: ", err, "\n", job.Stderr.String())
		}
	}()
	return nil
}

func (job *Job) DontRun(printOutput bool) error {
	job.printOutput = printOutput

	if job.Cmd != nil {
		Sugar.Info("Not Runing: ", job.Cmd)
	} else {
		Sugar.Info("Not Runing: ", job.Name)
	}

	job.QueueStatus()
	return nil
}
