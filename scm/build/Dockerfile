# Multi purpose Dockerfile for the following usecases
# 1. Building project in centos container (libc)
# 2. Building project in alpine container (musl)
# A. Mapping project files from host
ARG BUILDBASE=golang:1.13-alpine
ARG FINALBASE=alpine:latest


# Builder base image is parameterizable
# Restriction apply:
# It should be a centos with yum package manager
# OR
# It should be an alpine with apk package manager
FROM ${BUILDBASE} as base
MAINTAINER Levente Kale <levente.kale@nokia.com>
ARG PACKAGES="bash ca-certificates curl git libcap"
ENV GOPATH=/go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH

# ATTENTION! HACK
# What this line os instruction does is:
# 1. detects alpine apk or centos yum package manager
# 2. then detect that go if is available on the $PATH
# 3. then install $PACKAGES including golang when it is not available
#
# Either way base image should have package manager apk for alpine or yum for centos
# otherwise, should fast exit with error message
RUN   apk add --no-cache ${PACKAGES} $(go version >/dev/null 2>&1 || echo go) \
 || ( yum install -y     ${PACKAGES} $(go version >/dev/null 2>&1 || echo golang) \
   && yum clean -y all \
   && rm -rf /var/cache/yum )\
 || ( echo "UNSUPPORTED BASE IMAGE! Image should derive from 'alpine' or 'centos'" \
   && exit 1)

FROM base as builder
COPY scm/build/build.sh /build.sh
ENTRYPOINT /build.sh

FROM base as project
ARG GOPROXY
WORKDIR ${GOPATH}/src/github.com/nokia/danm
#COPY scm/build/build.sh /build.sh
ENV GOOS=linux
ENV CGO_ENABLED=0
#ENV GOARGS="-mod=vendor"
ENV GO111MODULE=on
COPY .gitignore ./.gitignore
COPY hack ./hack
COPY scm ./scm
COPY crd ./crd
COPY cmd ./cmd
COPY pkg ./pkg
COPY test ./test
COPY go.* ./
RUN go mod tidy
RUN go mod vendor
#RUN chmod a+x ./vendor/k8s.io/code-generator/generate-groups.sh
# Use code-generator boilerplate.go.txt as Copyright template for generated go files
RUN find ${GOPATH} -path '*code-generator@*boilerplate.go.txt' -exec cp {} hack/ \;
#RUN go generate ./...
COPY .git ./.git

FROM project as dlv
EXPOSE 2345
ENTRYPOINT dlv --listen=:2345 --headless=true --api-version=2 --accept-multiclient exec danm

FROM project as binaries
ARG LATEST_TAG=dev
ARG COMMIT_HASH=dev
ENV LDFLAGS="-extldflags '-static' -X main.version=${LATEST_TAG} -X main.commitHash=${COMMIT_HASH}"
RUN go install -mod=vendor -v -a -ldflags "${LDFLAGS}" ./cmd/...
# TODO: investigate arping need in project and best build option go install vs go get
RUN go install -v -a -ldflags "-extldflags '-static'" ./vendor/github.com/j-keck/arping

RUN chmod 750 /go/bin/{cnitest,danm,netwatcher,svcwatcher,webhook}
RUN setcap cap_sys_ptrace,cap_sys_admin,cap_net_admin=eip /go/bin/netwatcher \
           cap_net_raw=eip /go/bin/arping # TODO: ask Levo about /usr/sbin/arping usage

# TODO: is there any need for this?
FROM ${FINALBASE} as component
ARG COMPONENT
ENV COMPONENT=${COMPONENT}
COPY --from=binaries --chown=root:danm /go/bin/${COMPONENT} /
ENTRYPOINT /${COMPONENT}

FROM ${FINALBASE} as hyperdanm
WORKDIR /usr/local/bin
RUN adduser -u 147 -D -H -s /sbin/nologin danm
# TODO: investigate arping need in project
# NOTE: here binaries have needed capabilities already set, see binaries stage
COPY --from=binaries --chown=root:danm /go/bin/netwatcher /go/bin/svcwatcher /go/bin/webhook /go/bin/arping ./
USER danm

# Unit-test stage should be last because it is time consumind
FROM project as unit-test
RUN go install ./cmd/cnitest
RUN go test -count=1 ./test/uts/...


#
# Stage: CNI plugins daemonset
#
# Note that unlike the other containers, this needs to run as root as
# it places CNI plugins into the host's filesystem.
#
FROM alpine:3.11 AS danm-cni-plugins

VOLUME ["/host/cni"]

RUN mkdir /cni
COPY --from=builder /go/bin/danm /go/bin/fakeipam /cni/

COPY scm/build/cni_ds/entrypoint.sh /entrypoint.sh
ENTRYPOINT ["/entrypoint.sh"]
