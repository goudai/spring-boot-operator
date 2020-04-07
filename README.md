# spring-boot-operator
快速体验
```shell script
kubectl apply -f https://raw.githubusercontent.com/goudai/spring-boot-operator/master/manifests/deployment.yaml
```
对于拉取Docker.io很慢的同学可以使用阿里云镜像仓库
```shell script
kubectl apply -f https://raw.githubusercontent.com/goudai/spring-boot-operator/master/manifests/deployment_ali.yaml
```

等待启动完成
```shell script
kubectl get po -A | grep spring-boot-operator
```
编写第一个 spring-boot demo
Demo Github地址：https://github.com/goudai/operator-demo

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
kubectl apply -f https://raw.githubusercontent.com/goudai/spring-boot-operator/master/manifests/demo1.yaml
```


