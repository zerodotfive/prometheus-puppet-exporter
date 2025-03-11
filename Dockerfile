FROM golang:1.23-alpine AS build

ADD . /build
WORKDIR /build
RUN apk add make ;\
    make

FROM alpine
COPY --from=build /build/bin/prometheus-puppet-exporter /prometheus-puppet-exporter
EXPOSE 9140
ENTRYPOINT ["/prometheus-puppet-exporter", "--summary-file"]
CMD ["/last_run_summary.yaml"]
