FROM --platform=linux/amd64 golang:1.18-alpine

WORKDIR /
RUN mkdir -p /merak-bin
COPY services/scenario-manager/build/scenario-manager services/scenario-manager/config.yaml /merak-bin/
CMD [ "/merak-bin/scenario-manager" ]
