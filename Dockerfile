FROM        quay.io/prometheus/busybox:latest
MAINTAINER  The Prometheus Authors <prometheus-developers@googlegroups.com>

COPY flexswitch_exporter /bin/flexswitch_exporter

EXPOSE      9100
ENTRYPOINT  [ "/bin/flexswitch_exporter" ]
