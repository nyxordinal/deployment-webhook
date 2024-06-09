package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/gin-gonic/gin"
)

type App struct {
	App               string `json:"app"`
	Token             string `json:"token"`
	DockerImage       string `json:"docker_image"`
	DockerComposeFile string `json:"docker_compose_file"`
	CommitSha         string `json:"commit_sha"`
}

type Data struct {
	Data []App `json:"data"`
}

var logger = log.Default()

func main() {
	r := gin.Default()

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	r.POST("/deploy", func(c *gin.Context) {
		var requestApp App
		if err := c.BindJSON(&requestApp); err != nil {
			return
		}

		logger.Printf("received deploy request: %v\n", requestApp)

		jsonFile, err := os.Open("data.json")
		if err != nil {
			logger.Println(err)
		}
		defer jsonFile.Close()

		byteValue, _ := ioutil.ReadAll(jsonFile)
		var data Data
		json.Unmarshal(byteValue, &data)

		isTriggerDeployment := false
		var foundApp App
		for _, a := range data.Data {
			if a.App == requestApp.App && a.Token == requestApp.Token {
				logger.Printf("found app: %s for deployment\n", a.App)
				a.CommitSha = requestApp.CommitSha
				isTriggerDeployment = true
				foundApp = a
				break
			}
		}

		if isTriggerDeployment {
			go deploy(foundApp)
			c.JSON(200, gin.H{
				"message": "deploy process triggered",
			})
		} else {
			c.JSON(400, gin.H{
				"message": "app not found",
			})
		}

	})
	r.Run()
}

func deploy(app App) {
	logger.Printf("start deploying %s\n", app.App)

	logger.Printf("deploying %s, removing current deployment...\n", app.App)
	cmd := exec.Command("docker-compose", "-f", app.DockerComposeFile, "down")
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		logger.Printf("could not run command docker-compose down: %s\n", err.Error())
		return
	}

	logger.Printf("deploying %s, commit: %s\n", app.App, app.CommitSha)
	dockerImageWithCommitSha := app.DockerImage + ":" + app.CommitSha
	logger.Printf("pulling docker image: %s\n", dockerImageWithCommitSha)
	cmd = exec.Command("docker", "pull", dockerImageWithCommitSha)
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		logger.Printf("could not run command docker pull: %s\n", err)
		return
	}

	logger.Printf("updating docker compose file for %s\n", app.App)
	err := updateDockerComposeImage(app.DockerComposeFile, app.DockerImage, dockerImageWithCommitSha)
	if err != nil {
		logger.Printf("could not update Docker Compose file: %s\n", err.Error())
		return
	}

	logger.Printf("deploying %s with newest docker image...\n", app.App)
	cmd = exec.Command("docker-compose", "-f", app.DockerComposeFile, "up", "-d")
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		logger.Printf("could not run command docker-compose up: %s\n", err.Error())
		return
	}
	logger.Printf("%s deployment done\n", app.App)
}

func updateDockerComposeImage(filePath, dockerImageName, newImage string) error {
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}

	lines := strings.Split(string(content), "\n")

	for i, line := range lines {
		if strings.Contains(line, fmt.Sprintf("image: %s", dockerImageName)) {
			lines[i] = fmt.Sprintf("    image: %s", newImage)
		}
	}

	newContent := strings.Join(lines, "\n")

	err = ioutil.WriteFile(filePath, []byte(newContent), 0644)
	if err != nil {
		return err
	}

	return nil
}
