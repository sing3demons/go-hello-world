create-database:
	kubectl apply -f k8s/database/00-mongo-namespace.yml
	kubectl apply -f k8s/database/01-mongodb-secrets.yaml
	kubectl apply -f k8s/database/02-mongodb-pvc.yml 
	kubectl apply -f k8s/database/03-mongo-deployment.yml 
	kubectl apply -f k8s/database/04-mongo-service.yml 

run:
	kubectl apply -f k8s/00-namespace.yml 
	kubectl apply -f k8s/01-secret.yml 
	kubectl apply -f k8s/02-deployment.yml 
	kubectl apply -f k8s/03-service.yml 
	kubectl apply -f k8s/04-ingress.yml
