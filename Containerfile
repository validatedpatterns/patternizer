FROM registry.access.redhat.com/ubi10:10.0

RUN dnf install -y git && dnf clean all

WORKDIR /wd

RUN git clone https://github.com/validatedpatterns/common.git

COPY entrypoint.sh .
COPY src/patternizer .
COPY Makefile .
COPY values-secret.yaml.template .

WORKDIR /repo

CMD ["/wd/entrypoint.sh"]
