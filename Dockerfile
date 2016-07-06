FROM gcr.io/google_containers/nginx

MAINTAINER Jeremy Whitlock <jwhitlock@apache.org>

LABEL Description="A Pod-based router for Kubernetes."

COPY k8s-router /

CMD ["/k8s-router"]
