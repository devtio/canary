FROM centos:centos7

LABEL maintainer="wilson.devito@gmail.com"

ENV CANARY_HOME=/opt/devtio \
    PATH=$CANARY_HOME:$PATH

COPY canary $CANARY_HOME/

EXPOSE 8080

ENTRYPOINT ["/opt/devtio/canary"]

