# Overview

This project contains a proof of concept ingress controller for Kubernetes.  There are a few things that are done
differently for this ingress than your typical
[Kubernetes Ingress controller](http://kubernetes.io/v1.1/docs/user-guide/ingress.html):

* This version does pod-level routing instead of service-level routing
* This version does not use the
[Kubernetes Ingress Resource](http://kubernetes.io/v1.1/docs/user-guide/ingress.html#the-ingress-resource) and instead
uses pod-level annotations to wire things up *(This design was not my doing and was dictated by an internal design)*

The current state of this project is that this is a proof of concept driven by an internal design.  That design and
this implementation could change at any time.

# Design

This ingress controller is written in Go and upon startup, this controller will get a list of all pods across all
namespaces having the `microservice` label set to `true`.  These pods are then analyzed for the necessary wiring
configuration annotations:

* `trafficHosts`: This is a space delimited array of hosts that the pod should serve traffic for *(required)*
* `publicPaths`: This is the space delimited array of public paths that the pod should serve traffic for *(optional and
defaults to `/`)*
* `publicPort`: This is the container port that the host+path combination(s) should route to on the pod *(optional and
defaults to `80`)*

Once we've found all pods that are properly configured as microservices, we generate an nginx configuration file.

This initial list of pods is then cached and from this point forward we listen for pod events and alter our internal
cache accordingly based on the pod event.  *(The idea here was to allow for an initial hit to pull all pods but to then
use the events for as quick a turnaround as possible.)*  There is a rate limiter in place that only processes events
every 10 seconds.

# Building and Running

If you're testing this outside of Kubernetes, you can just use `go build` followed by
`KUBE_HOST=... NGINX_CONF=... ./k8s-pods-ingress`.  If you're building this to run on Kubernetes, you'll need to do the
following:

* `CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-w' -o k8s-pods-ingress .`
* `docker build ...`
* `docker tag ...`
* `docker push ...`

*(The `...` are there because your Docker comands will likely be different than mine or someone else's)*  We have an
example `rc.yaml` for deploying the k8s-pods-ingress to Kubernetes.  Here is how I test locally:

* `CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-w' -o k8s-pods-ingress .`
* `docker build -t k8s-pods-ingress .`
* `docker tag -f k8s-pods-ingress 192.168.64.1:5000/k8s-pods-ingress`
* `docker push 192.168.64.1:5000/k8s-pods-ingress`
* `kubectl create -f rc.yaml`

**Note:** This ingress is written to be ran within Kubernetes but for testing purpsoes, it can be ran outside of
Kubernetes.  When ran outside of Kubernetes, you will have to set the `KUBE_HOST` and `NGINX_CONF` environment variables
but nginx itself will not be started.  This might change in the future but for now, this support is only as a
convenience.

# Credit

This project was largely based after the `nginx-alpha` example in the
[kubernetes/contrib](https://github.com/kubernetes/contrib/tree/master/ingress/controllers/nginx-alpha) repository.

# Example

Below is an example replication that when deployed would be picked up by this ingress controller:

``` yaml
# kubectl run my-nginx --image=nginx --replicas=1 --port=80 --labels="microservice=true"
apiVersion: v1
kind: ReplicationController
metadata:
  name: my-nginx
  labels:
    app: my-nginx
    microservice: "true"
spec:
  replicas: 1
  selector:
    app: my-nginx
  template:
    metadata:
      annotations:
        trafficHosts: "testing.local test.local"
      labels:
        app: my-nginx
        microservice: "true"
    spec:
      containers:
      - image: nginx
        name: my-nginx
        ports:
        - containerPort: 80
```