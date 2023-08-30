# Load Balancer
This is a simple Load Balancer writen in Golang. The algorithm I used for this implementation is Round Robin.

## How to use 
Modify your docker-compose file to include the service of load-balancer like:
```
    services:
        load-balancer:
            build:
                context: .
                dockerfile: Dockerfile
            environment:
                - BACKEND_SERVERS=http://backend1:8080,http://backend2:8080
            ports:
                - "8080:8080" 
```
To run use commands:
```
    docker-compose build
    docker-compuse up
```
To add or remove backend servers modifie the environment variable BACKEND_SERVERS. **The backend server must be specified with the open port**.