FROM golang:1.15.2-buster AS build

ADD . /opt/app
WORKDIR /opt/app
RUN go build ./cmd/main.go

FROM ubuntu:20.04 AS release

# Make the "en_US.UTF-8" locale so postgres will be utf-8 enabled by default
RUN apt -y update && apt install -y locales gnupg2
RUN locale-gen en_US.UTF-8
RUN update-locale LANG=en_US.UTF-8

#
# Install postgresql
#
ENV PGVER 12
ENV DEBIAN_FRONTEND noninteractive
RUN apt-get -y update && apt-get install -y postgresql-$PGVER

# Run the rest of the commands as the ``postgres`` user created by the ``postgres-$PGVER`` package when it was ``apt-get installed``
USER postgres

# Create a PostgreSQL role named ``docker`` with ``docker`` as the password and
# then create a database `docker` owned by the ``docker`` role.
RUN /etc/init.d/postgresql start &&\
    psql --command "CREATE USER docker WITH SUPERUSER PASSWORD 'docker';" &&\
    createdb -O docker forum &&\
    /etc/init.d/postgresql stop

# Adjust PostgreSQL configuration so that remote connections to the
# database are possible.
RUN echo "host all  all    0.0.0.0/0  md5" >> /etc/postgresql/$PGVER/main/pg_hba.conf

# And add ``listen_addresses`` to ``/etc/postgresql/$PGVER/main/postgresql.conf``
# RUN echo "listen_addresses='*'" >> /etc/postgresql/$PGVER/main/postgresql.conf
# RUN echo "synchronous_commit = off" >> /etc/postgresql/$PGVER/main/postgresql.conf
# RUN echo "fsync = off" >> /etc/postgresql/$PGVER/main/postgresql.conf
# RUN echo "shared_buffers = 1024MB" >> /etc/postgresql/$PGVER/main/postgresql.conf
# Expose the PostgreSQL port
EXPOSE 5432

# Add VOLUMEs to allow backup of config, logs and databases
VOLUME  ["/etc/postgresql", "/var/log/postgresql", "/var/lib/postgresql"]

# Back to the root user
USER root

WORKDIR /usr/src/app

COPY . .
COPY --from=build /opt/app/main .

# Объявлем порт сервера
EXPOSE 5000

#
# Запускаем PostgreSQL и сервер
#
ENV PGPASSWORD docker
CMD service postgresql start && psql -h localhost -d forum -U docker -p 5432 -a -q -f ./configs/init.sql && ./main
