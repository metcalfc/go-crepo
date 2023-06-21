package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var cfgFile string

type Repository struct {
	Directory string `yaml:"directory"`
	Remote    string `yaml:"remote"`
	Refspec   string `yaml:"refspec"`
}

type Config []Repository

var config Config
var verbose bool

var rootCmd = &cobra.Command{
	Use:   "crepo",
	Short: "A CLI to manage a collection of git repositories",
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Clone repositories listed in crepo.yaml",
	Run: func(cmd *cobra.Command, args []string) {
		// Read config
		fmt.Println("Reading config")
		configData, err := ioutil.ReadFile(cfgFile)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		yaml.Unmarshal(configData, &config)

		fmt.Println(config)
		// Clone repos
		for _, repo := range config {
			fmt.Println(repo.Remote)

			if verbose {
				fmt.Printf("Cloning %s \n", repo.Remote)
			}

			r, err := git.PlainClone(repo.Directory, false, &git.CloneOptions{
				URL: repo.Remote,
			})
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			h, err := r.ResolveRevision(plumbing.Revision(repo.Refspec))
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			w, err := r.Worktree()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			err = w.Checkout(&git.CheckoutOptions{
				Hash: *h,
			})
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		}
	},
}

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Check if git repos are dirty",
	Run: func(cmd *cobra.Command, args []string) {

		configData, err := ioutil.ReadFile(cfgFile)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		yaml.Unmarshal(configData, &config)

		for _, repo := range config {
			r, err := git.PlainOpen(repo.Directory)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			w, err := r.Worktree()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			status, err := w.Status()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			if !status.IsClean() {
				fmt.Printf("%s is dirty\n", repo.Directory)
				os.Exit(1)
			}
		}
		os.Exit(0)
	},
}

var foreachCmd = &cobra.Command{
	Use:   "foreach",
	Short: "Execute a shell command in each repo",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println("Please provide a shell command to execute")
			os.Exit(1)
		}

		shellCmd := strings.Join(args, " ")

		configData, err := ioutil.ReadFile(cfgFile)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		yaml.Unmarshal(configData, &config)

		for _, repo := range config {
			cmd := exec.Command("sh", "-c", shellCmd)
			cmd.Dir = repo.Directory
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			err = cmd.Run()
			if err != nil {
				fmt.Printf("Error running command in %s: %s\n", repo.Directory, err)
				os.Exit(1)
			}
		}
	},
}

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate config file",
	Run: func(cmd *cobra.Command, args []string) {
		configData, err := ioutil.ReadFile(cfgFile)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		err = yaml.Unmarshal(configData, &config)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		for _, repo := range config {
			if repo.Directory == "" {
				fmt.Println("Directory missing for repo")
				os.Exit(1)
			}
			if repo.Remote == "" {
				fmt.Println("Remote missing for repo")
				os.Exit(1)
			}
			if repo.Refspec == "" {
				fmt.Println("Remote missing for repo")
				os.Exit(1)
			}
		}
		fmt.Println("Config file is valid")
	},
}

func main() {

	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "crepo.yaml", "config file")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose")
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(checkCmd)
	rootCmd.AddCommand(foreachCmd)
	rootCmd.AddCommand(validateCmd)
	rootCmd.Execute()
}
