FROM golang:1.23-alpine3.20 AS BuildStage
ENV GO111MODULE="on"
# This links go binaries with go builds instead of C builds
# Keeps image smaller than using a base image like ubuntu
# but also has a performance hit
ENV CGO_ENABLED="0"
WORKDIR /
COPY src/ / 
RUN apk add git && go get && go build -o /web-drones

FROM golang:1.23-alpine3.20
WORKDIR /app
RUN ["mkdir", "/app/schemas"]
COPY --from=BuildStage /web-drones /app/
COPY ./src/schemas/schemas.sql /app/schemas
CMD ["/app/web-drones", "SERVER"]

