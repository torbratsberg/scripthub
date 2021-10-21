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

/*
	==========
	Utillities
	==========
*/

// Returns the users home directory
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

// Returns the script struct of a given script
func getScriptStruct(name string) (res script, err error) {
	scripts := getScripts()

	// Find the specified script
	for i := 0; i < len(scripts); i += 1 {
		if scripts[i].name == name {
			res = scripts[i]
			break
		}
	}

	if res.name == "" {
		err = fmt.Errorf("Error: Could not find script \"%s\" in scripts", name)
	}
	return
}

/*
	================
	Action functions
	================
*/

// Appends a script to the script file
func addScript(newScript script) (err error) {
	cwd, err := os.Getwd()
	if err != nil {
		return
	}

	// Make the executable path
	newScript.executable = filepath.Join(cwd, newScript.executable)

	// Check if editable is set, else use executable
	if newScript.editable == "" {
		newScript.editable = newScript.executable
	} else {
		newScript.editable = filepath.Join(cwd, newScript.editable)
	}

	// Write new script entry to the scripts file
	file, err := os.OpenFile(configScriptsFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}

	defer file.Close()

	// Put togheter the script entry string and write to file
	scriptString := newScript.name + " : " + newScript.executable + " : " + newScript.editable + "\n"
	if _, err = file.WriteString(scriptString); err != nil {
		return
	}
	file.Close()

	return
}

// Opens the editable script file in the EDITOR env variable
func editScript(name string) (err error) {
	// Get all the scripts
	scripts := getScripts()

	var scriptPath string

	// Find the editable path for specified script
	for i := 0; i < len(scripts); i += 1 {
		if scripts[i].name == name {
			scriptPath = scripts[i].editable
			break
		}
	}

	if scriptPath == "" {
		return fmt.Errorf("Error: Could not find script \"%s\" in scripts", name)
	}

	// Open script in $EDITOR env variable (Defaults to `vim`)
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vim"
	}

	cmd := exec.Command(editor, scriptPath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	err = cmd.Run()
	check(err)

	return
}

// Removes script from the script file
func removeScript(name string) (err error) {
	scriptsFile, err := os.ReadFile(configScriptsFilePath)

	scriptLines := string(scriptsFile)

	// Remove line starting with {name}
	re := regexp.MustCompile("(?m)[\r\n]+^" + name + ".*$")
	res := re.ReplaceAllString(scriptLines, "")

	// Overwrite the scripts file
	os.WriteFile(configScriptsFilePath, []byte(res), 0777)

	return
}

// Runs the script with the same name as passed in as argument
func runScript(name string) (err error) {
	var executablePath string

	res, err := getScriptStruct(name)
	check(err)

	executablePath = res.executable
	if executablePath == "" {
		err = fmt.Errorf("Error: Could not find script \"%s\" in scripts", name)
	}

	cmd := exec.Command(executablePath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	err = cmd.Run()
	check(err)

	return
}

// Returns the path of the given script, can be specified using the option param
func getPath(name string, option string) (res string, err error) {
	temp, err := getScriptStruct(name)
	check(err)

	if option == "executable" || option == "x" {
		res = temp.executable
	} else if option == "editable" || option == "e" {
		res = temp.editable
	} else {
		res += fmt.Sprintf("Name       : %s\n", temp.name)
		res += fmt.Sprintf("Editable   : %s\n", temp.editable)
		res += fmt.Sprintf("Executable : %s\n", temp.executable)
	}

	return
}

// Sets up scripthub files and folders is needed
func setup() {
	// If scripts file does not exist is scripthub config folder
	if _, err := os.Stat(configScriptsFilePath); os.IsNotExist(err) {
		fmt.Println("Could not find scripts file. Generating one.")
		fmt.Println("")

		if _, err := os.Stat(configScripthubPath); os.IsNotExist(err) {
			// If scripthub config folder does not exist
			err := os.Mkdir(configScripthubPath, 0666)
			check(err)
			err = os.WriteFile(configScriptsFilePath, []byte(""), 0777)
			check(err)
			fmt.Printf("Generated scripts file at: %s\n", configScriptsFilePath)
		} else {
			// If scripthub config folder does exist, but scripts file doesn't
			err = os.WriteFile(configScriptsFilePath, []byte(""), 0777)
			check(err)
			fmt.Printf("Generated scripts file at: %s\n", configScriptsFilePath)
		}
	} else {
		// Scriphub is already set up
		fmt.Println("Scripts file found. No setup needed")
		fmt.Println("Here are the scripts in your library: ")

		scriptStructs := getScripts()
		fmt.Println("")
		for i := 0; i < len(scriptStructs); i += 1 {
			fmt.Printf("Name       : %s\n", scriptStructs[i].name)
			fmt.Printf("Executable : %s\n", scriptStructs[i].executable)
			fmt.Printf("Editable   : %s\n", scriptStructs[i].editable)
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
				Action: func(*cli.Context) (err error) {
					scriptStructs := getScripts()

					for i := 0; i < len(scriptStructs); i += 1 {
						fmt.Println("Name       : " + scriptStructs[i].name)
						fmt.Println("Executable : " + scriptStructs[i].executable)
						fmt.Println("Editable   : " + scriptStructs[i].editable)
						fmt.Println("============")
					}

					return
				},
			},
			// ================================================================
			{
				Name:    "edit",
				Aliases: []string{"e"},
				Usage:   "Edit the given script",
				Action: func(c *cli.Context) (err error) {
					editScriptName := c.Args().First()
					err = editScript(editScriptName)
					check(err)
					return
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
				Action: func(c *cli.Context) (err error) {
					err = addScript(newScriptStruct)
					check(err)
					return
				},
			},
			// ================================================================
			{
				Name:      "remove",
				Aliases:   []string{"rm"},
				Usage:     "Remove a script from your library",
				ArgsUsage: "{script name}",
				Action: func(c *cli.Context) (err error) {
					name := c.Args().First()
					err = removeScript(name)
					check(err)
					return
				},
			},
			// ================================================================
			{
				Name:      "run",
				Aliases:   []string{"r"},
				Usage:     "Run a script from your library",
				ArgsUsage: "{script name}",
				Action: func(c *cli.Context) (err error) {
					name := c.Args().First()
					err = runScript(name)
					check(err)
					return
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
						Destination: &specifier,
					},
				},
				Action: func(c *cli.Context) (err error) {
					name := c.Args().First()
					res, err := getPath(name, specifier)
					fmt.Println(res)
					check(err)
					return
				},
			},
			// ================================================================
			{
				Name:  "setup",
				Usage: "Set up scripthub",
				Action: func(c *cli.Context) (err error) {
					setup()
					return
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
