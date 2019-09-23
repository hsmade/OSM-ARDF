FROM ubuntu:bionic
RUN apt update && apt install -y python-pip python-pil python-yaml libproj12
RUN pip install mapproxy
ADD init.sh /
CMD ["/init.sh"]
EXPOSE 8080/tcp
