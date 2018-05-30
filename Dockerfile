FROM ubuntu

ADD main /home/

ENTRYPOINT [ "/home/main" ]
