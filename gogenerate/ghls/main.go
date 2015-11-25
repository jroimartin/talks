package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"text/template"
)

const apiUrl = "https://api.github.com"

var (
	debugFlag  = flag.Bool("debug", false, "print debug messages")
	outputTmpl = template.Must(template.New("").Parse(outputText))
)

func main() {
	flag.Usage = usage
	flag.Parse()

	if flag.NArg() != 1 {
		usage()
		os.Exit(2)
	}

	username := flag.Arg(0)

	usr, err := getUser(username)
	if err != nil {
		log.Fatal(err)
	}

	debug("User:", usr)

	repos, err := getRepos(username)
	if err != nil {
		log.Fatal(err)
	}

	debug("Repos:", repos)

	data := struct {
		User  User
		Repos []Repo
	}{usr, repos}
	outputTmpl.Execute(os.Stdout, data)
}

func getUser(username string) (User, error) {
	resp, err := http.Get(fmt.Sprintf("%s/users/%s", apiUrl, username))
	if err != nil {
		return User{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return User{}, errors.New("status code is not 200 OK")
	}

	usr := User{}
	if err := json.NewDecoder(resp.Body).Decode(&usr); err != nil {
		return User{}, err
	}
	return usr, nil
}

func getRepos(username string) ([]Repo, error) {
	resp, err := http.Get(fmt.Sprintf("%s/users/%s/repos", apiUrl, username))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("status code is not 200 OK")
	}

	repos := []Repo{}
	if err := json.NewDecoder(resp.Body).Decode(&repos); err != nil {
		return nil, err
	}
	return repos, nil
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: ghls [options] username")
	fmt.Fprintln(os.Stderr, "options:")
	flag.PrintDefaults()
}

func debug(a ...interface{}) {
	if !*debugFlag {
		return
	}
	log.Println(a...)
}

const outputText = `Username: {{.User.Name}}
# Repos: {{.User.PublicRepos}}
Repositories:
{{range .Repos}}  - {{.Name}}: {{.Description}}
{{end}}`
