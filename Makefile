all: deps prometheus-puppet-exporter

deps:
	go get ./cmd/prometheus-puppet-exporter

prometheus-puppet-exporter:
	go build -o ./bin/prometheus-puppet-exporter ./cmd/prometheus-puppet-exporter

docker:
	docker build -t zerodotfive/prometheus-puppet-exporter .

docker-run: docker
	docker run -v /proc:/proc -v /var/cache/puppet/public/last_run_summary.yaml:/last_run_summary.yaml --network=host -it zerodotfive/prometheus-puppet-exporter

docker-push: docker
	docker push zerodotfive/prometheus-puppet-exporter

clean:
	rm -f ./bin/prometheus-puppet-exporter
