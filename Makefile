PROJECT_ID=glo-cli-ebay-datastreaming
GCP_REGION=us-west3

connect_to_project:
	gcloud config set project $(PROJECT_ID)
	gcloud auth application-default set-quota-project $(PROJECT_ID)

build_and_push_docker_image: connect_to_project
	gcloud builds submit --tag us-docker.pkg.dev/glo-cli-ebay-datastreaming/ebay-dsi/cache-updater

deploy_cloud_run: connect_to_project
	gcloud run deploy 'ebay-streaming-cache-updater' \
		--image us-docker.pkg.dev/glo-cli-ebay-datastreaming/ebay-dsi/cache-updater:latest \
      	--region=${GCP_REGION} \
        --update-secrets=REDIS_PASSWORD=redis-password:latest \
        --set-env-vars='REDIS_HOST=34.106.243.151,REDIS_PORT=6379' \
        --service-account=ebay-streaming-cache-updater@glo-cli-ebay-datastreaming.iam.gserviceaccount.com \
        --allow-unauthenticated