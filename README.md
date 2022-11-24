To build using Docker compose, use

docker compose up --build -d

While using docker, change ports and file path using environment variables present in
docker-compose.yml. 

ENV for Go Server

FILE_PATH is for the file path inside the container. Can be changed to store
on the server's persistent storage by using volume mount if required.

ENV for Go Proxy

SERVER_IP is the IP address of the Go Server (in this case, internal hostname of the docker container)

SERVER_PORT is the IP address of the Go Server (in this case, internal port of the docker container)

This is assuming both containers are running on the same machine. Change external ports for outside access
in docker-compose.yml under ports section for server and proxy respectively.
