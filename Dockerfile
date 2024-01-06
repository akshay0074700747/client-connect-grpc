FROM golang:1.21.5-bullseye AS build

RUN apt-get update && apt-get install -y git

WORKDIR /app

RUN echo client-connect

RUN echo client-connect

RUN git clone https://github.com/akshay0074700747/client-connect-grpc.git .

RUN go mod download

WORKDIR /app/cmd

RUN go build -o bin/client-connect

COPY /cmd/.env /app/cmd/bin/

FROM busybox:latest

WORKDIR /client-connect

COPY --from=build /app/cmd/bin/client-connect .

COPY --from=build /app/cmd/bin/.env .

EXPOSE 50001

CMD [ "./client-connect" ]