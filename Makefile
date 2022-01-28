create-database:
	kubectl apply -f k8s/database/00-mongo-namespace.yml 
	kubectl apply -f k8s/database/01-mongo-deployment.yml 
	kubectl apply -f k8s/database/02-mongo-service.yml 

run-go:
	kubectl apply -f k8s/00-namespace.yml 
	kubectl apply -f k8s/01-deployment.yml 
	kubectl apply -f k8s/02-service.yml 
	kubectl apply -f k8s/03-ingress.yml
