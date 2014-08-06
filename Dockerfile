FROM flynn/busybox
MAINTAINER CMGS <ilskdw@gmail.com>

ADD ./levi /bin/levi

ENTRYPOINT ["/bin/levi"]
CMD []

