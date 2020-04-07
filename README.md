# spring-boot-operator
快速体验
```shell script
kubectl apply -f https://github.com/goudai/spring-boot-operator/blob/master/deployment.yaml
```
等待启动完成
```shell script
kubectl get po -A | grep spring-boot-operator
```
编写第一个 spring-boot demo
```yaml
apiVersion: springboot.qingmu.io/v1alpha1
kind: SpringBootApplication
metadata:
  name: operator-demo
  namespace: default
spec:
  springBoot:
    version: v1.0.0
    replicas: 1
```
部署
```shell script
kubectl apply -f https://github.com/goudai/spring-boot-operator/blob/master/demo1.yaml
```


