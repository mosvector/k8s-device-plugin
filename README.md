# k8s-device-plugin

A proof-of-concept Kubernetes device plugin for testing mosvector/linux-char-device-driver

## TL;DR

```shell
docker build -t github.com/mosvector/k8s-device-plugin:latest .
kind create cluster
kind load docker-images github.com/mosvector/k8s-device-plugin:latest
kubectl apply -f kube/daemonset.yml
kubectl apply -f kube/test-pod.yml
```

## Debugging

```console
$ kubectl get pod -n kube-system -l app=mosvector-device-plugin
NAME                            READY   STATUS    RESTARTS   AGE
mosvector-device-plugin-5ppvf   1/1     Running   0          38m
mosvector-device-plugin-v69v2   1/1     Running   0          38m

$ kubectl logs -n kube-system $(kubectl get pods -n kube-system -l app=mosvector-device-plugin -o jsonpath='{.items[?(@.spec.nodeName=="kind-worker")].metadata.name}')
2024/10/26 01:56:32 Starting MosvectorDevicePlugin...
2024/10/26 01:56:32 Socket found; server is up and running.
2024/10/26 01:56:32 Registering MosvectorDevicePlugin with Kubelet...
2024/10/26 01:56:32 Serving gRPC server for MosvectorDevicePlugin...
2024/10/26 01:56:32 GetDevicePluginOptions called.
2024/10/26 01:56:32 ListAndWatch called. Sending initial device list...
2024/10/26 01:56:32 Successfully registered MosvectorDevicePlugin with Kubelet.
2024/10/26 01:56:32 MosvectorDevicePlugin is running. Waiting indefinitely...
2024/10/26 01:56:32 Initial device list sent successfully.
2024/10/26 01:56:32 ListAndWatch: Sending updated device health status...

$ kubectl exec -it pod/test-mosvector-device -- cat /dev/mosvector
Hello, World!
```

## Dev Environment

```console
$ uname -a
Linux RaspberryPi 6.6.31+rpt-rpi-2712 #1 SMP PREEMPT Debian 1:6.6.31-1+rpt1 (2024-05-29) aarch64 GNU/Linux
$ go version
go version go1.23.2 linux/arm64
$ docker version
Client: Docker Engine - Community
 Version:           27.3.1
 API version:       1.47
 Go version:        go1.22.7
 Git commit:        ce12230
 Built:             Fri Sep 20 11:41:19 2024
 OS/Arch:           linux/arm64
 Context:           default

Server: Docker Engine - Community
 Engine:
  Version:          27.3.1
  API version:      1.47 (minimum version 1.24)
  Go version:       go1.22.7
  Git commit:       41ca978
  Built:            Fri Sep 20 11:41:19 2024
  OS/Arch:          linux/arm64
  Experimental:     false
 containerd:
  Version:          1.7.22
  GitCommit:        7f7fdf5fed64eb6a7caf99b3e12efcf9d60e311c
 runc:
  Version:          1.1.14
  GitCommit:        v1.1.14-0-g2c9f560
 docker-init:
  Version:          0.19.0
  GitCommit:        de40ad0
$ kind version
kind v0.24.0 go1.22.6 linux/arm64
$ kubectl get nodes
NAME                 STATUS   ROLES           AGE   VERSION
kind-control-plane   Ready    control-plane   31d   v1.31.0
kind-worker          Ready    <none>          31d   v1.31.0
kind-worker2         Ready    <none>          31d   v1.31.0
```
