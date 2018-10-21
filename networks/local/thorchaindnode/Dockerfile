FROM alpine:3.7
MAINTAINER Greg Szabo <greg@tendermint.com>

RUN apk update && \
    apk upgrade && \
    apk --no-cache add curl jq file

VOLUME [ /thorchaind ]
WORKDIR /thorchaind
EXPOSE 26656 26657
ENTRYPOINT ["/usr/bin/wrapper.sh"]
CMD ["start"]
STOPSIGNAL SIGTERM

COPY wrapper.sh /usr/bin/wrapper.sh

