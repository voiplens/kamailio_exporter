ARG ARCH="amd64"
ARG OS="linux"
FROM quay.io/prometheus/busybox-${OS}-${ARCH}:latest
LABEL maintainer="The Angarium dev team <exporter-developers@angarium.io>"

ARG ARCH="amd64"
ARG OS="linux"
COPY .build/${OS}-${ARCH}/kamailio_exporter /bin/systemd_exporter

EXPOSE      9558
USER        nobody
ENTRYPOINT  ["/bin/kamailio_exporter"]
