# kubebuilder
Use kubebuilder to build my own operator

## Description
### Preparation
```sh
# 初始化
$ kubebuilder init --plugins go/v3 --domain huozj.io --license apache2 --owner "huo" --repo github.com/ZhengjunHUO/kubebuilder
# 创建api
$ kubebuilder create api --group cat --version v1alpha2 --kind Fufu
# 自定义结构并重新生成code
$ vim api/v1alpha2/fufu_types.go
$ make
# 生成对应manifests
$ make manifests 
# 安装crd (需要一个k8s cluster)
$ make install
# Implement reconciling logic and test under controllers/
# Code Test
$ make test
```
### Test
```sh
# Run the controller
$ make run

# In a new terminal, create a CR
$ kubectl create ns fufu
$ kubectl apply -f config/samples/cat_v1alpha2_fufu.yaml
$ kubectl get fufu,pod,svc,hpa -n fufu
NAME                          COLOR    REPLICAS   EXTERNALIP
fufu.cat.huozj.io/fufu-test   orange   2          172.18.0.101

NAME                                    READY   STATUS    RESTARTS   AGE
pod/fufu-test-deploy-76949c9d9d-2vdgx   1/1     Running   0          19s
pod/fufu-test-deploy-76949c9d9d-ctj9z   1/1     Running   0          34s

NAME                    TYPE           CLUSTER-IP    EXTERNAL-IP    PORT(S)        AGE
service/fufu-test-svc   LoadBalancer   10.96.234.7   172.18.0.101   80:30277/TCP   34s

NAME                                                REFERENCE                     TARGETS         MINPODS   MAXPODS   REPLICAS   AGE
horizontalpodautoscaler.autoscaling/fufu-test-hpa   Deployment/fufu-test-deploy   <unknown>/60%   2         5         0          4s

# Remove some resources and watch what happened
$ kubectl delete hpa fufu-test-hpa -n fufu
$ kubectl delete deploy fufu-test-deploy -n fufu
```

## Getting Started
You’ll need a Kubernetes cluster to run against. You can use [KIND](https://sigs.k8s.io/kind) to get a local cluster for testing, or run against a remote cluster.
**Note:** Your controller will automatically use the current context in your kubeconfig file (i.e. whatever cluster `kubectl cluster-info` shows).

### Running on the cluster
1. Install Instances of Custom Resources:

```sh
kubectl apply -f config/samples/
```

2. Build and push your image to the location specified by `IMG`:
	
```sh
make docker-build docker-push IMG=<some-registry>/kubebuilder:tag
```
	
3. Deploy the controller to the cluster with the image specified by `IMG`:

```sh
make deploy IMG=<some-registry>/kubebuilder:tag
```

### Uninstall CRDs
To delete the CRDs from the cluster:

```sh
make uninstall
```

### Undeploy controller
UnDeploy the controller to the cluster:

```sh
make undeploy
```

## Contributing
// TODO(user): Add detailed information on how you would like others to contribute to this project

### How it works
This project aims to follow the Kubernetes [Operator pattern](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/)

It uses [Controllers](https://kubernetes.io/docs/concepts/architecture/controller/) 
which provides a reconcile function responsible for synchronizing resources untile the desired state is reached on the cluster 

### Test It Out
1. Install the CRDs into the cluster:

```sh
make install
```

2. Run your controller (this will run in the foreground, so switch to a new terminal if you want to leave it running):

```sh
make run
```

**NOTE:** You can also run this in one step by running: `make install run`

### Modifying the API definitions
If you are editing the API definitions, generate the manifests such as CRs or CRDs using:

```sh
make manifests
```

**NOTE:** Run `make --help` for more information on all potential `make` targets

More information can be found via the [Kubebuilder Documentation](https://book.kubebuilder.io/introduction.html)

## License

Copyright 2022 huo.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

