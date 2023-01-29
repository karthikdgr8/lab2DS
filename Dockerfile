# syntax=docker/dockerfile:1

FROM golang:1.19.5-alpine

WORKDIR /app

ARG PORT=8081

COPY /src ./src
COPY /files ./files

COPY go.mod ./
RUN go mod download github.com/holiman/uint256

WORKDIR /app/src/main
RUN go build -o /peer

EXPOSE $PORT

CMD exec /peer -a $IP -p $PORT --ja $JOIN_IP --jp $JOIN_PORT --ts $STABILIZE_TIME -tff $FIX_FINGER_TIME -tcp $CHECK_PRED_TIME -i $UNIQ_ID -r $SUCC_NO
