package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/urfave/cli/v2"
)

type script struct {
	name       string
	executable string
	editable   string
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func getScripts() []script {
	scriptPaths, err := os.ReadFile(
		"/Users/torbratsberg/.config/scripthub/scripts",
	)
	check(err)

	scriptLines := strings.Split(string(scriptPaths), "\n")

	var scriptsStruct []script

	for i := 0; i < len(scriptLines)-1; i += 1 {
		split := strings.Split(scriptLines[i], " : ")

		scriptsStruct = append(
			scriptsStruct,
			script{
				name:       split[0],
				executable: split[1],
				editable:   split[2],
			},
		)
	}

	return scriptsStruct
}

func addScript(newScript script) error {
	// Make the executable path
	executablePath, err := os.Getwd()
	newScript.executable = filepath.Join(executablePath, newScript.executable)
	check(err)

	// Check if editable is set, else use executable
	if newScript.editable == "" {
		newScript.editable = newScript.executable
	} else {
		editablePath, err := os.Getwd()
		newScript.editable = filepath.Join(editablePath, newScript.editable)
		check(err)
	}

	// Write new script entry to the scripts file
	file, err := os.OpenFile("/Users/torbratsberg/.config/scripthub/scripts", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	if err != nil {
		log.Println(err)
		return err
	}

	defer file.Close()

	// Put togheter the script entry string
	scriptString := newScript.name + " : " + newScript.executable + " : " + newScript.editable + "\n"

	if _, err := file.WriteString(scriptString); err != nil {
		log.Println(err)
		return err
	}

	file.Close()

	return nil
}

func editScript(editScript string) error {
	// Get all the scripts
	scripts := getScripts()

	// Find the editable path for specified script
	for i := 0; i < len(scripts); i += 1 {
		if scripts[i].name == editScript {
			editScript = scripts[i].editable
		}
	}

	// Open script in neovim
	cmd := exec.Command(os.Getenv("EDITOR"), editScript)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	err := cmd.Run()
	check(err)

	return nil
}

func removeScript(name string) error {
	scriptPaths, err := os.ReadFile("/Users/torbratsberg/.config/scripthub/scripts")
	check(err)

	scriptLines := strings.Split(string(scriptPaths), "\n")

	var index int
	for ; index < len(scriptLines)-1; index += 1 {
		res, err := regexp.MatchString(name, scriptLines[index])
		check(err)

		fmt.Println(res)
		if res == true {
			break
		}
	}

	return nil
}

func runScript(name string) error {

	// Get all the scripts
	scripts := getScripts()

	var executablePath string

	// Find the editable path for specified script
	for i := 0; i < len(scripts); i += 1 {
		if scripts[i].name == name {
			executablePath = scripts[i].executable
		}
	}

	cmd := exec.Command(executablePath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	err := cmd.Run()
	check(err)

	return nil
}

func main() {
	var newScriptStruct script

	app := &cli.App{
		Name:  "scripthub",
		Usage: "Keep track of all your scripts",
		Commands: []*cli.Command{
			// ================================================================
			{
				Name:    "list",
				Aliases: []string{"ls"},
				Usage:   "List all available scripts",
				Action: func(*cli.Context) error {
					scriptStructs := getScripts()

					for i := 0; i < len(scriptStructs); i += 1 {
						fmt.Println("Name       : " + scriptStructs[i].name)
						fmt.Println("Executable : " + scriptStructs[i].executable)
						fmt.Println("Editable   : " + scriptStructs[i].editable)
						fmt.Println("============")
					}

					return nil
				},
			},
			// ================================================================
			{
				Name:    "edit",
				Aliases: []string{"e"},
				Usage:   "Edit the given script",
				Action: func(c *cli.Context) error {
					editScriptName := c.Args().First()
					err := editScript(editScriptName)
					check(err)
					return nil
				},
				ArgsUsage: "scriptName",
			},
			// ================================================================
			{
				Name:    "add",
				Aliases: []string{"a"},
				Usage:   "Add a script to your library",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "name",
						Aliases:     []string{"n"},
						Usage:       "Name of script to add",
						Required:    true,
						Value:       "myScript",
						DefaultText: "",
						Destination: &newScriptStruct.name,
					},
					&cli.StringFlag{
						Name:        "editable",
						Aliases:     []string{"ed"},
						Usage:       "Path to editable",
						Value:       "",
						DefaultText: "Same as --executable",
						Destination: &newScriptStruct.editable,
					},
					&cli.StringFlag{
						Name:        "executable",
						Aliases:     []string{"ex"},
						Usage:       "Path to executable",
						Required:    true,
						Value:       "FILE",
						DefaultText: "",
						Destination: &newScriptStruct.executable,
					},
				},
				Action: func(c *cli.Context) error {
					err := addScript(newScriptStruct)
					check(err)
					return nil
				},
			},
			// ================================================================
			{
				Name:      "remove",
				Aliases:   []string{"rm"},
				Usage:     "Remove a script from your library",
				ArgsUsage: "{script name}",
				Action: func(c *cli.Context) error {
					name := c.Args().First()
					err := removeScript(name)
					check(err)
					return nil
				},
			},
			// ================================================================
			{
				Name:      "run",
				Aliases:   []string{"r"},
				Usage:     "Run a script from your library",
				ArgsUsage: "{script name}",
				Action: func(c *cli.Context) error {
					name := c.Args().First()
					err := runScript(name)
					check(err)
					return nil
				},
			},
			// ================================================================
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
