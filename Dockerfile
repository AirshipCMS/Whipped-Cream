FROM scratch

ADD ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

ENV PORT=8080 \
    MEMCACHE_URL=127.0.0.1:11211 \
    TTL=3600

ADD Whipped-Cream /

EXPOSE 8080

CMD ["/Whipped-Cream"]
