FROM scratch
ADD /gofwd /
COPY ssl/ /etc/ssl/
ENTRYPOINT ["/gofwd"]
