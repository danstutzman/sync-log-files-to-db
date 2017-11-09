FROM scratch
COPY ./sync-log-files-to-db /root/sync-log-files-to-db
VOLUME /root/config
CMD ["/root/sync-log-files-to-db"]
