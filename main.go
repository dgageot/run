package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	if len(os.Args) <= 1 {
		return
	}

	aliasFiles, err := findAliasFiles()
	if err != nil {
		log.Fatal(err)
	}

	aliases, err := readAliases(aliasFiles)
	if err != nil {
		log.Fatal(err)
	}

	command := expand(os.Args[1:], aliases)

	cmd := exec.Command("docker", command...)
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout

	err = cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
}

func findAliasFiles() ([]string, error) {
	var aliasFiles []string

	currentPath, err := filepath.Abs(".")
	if err != nil {
		return nil, err
	}

	for {
		path := filepath.Join(currentPath, ".docker-aliases")

		if _, err = os.Stat(path); err == nil {
			aliasFiles = append([]string{path}, aliasFiles...)
		}

		parent := filepath.Dir(currentPath)
		if parent == currentPath {
			return aliasFiles, nil
		}

		currentPath = parent
	}
}

func expand(args []string, aliases map[string][]string) []string {
	if len(args) == 0 {
		return args
	}

	command, found := aliases[args[0]]
	if !found {
		return args
	}

	expanded := append(command, args[1:]...)

	if strings.HasPrefix(expanded[0], "@") {
		expanded[0] = strings.TrimLeft(expanded[0], "@")
		return expand(expanded, aliases)
	}

	return expanded
}

func readAliases(paths []string) (map[string][]string, error) {
	aliases := map[string][]string{}

	for _, path := range paths {
		file, err := os.Open(path)
		if err != nil {
			return nil, err
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()

			if strings.HasPrefix(line, "#") {
				continue
			}

			if len(strings.TrimSpace(line)) == 0 {
				continue
			}

			key, command, err := parseLine(line)
			if err != nil {
				return nil, err
			}

			aliases[key] = command
		}

		if err := scanner.Err(); err != nil {
			return nil, err
		}
	}

	return aliases, nil
}

func parseLine(line string) (key string, command []string, err error) {
	parts := strings.SplitN(line, "=", 2)
	if len(parts) != 2 {
		return "", nil, fmt.Errorf("Malformed alias: %s", line)
	}

	key = parts[0]
	command = strings.Split(parts[1], " ")
	return
}
