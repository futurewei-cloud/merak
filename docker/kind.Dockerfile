FROM kindest/node:v1.25.2@sha256:f52781bc0d7a19fb6c405c2af83abfeb311f130707a0e219175677e366cc45d1
RUN apt-get update && apt-get install -y openvswitch-switch
