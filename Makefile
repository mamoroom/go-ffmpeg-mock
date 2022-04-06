dev: 
	docker-compose up --build

login:
	gcloud auth login

config:
	gcloud config set project xxx

deploy: config
	gcloud app deploy app-server-stg.yaml
