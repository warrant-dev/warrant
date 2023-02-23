FROM alpine:3.17.2

RUN addgroup -S warrant && adduser -S warrant -G warrant
USER warrant

WORKDIR ./
COPY ./warrant ./

ENTRYPOINT ["./warrant"]

EXPOSE 8000
