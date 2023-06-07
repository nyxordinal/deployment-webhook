package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"os/exec"

	"github.com/gin-gonic/gin"
)

type App struct {
	App               string `json:"app"`
	Token             string `json:"token"`
	DockerImage       string `json:"docker_image"`
	DockerComposeFile string `json:"docker_compose_file"`
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
			if a.App == requestApp.App || a.Token == requestApp.Token {
				logger.Printf("found app: %s for deployment\n", a.App)
				isTriggerDeployment = true
				foundApp = a
				break
			}
		}

		if isTriggerDeployment {
			deploy(foundApp)
		}

		c.JSON(200, gin.H{
			"message": "deploy process triggered",
		})
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
	}

	logger.Printf("deploying %s, pulling newest docker image...\n", app.App)
	cmd = exec.Command("docker", "pull", app.DockerImage)
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		logger.Printf("could not run command docker pull: %s\n", err)
	}

	logger.Printf("deploying %s with newest docker image...\n", app.App)
	cmd = exec.Command("docker-compose", "-f", app.DockerComposeFile, "up", "-d")
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		logger.Printf("could not run command docker-compose up: %s\n", err.Error())
	}
}
