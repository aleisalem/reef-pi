FROM debian:stretch
LABEL maintainers="MASMART Developer"
LABEL description="MASMART Monolith Reef-Pi"


RUN mkdir -p /home/ubuntu

#RUN useradd -rm -d /home/ubuntu -s /bin/bash -g root -G sudo -u 1001 ubuntu
ADD ui /home/ubuntu/ui
COPY reef-pi /home/ubuntu/reef-pi
RUN chmod +x /home/ubuntu/reef-pi
#USER ubuntu
WORKDIR /home/ubuntu

# Copy from build directory

# Define default command
ENTRYPOINT ["/home/ubuntu/reef-pi"]

# Run the executable
#CMD ["./reef-pi"]