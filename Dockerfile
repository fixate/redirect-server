FROM alpine:3.2

#RUN apk update && \
  #apk add \
    #ca-certificates && \
  #rm -rf /var/cache/apk/*

COPY redirect-server /bin

ENTRYPOINT ["/bin/redirect-server"]
