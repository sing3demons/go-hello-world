1. 
$ kubectl apply -f k8s/00-namespace.yml 
$ kubectl get ns 

2.
$ kubectl apply -f k8s/01-deployment.yml
$ kubectl get pods -n go-hello-world

3.
$ kubectl apply -f k8s/02-service.yml  
$ kubectl get svc -n go-hello-world


clean 
$ kubectl delete -f k8s/ 