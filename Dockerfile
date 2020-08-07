# build stage
FROM quay.io/fedora/fedora:32-x86_64 AS build-env
RUN dnf install -y golang
ADD . /src
RUN cd /src && go build -o trellohub

# final stage
FROM registry.fedoraproject.org/fedora-minimal:31
WORKDIR /app
COPY --from=build-env /src/trellohub /app/
ENTRYPOINT ./trellohub
