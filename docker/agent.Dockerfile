FROM phudtran/aca:dev

# Merak
COPY . /merak
WORKDIR /merak
RUN go mod download
RUN make merak-agent
CMD [ "/merak/services/merak-agent/build/merak-agent" ]
