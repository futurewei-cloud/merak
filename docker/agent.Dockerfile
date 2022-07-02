FROM phudtran/aca:dev

# Merak
WORKDIR /
RUN mkdir -p /merak-bin
COPY services/merak-agent/build/merak-agent /merak-bin/merak-agent
CMD [ "/merak-bin/merak-agent" ]
