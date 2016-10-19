package main

import (
	"code.cloudfoundry.org/cli/plugin"
	"fmt"
	"github.com/gdey/jsonpath"
	"os"
	"strings"
	"encoding/json"
)

type GetEnvPlugin struct {
	appName    string
	applicator jsonpath.Applicator
}

func main() {
	plugin.Start(new(GetEnvPlugin))
}

func (p *GetEnvPlugin) Run(cliConnection plugin.CliConnection, args []string) {

	if len(args) > 0 && args[0] == "CLI-MESSAGE-UNINSTALL" {
		return
	}

	switch args[0] {
	case "get-env":

		p.setup(args)

		env := p.fetchEnv(cliConnection)
		selectedValue := p.selectValue(env)

		fmt.Print(selectedValue)
	}
}

func (p *GetEnvPlugin) selectValue(env map[string]interface{}) interface{} {

	selectedValue, jsonPathError := p.applicator.Apply(env)

	if jsonPathError != nil {
		msg, _ := fmt.Printf("Failed to apply JSON path: %s", jsonPathError)
		fmt.Println(msg)
		os.Exit(1)
	}

	return selectedValue
}

func (p *GetEnvPlugin) setup(args []string) {

	if len(args) < 2 {
		fmt.Println("App name must be provided")
		os.Exit(1)
	}

	p.appName = args[1]

	if len(args) < 3 {
		fmt.Println("JSON-Path expression must be provided")
		os.Exit(1)
	}

	applicator, parseErr := jsonpath.Parse(args[2])

	if parseErr != nil {
		msg, _ := fmt.Printf("Failed to parse argument '%s' as valid JSON-path: %s", args[2], parseErr)
		fmt.Println(msg)
		os.Exit(1)
	}

	p.applicator = applicator
}

func (p *GetEnvPlugin) parseJsonPath(pathExpression string) jsonpath.Applicator {
	applicator, parseErr := jsonpath.Parse(pathExpression)

	if parseErr != nil {
		msg, _ := fmt.Printf("Failed to parse argument '%s' as valid JSON-path: %s", pathExpression, parseErr)
		fmt.Println(msg)
		os.Exit(1)
	}

	return applicator
}

func (p *GetEnvPlugin) fetchEnv(cliConnection plugin.CliConnection) map[string]interface{} {

	app, err := cliConnection.GetApp(p.appName)

	if err != nil {
		msg := fmt.Sprintf("Failed to retrieve enviroment for '%s'. %s", p.appName, err)
		fmt.Println(msg)
		os.Exit(1)
	}

	url := fmt.Sprintf("/v2/apps/%s/env", app.Guid)
	output, err := cliConnection.CliCommandWithoutTerminalOutput("curl", url)

	if err != nil {
		msg := fmt.Sprintf("Failed to retrieve enviroment for '%s'. %s", p.appName, err)
		fmt.Println(msg)
		os.Exit(1)
	}

	envAsString := strings.Join(output, "")

	env := make(map[string]interface{})
	json.Unmarshal([]byte(envAsString), &env)

	return env;
}

func fatalIf(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func (c *GetEnvPlugin) GetMetadata() plugin.PluginMetadata {
	return plugin.PluginMetadata{
		Name: "Get-Env",
		Commands: []plugin.Command{
			{
				Name:     "get-env",
				HelpText: "Get value from the environment from an env by a JSON path expression.",
				UsageDetails: plugin.Usage{
					Usage: "cf get-env APP_NAME JSON_PATH",
				},
			},
		},
	}
}
