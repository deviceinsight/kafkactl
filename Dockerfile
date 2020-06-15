FROM ubuntu:latest
ENV USER docker
ENV BROKER localhost:9092
COPY kafkactl /
ENTRYPOINT ["/kafkactl"]
