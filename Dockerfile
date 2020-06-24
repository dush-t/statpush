FROM p4lang/p4app:latest

WORKDIR /

# Create a directory 'monitor' and dump all the switch code and driver
# code in it
RUN mkdir monitor
WORKDIR /monitor
ADD switch .
ADD run_network .

# ENTRYPOINT ["tail", "-f", "/dev/null"]
ENTRYPOINT [ "python",  "/monitor/run.py" ]
