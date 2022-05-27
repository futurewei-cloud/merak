FROM yanmo96/aca_build_standard
COPY src/interface_create.py /interface_create.py
RUN pip3 install requests
RUN apt install -y vim
