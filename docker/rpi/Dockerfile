FROM hypriot/rpi-golang:1.5
MAINTAINER Christian Renz <crenz@web42.com>

RUN go get -t -v github.com/crenz/mqttrules

VOLUME /var/lib/mqttrules

COPY mqttrules.json /var/lib/mqttrules/

CMD ["/gopath1.5/bin/mqttrules", "-config /var/lib/mqttrules/mqttrules.json"]

