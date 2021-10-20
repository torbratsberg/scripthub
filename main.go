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

func getHomeDir() string {
	homeDir, err := os.UserHomeDir()
	check(err)

	return filepath.Join(homeDir)
}

var configScriptsFilePath string = filepath.Join(getHomeDir(), "/.config/scripthub/scripts")
var configScripthubPath string = filepath.Join(getHomeDir(), "/.config/scripthub")

// Error checker
func check(e error) {
	if e != nil {
		log.Fatal(e)
	}
}

// Returns a list of all the scripts in the config file
func getScripts() []script {
	scriptPaths, err := os.ReadFile(configScriptsFilePath)
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

// Appends a script to the script file
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
	file, err := os.OpenFile(configScriptsFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

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

// Opens the editable script file in the EDITOR env variable
func editScript(editScript string) error {
	// Get all the scripts
	scripts := getScripts()

	// Find the editable path for specified script
	for i := 0; i < len(scripts); i += 1 {
		if scripts[i].name == editScript {
			editScript = scripts[i].editable
			break
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

// Removes script from the script file
func removeScript(name string) error {
	scriptPaths, err := os.ReadFile(configScriptsFilePath)
	check(err)

	scriptLines := string(scriptPaths)

	// Remove line starting with {name}
	re := regexp.MustCompile("(?m)[\r\n]+^" + name + ".*$")
	res := re.ReplaceAllString(scriptLines, "")

	os.WriteFile(configScriptsFilePath, []byte(res), 0644)

	return nil
}

// Runs the script with the same name as passed in as argument
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

	if executablePath == "" {
		return fmt.Errorf("Error: Could not find script \"%s\" in scripts", name)
	}

	cmd := exec.Command(executablePath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	err := cmd.Run()
	check(err)

	return nil
}

func getPath(name string, option string) (res string, err error) {
	// Get all the scripts
	scripts := getScripts()

	var temp script

	// Find the editable path for specified script
	for i := 0; i < len(scripts); i += 1 {
		if scripts[i].name == name {
			temp = scripts[i]
			break
		}
	}

	if temp.editable == "" || temp.executable == "" {
		err = fmt.Errorf("Error: Could not find script \"%s\" in scripts", name)
	} else {
		if option == "executable" || option == "x" {
			res = temp.executable
		} else if option == "editable" || option == "e" {
			res = temp.editable
		} else {
			res += fmt.Sprintf("Name       : %s\n", temp.name)
			res += fmt.Sprintf("Editable   : %s\n", temp.editable)
			res += fmt.Sprintf("Executable : %s\n", temp.executable)
		}
	}

	return
}

func setup() {
	if _, err := os.Stat(configScriptsFilePath); os.IsNotExist(err) {
		fmt.Println("Could not find scripts file. Generating one.")
		fmt.Println("")

		if _, err := os.Stat(configScripthubPath); os.IsNotExist(err) {
			err := os.Mkdir(configScripthubPath, 0666)
			check(err)
			err = os.WriteFile(configScriptsFilePath, []byte(""), 0777)
			check(err)
			fmt.Println("Generated scripts file at: " + configScriptsFilePath)
		}
	} else {
		fmt.Println("Scripts file found. No setup needed")
		fmt.Println("Here are the scripts in your library: ")

		scriptStructs := getScripts()
		fmt.Println("")
		for i := 0; i < len(scriptStructs); i += 1 {
			fmt.Println("Name       : " + scriptStructs[i].name)
			fmt.Println("Executable : " + scriptStructs[i].executable)
			fmt.Println("Editable   : " + scriptStructs[i].editable)
			fmt.Println("============")
		}
	}

}

func main() {
	var newScriptStruct script
	var specifier string

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
						DefaultText: "",
						Destination: &newScriptStruct.name,
					},
					&cli.StringFlag{
						Name:        "editable",
						Aliases:     []string{"e"},
						Usage:       "Path to editable",
						Value:       "",
						DefaultText: "Same as --executable",
						Destination: &newScriptStruct.editable,
					},
					&cli.StringFlag{
						Name:        "executable",
						Aliases:     []string{"x"},
						Usage:       "Path to executable",
						Required:    true,
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
			{
				Name:      "path",
				Aliases:   []string{"p"},
				Usage:     "Get paths of a script",
				ArgsUsage: "{script name}",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "specifier",
						Aliases:     []string{"s"},
						Usage:       "Pass in `x` to only get executable path, or `e` to only get editable path",
						Destination: &specifier,
					},
				},
				Action: func(c *cli.Context) error {
					name := c.Args().First()
					res, err := getPath(name, specifier)
					fmt.Println(res)
					check(err)
					return nil
				},
			},
			// ================================================================
			{
				Name:  "setup",
				Usage: "Set up scripthub",
				Action: func(c *cli.Context) error {
					setup()
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
