FROM golang as build

WORKDIR /build
COPY ./ /build
RUN make build-linux64


FROM scratch

ADD ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

ENV HTTP_PORT=8080 \
    HTTPS_PORT=4433

COPY --from=build /build/bin/Whipped-Cream /

EXPOSE 8080
EXPOSE 4433
VOLUME /data

CMD ["/Whipped-Cream"]
