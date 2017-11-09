FROM scratch
COPY ./sync-logs-from-s3 /root/sync-logs-from-s3
VOLUME /root/config
CMD ["/root/sync-logs-from-s3"]
