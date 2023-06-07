# Nyxordinal Deployment Webhook

This app will listen to a webhook URL.
When the URL is called, a deployment process will be triggered based on the provided data in the request.

## How to Use

1. Copy the `data.json.example` file and rename it as `data.json`. Fill in the necessary data for your app. You can define multiple apps within the file.
   - `app`: The name of the app.
   - `token`: Verification token. Make sure to keep it in a secure place.
   - `docker_image`: Docker image name along with its tag.
   - `docker_compose_file`: Please use an absolute path to the docker-compose file. Ensure that its content does not disrupt any ongoing deployments on your machine.
2. Run the app using `go run main.go`. The app will listen on `port 8080`.
3. Make an HTTP `POST /deploy` call with the following request body to trigger the deployment process:

```json
{
  "app": "plutus",
  "token": "token1"
}
```

## Developer Team

Developed with passion by [Nyxordinal](https://nyxordinal.tech/)
