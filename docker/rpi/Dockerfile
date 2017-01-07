FROM hypriot/rpi-golang:1.5
MAINTAINER Christian Renz <crenz@web42.com>

RUN go get -t -v github.com/crenz/mqttrules/cmd/mqttrules

VOLUME /var/lib/mqttrules

CMD ["/go/bin/mqttrules"]
