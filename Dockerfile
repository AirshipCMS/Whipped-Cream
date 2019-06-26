FROM golang as build

WORKDIR /build
COPY ./ /build
RUN make certs build-linux64

FROM scratch

COPY ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

COPY --from=build /build/bin/Whipped-Cream /
COPY --from=build /build/certs/ /certs/

ENV CERT_PATH=/certs/wc.crt \
    CERT_KEY_PATH=/certs/wc.key \
    HTTP_PORT=8080 \
    HTTPS_PORT=4433

EXPOSE 8080 4433
VOLUME /data

CMD ["/Whipped-Cream"]
