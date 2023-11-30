PROJECT_ID=xxx
GCP_REGION=xxx

connect_to_project:
	gcloud config set project $(PROJECT_ID)
	gcloud auth application-default set-quota-project $(PROJECT_ID)

build_and_push_docker_image: connect_to_project
	gcloud builds submit --tag us-docker.pkg.dev/x/y/x:latest

deploy_cloud_run: connect_to_project
	gcloud run deploy 'ebay-streaming-cache-updater' \
		--imageus-docker.pkg.dev/x/y/x:latest \
      	--region=${GCP_REGION} \
        --update-secrets=REDIS_PASSWORD=redis-password:latest \
        --set-env-vars='REDIS_HOST=0.0.0.0,REDIS_PORT=6379' \
        --service-account=sa@some-project.iam.gserviceaccount.com