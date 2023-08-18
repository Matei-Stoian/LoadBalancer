# Load Balancer
This is a simple Load Balancer writen in Golang. The main algorithm used is Round Robin.

## How to use 
To use the load balancer you will need to get the docker file and buld an image using 
```
    docker build -t your-image-name:your-image-tag .
```
Then in your docker-compose file, you will add the following
```
    services:
        load-balancer:
            image: your-image-name:your-image-tag
            environment:
                - BACKEND_SERVERS=http://backend1:8080,http://backend2:8080
            ports:
                - "8080:8080" 
```
To add or remove backend servers modivie the environment variable BACKEND_SERVERS. **The backend server must be specified with the open port**.