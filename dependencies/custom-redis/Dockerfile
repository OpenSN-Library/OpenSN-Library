FROM redis
COPY redis.conf /usr/local/etc/redis/redis.conf
# RUN echo "vm.overcommit_memory=1" >> /etc/sysctl.conf
EXPOSE 6379
CMD [ "redis-server", "/usr/local/etc/redis/redis.conf" ]
