FROM scratch

WORKDIR /app

COPY ./bin/server /app/server
# RUN chmod +x /app/server


CMD ["/app/server"]
 
