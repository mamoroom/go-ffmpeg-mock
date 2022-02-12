FROM golang:1.16
RUN apt update
RUN apt upgrade
RUN apt install -y ffmpeg
RUN go get github.com/cosmtrek/air
EXPOSE 5000
CMD ["/go/bin/air"]