FROM        golang
RUN         go get github.com/sashajeltuhin/ket/provision/exec/provision-web
RUN 	    go install github.com/sashajeltuhin/ket/provision/exec/provision-web
EXPOSE	    8013
ENTRYPOINT  $GOBIN/provision-web