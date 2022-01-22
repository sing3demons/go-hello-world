1. ## Run command
   ```
   kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v1.1.1/deploy/static/provider/aws/deploy.yaml
   ```
    namespace/ingress-nginx created...

2. Run command to create namespace
   ```
   kubectl apply -f k8s/00-namespace.yml
   ```
   namespace/go-hello-world created

3. Check for namespace (go-hello-world)
   ```
   kubectl get ns
   ```
   NAME STATUS AGE
   default Active 78d
   go-hello-world Active 52s
   kube-node-lease Active 78d
   kube-public Active 78d
   kube-system Active 78d

4. Run command to create namespace
   ```
   kubectl apply -f k8s/01-deployment.yml
   ```
   deployment.apps/go-server created

5. Check for go-server deployment
   ```
   kubectl get pods -n go-hello-world
   ```
   NAME READY STATUS RESTARTS AGE
   go-server-5cd4c8c54-jr4h7 1/1 Running 0 73s
   go-server-5cd4c8c54-lgkmd 1/1 Running 0 73s

6. Create service for go-server
   ```
   kubectl apply -f k8s/02-service.yml  
   ```
   service/go-server created

7. Check for service go-server
   ```
   kubectl get svc -n go-hello-world
   ```
   NAME TYPE CLUSTER-IP EXTERNAL-IP PORT(S) AGE
   go-server ClusterIP 10.103.229.168 <none> 8080/TCP 47s

8. Create ingress rule to nginx service
   ```
   kubectl apply -f k8s/03-ingress.yml
   ```
   ingress.networking.k8s.io/myingress created

9. Check for ingress rule
   ```
   kubectl get ing -n go-hello-world
   ```
   NAME CLASS HOSTS ADDRESS PORTS AGE
   myingress <none> kubernetes.docker.internal localhost 80 103s

10. Test access service via ingress
    ```
    curl -X GET "http://kubernetes.docker.internal"
    ```
    "hello, world!"%

11. Run command to cleanup
    ```
    kubectl delete ns go-hello-world
    ```
    ```
    kubectl delete ns ingress-nginx
    ```
