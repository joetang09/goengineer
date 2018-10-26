package engineer

import (
	"bufio"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/spf13/cobra"
)

const (
	EnginerLoggerKey = "engineer"
)

var (
	appCMD        *cobra.Command
	daemon        *bool
	forever       *bool
	cmd           *exec.Cmd
	pName         string
	ec            = make(chan struct{})
	enginerLogger = GetLogger("enginer")
)

func init() {

	pName = os.Args[0]

	appCMD = &cobra.Command{Use: pName}

	daemon = BindCMDArgsBool("daemon", false, "daemon the process")
	forever = BindCMDArgsBool("forever", false, "forever the process")

	BindCMDArgsStrP(cf, "config", "f", defaultConfigFile, "config file path")

}

func BindCMDArgsStr(val *string, key, def, desc string) {

	appCMD.PersistentFlags().StringVar(val, key, def, desc)

}

func BindCMDArgsUint(val *uint, key string, def uint, desc string) {

	appCMD.PersistentFlags().UintVar(val, key, def, desc)

}

func BindCMDArgsStrP(val *string, key, short, def, desc string) {

	appCMD.PersistentFlags().StringVarP(val, key, short, def, desc)

}

func BindCMDArgsBool(key string, def bool, desc string) *bool {
	return appCMD.PersistentFlags().Bool(key, def, desc)
}

func BuildEnv(configable bool) {
	runtime.GOMAXPROCS(runtime.NumCPU())

	if err := appCMD.Execute(); err != nil {
		panic("start process : " + pName + " failed")
	}
	if configable {
		Use(ConfigCpnt{})
	}

	if !IsUsed(logPkg) {
		Use(LogCpnt{})
	}

	enginerLogger.Info("start process : ", pName)

}

func ConfigLog(c LogConfig) {
	Use(LogCpnt{}, &c)
}

func GetConfig() config {
	return configIns
}

func beDaemon() {
	if *daemon {

		args := []string{}
		for _, arg := range os.Args[1:] {
			if arg != "--daemon" {
				args = append(args, arg)
			}
		}
		cmd = exec.Command(os.Args[0], args...)
		cmd.Start()
		f := ""
		if *forever {
			f = "[forever]"
		}
		enginerLogger.Infof(" %s%s with PID [%d] is running...\n", os.Args[0], f, cmd.Process.Pid)
		exit(0)
	}
}

func beForever() bool {
	if *forever {
		args := []string{}
		for _, arg := range os.Args[1:] {
			if arg != "--forever" {
				args = append(args, arg)
			}
		}
		go func() {
			defer func() {
				ec <- struct{}{}
			}()
			for {
				if cmd != nil {
					cmd.Process.Kill()
				}
				cmd = exec.Command(os.Args[0], args...)
				cmdReaderStderr, err := cmd.StderrPipe()
				if err != nil {
					enginerLogger.Infof("start error : %s, restarting...", err)
					continue
				}
				cmdReader, err := cmd.StdoutPipe()
				if err != nil {
					enginerLogger.Infof("start error : %s, restarting...", err)
					continue
				}
				scanner := bufio.NewScanner(cmdReader)
				scannerStdErr := bufio.NewScanner(cmdReaderStderr)
				go func() {
					for scanner.Scan() {
						enginerLogger.Info("scan : ", scanner.Text())
					}
				}()
				go func() {
					for scannerStdErr.Scan() {
						enginerLogger.Info("scan std err : ", scannerStdErr.Text())
					}
				}()
				if err := cmd.Start(); err != nil {
					enginerLogger.Infof("start error : %s, restarting...", err)
					continue
				}
				pid := cmd.Process.Pid
				enginerLogger.Infof("worker[%s] with PID[%d] is running...", os.Args[0], pid)
				if err := cmd.Wait(); err != nil {
					enginerLogger.Infof("error : %s, restarting...", err)
					continue
				}
				enginerLogger.Infof("worker[%s] with PID[%d] unexpected exited, restarting...", os.Args[0], pid)
				time.Sleep(time.Second * 5)
			}
		}()

		return true
	}
	return false
}

func Start() {

	beDaemon()

	if !beForever() {
		for _, svr := range svrBox {

			go serverWrapper(svr)
		}
	}

	go handleSysSignal()
	<-ec
	exit(0)
}

func serverWrapper(svr Server) {
	defer func() {
		ec <- struct{}{}

	}()
	svr.Serve()
}

func Stop() {
	for _, svr := range svrBox {
		svr.Stop()
	}
}

func exit(i int) {
	os.Exit(i)
}

func handleSysSignal() {
	enginerLogger.Info("start monitor system signal...")
	sChan := make(chan os.Signal)
	for {
		signal.Notify(sChan, os.Interrupt, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
		sig := <-sChan
		enginerLogger.Infof("received signal : %v\n", sig)
		switch sig {
		case os.Interrupt, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
			Stop()
			if cmd != nil {
				enginerLogger.Infof("clean child process %d", cmd.Process.Pid)
				cmd.Process.Kill()
			}
			exit(0)

		}

	}

}
