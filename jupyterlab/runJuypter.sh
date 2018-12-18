PATH=$1
docker build -t jupyter-notebook .
docker run -p 8888:8888 --rm -e JUPYTER_LAB_ENABLE=yes -v $PATH:/home/jovyan/work jupyter-notebook 
