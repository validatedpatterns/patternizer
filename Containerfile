FROM registry.access.redhat.com/ubi10:10.0

RUN dnf install -y git && dnf clean all

WORKDIR /wd

COPY pattern.sh .
COPY default-cmd.sh .
COPY src/patternizer .
COPY values-secret.yaml.template .

WORKDIR /repo

ENV USE_SECRETS=false

CMD ["/wd/default-cmd.sh"]
