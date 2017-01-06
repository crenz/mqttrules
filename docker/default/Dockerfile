FROM golang:1.7
MAINTAINER Christian Renz <crenz@web42.com>

RUN go get -t -v github.com/crenz/mqttrules
RUN go install github.com/crenz/mqttrules

VOLUME /var/lib/mqttrules

CMD ["/go/bin/mqttrules"]