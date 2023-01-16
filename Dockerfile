# build stage
FROM quay.io/fedora/fedora:37-x86_64 AS build-env
RUN dnf install -y golang
ADD . /src
RUN cd /src && go build -o todohub

# final stage
FROM quay.io/fedora/fedora:37-x86_64
WORKDIR /app
COPY --from=build-env /src/todohub /app/
ENTRYPOINT ./todohub
