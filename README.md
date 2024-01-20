# Tele message scheduler BE service
- exposes HTTP endpoints and independent from FE service


# Deployment (gcloud)
- URL: https://tele-scheduler-qun2ob4jna-as.a.run.app
- Google cloud run: https://console.cloud.google.com/run?hl=en&project=scheduler-411812&supportedpurview=organizationId,folder,project 

### Secret management
- create secrets in secret manager
- reference the secrets created in secret manager in the deployment of the instance

**required env vars**
- store AUTH_USERNAME & AUTH_PASSWORD in local password manager
- get from samyipsh

