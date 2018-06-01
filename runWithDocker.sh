EXEC=$1
ARGS=$2

docker build -t goapp --build-arg exec_name=$EXEC .
docker run --net="host" --rm -d --name goapp goapp $ARGS
echo "Container runs in background with name goapp"