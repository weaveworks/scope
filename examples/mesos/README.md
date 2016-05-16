# Launching Weave Scope on Apache Mesos Cluster

## Launching on minimesos Cluster

`minimesos` project enables running Apache Mesos cluster on a single machine for prototyping and unit testing of Mesos 
frameworks. All running components of `minimesos` cluster are represented by docker containers. To enable reuse of local 
docker images, docker in Mesos Agent containers reuses `docker.sock` file of the host. Therefore all started in Agents
containers are also seen on the host. This setup gives single Weave Scope container access to all containers running in 
`minimesos` cluster. 

Launching of Weave Scope by providing content `minimesos.json` file to Marathon is a part of default `minimesos` cluster. If
removed from configuration, Weave Scope can be added to the running cluster by `minimesos install --marathonFile <URL or path to minimesos.json>` 

UI of Weave Scope can be accessed on http://172.17.0.1:4040/
