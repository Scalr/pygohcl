FROM quay.io/pypa/manylinux1_x86_64:latest

ARG golang_version=1.14

RUN cd /opt && curl https://storage.googleapis.com/golang/go${golang_version}.linux-amd64.tar.gz --silent --location | tar -xz

COPY build-entry.sh /

CMD /build-entry.sh
