# build stage
FROM registry.fedoraproject.org/fedora:31 AS build-env
RUN dnf install -y golang
ADD . /src
RUN cd /src && go build -o trellohub

# final stage
FROM registry.fedoraproject.org/fedora-minimal:31
WORKDIR /app
COPY --from=build-env /src/trellohub /app/
ENTRYPOINT ./trellohub
