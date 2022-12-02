FROM prom/prometheus
ADD services/merak-agent/build/merak-agent/prometheus.yml /etc/prometheus/
