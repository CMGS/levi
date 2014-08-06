FROM flynn/busybox
MAINTAINER CMGS <ilskdw@gmail.com>

ADD ./levi /bin/levi
ADD ./etc/site.tmpl /etc/site.tmpl

ENTRYPOINT ["/bin/levi"]
CMD []

