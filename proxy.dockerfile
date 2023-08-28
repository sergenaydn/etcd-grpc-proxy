# Use the Ubuntu base image
FROM ubuntu

# Update package repositories and install necessary packages
RUN apt-get update && apt-get install -y etcd curl

# Set the working directory inside the container to /app
WORKDIR /app

# Create a directory named /app
RUN mkdir -p /app

# Copy the script named start-etcd.sh from the host to the container's /app directory
COPY start-etcd.sh /app/start-etcd.sh

# Give execute permission to the start-etcd.sh script
RUN chmod +x /app/start-etcd.sh

# Copy the executable file named 'main' from the host to the current working directory in the container
COPY main .

# Expose ports 2379 and 2380 for communication
EXPOSE 2379 2380

# Set the entry point command to run the start-etcd.sh script when the container starts
ENTRYPOINT ["/app/start-etcd.sh"]
