-include .env
export

PORT=8080 
IMAGE_URL="asia-southeast1-docker.pkg.dev/scheduler-411812/cloud-run-source-deploy/tele-scheduler:latest"

.PHONY: tidy 
tidy:
	go mod tidy

.PHONY: build-local
build-local:
	docker build -t sched-app:latest . 
	
.PHONY: run-local
run-local: 
	echo "Building docker image..."
	echo "publishing to port 9090. Mapped to port 8080 in container."
	docker run -p 9090:${PORT} -e PORT=${PORT} -e AUTH_USERNAME=${AUTH_USERNAME} -e AUTH_PASSWORD=${AUTH_PASSWORD} sched-app:latest

.PHONY: deploy-cloud
deploy-cloud:
# gcloud run deploy --source
	gcloud builds submit --tag ${IMAGE_URL} /Users/samuelyip/Documents/projs/tele-scheduler
	gcloud run deploy tele-scheduler --image ${IMAGE_URL} --region asia-southeast1 --allow-unauthenticated