package main

import (
	"bytes"
	"errors"
	"os/exec"
	"strconv"
)

func HandleGit(mode string, v GitConfig, run bool, printOutput bool, home string) (bool, error) {
	var job Job
	switch mode {
	case "clone":
		job = CreateJobFromFunction(func() error {
			return GitClone(v.Rep, v.Directory, home, v.PersonalToken)
		}, mode+" "+v.Name)
	case "pull":
		job = CreateJobFromFunction(func() error {
			err := GitCreateRemote(v.Directory, home, v.Rep)
			if err != nil {
				return err
			} else {
				return GitPull(v.Directory, home, v.PersonalToken)
			}
		}, mode+" "+v.Name)
	default:
		return false, errors.New("Not supported Mode: " + mode)
	}

	var err error

	if run {
		err = job.RunJob(printOutput)
	} else {
		err = job.RunJobBackground(printOutput)
	}

	return err == nil, err
}
func HandleBackup(cmd *exec.Cmd, name string, printOutput bool, test bool, run bool) error {
	job := CreateJobFromCommand(cmd, name)
	var err error
	if test {
		err = job.DontRun(printOutput)
	} else {
		if run {
			err = job.RunJob(printOutput)
		} else {
			err = job.RunJobBackground(printOutput)
		}
	}

	return err

}

func HandleMount(job Job, printOutput bool, test bool, run bool, buffer bytes.Buffer) bool {
	var err error
	if test {
		err = job.DontRun(printOutput)
		return handleError(job, err, ERROR_RUNMOUNT, buffer)
	} else {
		if run {
			err = job.RunJob(printOutput)
			return handleError(job, err, ERROR_RUNMOUNT, buffer)
		} else {
			err = job.RunJobBackground(printOutput)
			return handleError(job, err, ERROR_RUNMOUNT, buffer)
		}
	}
}

func HandleMountFolders(cmds []*exec.Cmd, printOutput bool, test bool, run bool) (string, bool) {
	ok := true
	var buffer bytes.Buffer
	for k, v := range cmds {
		job := CreateJobFromCommand(v, "mount"+strconv.Itoa(k))
		if !HandleMount(job, printOutput, test, run, buffer) {
			ok = false
		}
	}
	return buffer.String(), ok

}

func DoSeal(token string) error{
	err := Seal(AgentConfiguration.VaultConfig, token)
	if err != nil {
		return err
	}
	b, err := IsSealed(AgentConfiguration.VaultConfig)
	if err != nil {
		return err
	}

	if b{
		return nil
	} else{
		return errors.New(ERROR_VAULT_NOT_SEALED)
	}
}

func DoUnseal(token string) (bool,error ) {
	resp, err := Unseal(AgentConfiguration.VaultConfig, token)
	if err != nil {
		return false,err
	}
	Sugar.Info(REST_VAULT_SEAL_MESSAGE, resp.Sealed)
	return resp.Sealed, nil
}

func DoMountVerbose(token string) (string, error) {
	return DoMount(token,false,true,false,true)
}

func DoMount(token string, debug bool, printOutput bool, test bool, run bool) (string, error) {
	config, err := CreateConfigFromVault(token, AgentConfiguration.Hostname, AgentConfiguration.VaultConfig)
	if err != nil {
		return "", err
	}

	err = config.GetGocryptConfig()
	if err != nil {
		return "", err
	}
	out := MountFolders(config.Agent.HomeFolder, config.Gocrypt)

	if debug {
		Sugar.Debug("Config", config.Gocrypt)
		for k, v := range out {
			Sugar.Info("Command", k, ": ", v.String())
		}
	}
	str, ok := HandleMountFolders(out, printOutput, test, run)
	if ok {
		return str, nil
	} else {
		return str, errors.New(ERROR_RUNMOUNT)
	}
}

func DoBackupVerbose(token string, mode string) error {
	return DoBackup(token, mode, true, false,false,true)
}

func DoBackup(token string, mode string, printOutput bool, debug bool, test bool, run bool) error {
	config, err := CreateConfigFromVault(token, AgentConfiguration.Hostname, AgentConfiguration.VaultConfig)
	if err != nil {
		return err
	}

	err = config.GetResticConfig()
	if err != nil {
		return err
	}

	var cmd *exec.Cmd
	switch mode {
	case "init":
		cmd = InitRepo(config.Restic.Environment, config.Agent.HomeFolder)
	case "exist":
		cmd = ExistsRepo(config.Restic.Environment, config.Agent.HomeFolder)
	case "check":
		cmd = CheckRepo(config.Restic.Environment, config.Agent.HomeFolder)
	case "backup":
		cmd = Backup(
			config.Restic.Path,
			config.Restic.Environment,
			config.Agent.HomeFolder,
			config.Restic.ExcludePath,
			2000,
			2000)
	case "unlock":
		cmd = UnlockRepo(config.Restic.Environment, config.Agent.HomeFolder)
	case "list":
		cmd = ListRepo(config.Restic.Environment, config.Agent.HomeFolder)
		printOutput = true
	case "forget":
		cmd = ForgetRep(config.Restic.Environment, config.Agent.HomeFolder)
	default:
		return errors.New("Not supported Mode: " + mode)
	}
	if debug {
		Sugar.Debug("Command: ", cmd.String())
		Sugar.Info("Config", config.Restic)
	}
	return HandleBackup(cmd, mode, printOutput, test, run)
}

func DoGitVerbose(token string, mode string)( string,bool,error ){
	return DoGit(token,mode, true,true)
}

func DoGit(token string, mode string, run bool, printOutput bool) ( string,bool,error ){
	config, err := CreateConfigFromVault(token, AgentConfiguration.Hostname, AgentConfiguration.VaultConfig)
	if err != nil {
		return "",false,err
	}

	err = config.GetGitConfig()
	if err != nil {
		return "", false,err
	}

	var buffer bytes.Buffer
	ok := true
	for _, v := range config.Git {
		ok, err = HandleGit(mode, v, run, printOutput, config.Agent.HomeFolder)
		if !ok && err != nil {
			buffer.WriteString("\nJob: " + v.Name + " " + err.Error())
		}
	}
	return buffer.String(), ok,nil
}
