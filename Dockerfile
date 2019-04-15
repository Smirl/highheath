FROM nginx:1.15.11
ADD https://github.com/gohugoio/hugo/releases/download/v0.46/hugo_0.46_Linux-64bit.deb /tmp/hugo.deb
COPY . /opt/code
RUN dpkg -i /tmp/hugo.deb \
    && rm /tmp/hugo.deb \
    && hugo -s /opt/code -d /usr/share/nginx/html  --cleanDestinationDir
