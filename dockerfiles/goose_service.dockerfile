FROM gomicro/goose

ADD migrations/*.sql /migrations/
ADD dockerfiles/inv/goose/entrypoint.sh /migrations/

ENTRYPOINT [ "bash", "entrypoint.sh" ]