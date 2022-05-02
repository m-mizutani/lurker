FROM golang:1.18.1

RUN apt-get update && apt-get install -y libpcap0.8-dev
COPY . /src
WORKDIR /src
RUN go build -o /lurker .

# FROM gcr.io/distroless/base
# COPY --from=build-go /src/lurker /lurker

WORKDIR /
ENTRYPOINT ["/lurker"]
