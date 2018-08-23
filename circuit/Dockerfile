FROM mysql/mysql-server:5.7

LABEL Author="Matt Parks"
LABEL Description="My First Dockerfile"

ENV MYSQL_ROOT_PASSWORD=root
ENV MYSQL_ROOT_HOST=%
ENV MYSQL_LOG_CONSOLE=true
EXPOSE 3306

COPY startup.sh .
COPY createdb.sql /docker-entrypoint-initdb.d/

CMD ["./startup.sh"]
