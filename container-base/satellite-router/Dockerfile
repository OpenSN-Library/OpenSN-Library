FROM build-base

COPY materials/daemons /etc/frr/daemons
COPY scripts/entrypoint /Workspace/entrypoint
RUN chmod +x /Workspace/entrypoint
CMD [ "/Workspace/entrypoint"]
