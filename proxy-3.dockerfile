# Use the Ubuntu base image
FROM ubuntu

# Update package repositories and install necessary packages (etcd and curl)
RUN apt-get update && apt-get install -y etcd curl

# Set the working directory inside the container to /app
WORKDIR /app

# Create a directory named /app in the container filesystem
RUN mkdir -p /app

# Copy the script named start-etcd.sh from the host machine to the container's /app directory
COPY start-etcd-3.sh /app/start-etcd-3.sh

# Give execute permission to the start-etcd.sh script to allow it to be run
RUN chmod +x /app/start-etcd-3.sh

# Copy the executable file named 'main' from the host machine to the current working directory in the container
COPY main .

# Expose ports 2379 and 2380 on the container, allowing communication through these ports
EXPOSE 2379 2380

# Set the entry point command to run the start-etcd.sh script when the container starts
ENTRYPOINT ["/app/start-etcd-2.sh"]
