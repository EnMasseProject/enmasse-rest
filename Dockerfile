FROM golang

ADD cmd/enmasse-rest-server/enmasse-rest-server /server

EXPOSE 8080

CMD ["/server", "--port", "8080", "--host", "0.0.0.0" ]
