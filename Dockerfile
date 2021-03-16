FROM alpine:3.12.1 AS compiler

WORKDIR /build

RUN apk add --no-cache go
COPY . .
RUN go get && go build -v -o aws-env

FROM alpine:3.12.1
COPY --from=compiler /build/aws-env /bin
RUN ["chmod", "+x", "/bin/aws-env"]
CMD sh