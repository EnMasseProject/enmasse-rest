FROM fedora:25

RUN dnf -y install golang qpid-proton-c-devel && dnf clean all -y

ADD cmd/enmasse-rest-server/enmasse-rest-server /server

EXPOSE 8080

CMD ["/server", "--port", "8080", "--host", "0.0.0.0" ]
