# memhashd - The distributed key-value storage


## Installation

The installation require the following tools:

- Go compiler (1.7 version or higher)

- Docker (optional, 17.05 or higher)

- docker-compose (optional, 1.14.0 version or higher)

### Build binaries

To compile project, it should be store under ```$GOPATH/src/memhashd```
directory. Then to compile the binary, execute the following command:
```sh
% go build -o /usr/bin/memhashd memhashd
```

The binary accepts the following parameters:

- ```-server-addr``` is an address used for a peer-to-peer communication.

- ```-client-addr``` is an address used to accept client requests.

- ```-join``` is an address of a node to join to the cluster.

- ```-join-retries``` a number of attempts used to join to the cluster.

### Start a cluster
In order to start a cluster of three nodes, execute the following command:
```sh
% docker-compose up -d
```
```sh
docker ps -a
CONTAINER ID        IMAGE               COMMAND                  CREATED             STATUS                     PORTS                                            NAMES
1820db0ad149        memhashd_node3      "memhashd -server-..."   7 seconds ago       Up 4 seconds               0.0.0.0:2373->2373/tcp, 0.0.0.0:8003->8003/tcp   memhashd_node3_1
5109011b9787        memhashd_node1      "memhashd -server-..."   7 seconds ago       Up 4 seconds               0.0.0.0:2371->2371/tcp, 0.0.0.0:8001->8001/tcp   memhashd_node1_1
a891ed0058f3        memhashd_node2      "memhashd -server-..."   7 seconds ago       Up 4 seconds               0.0.0.0:2372->2372/tcp, 0.0.0.0:8002->8002/tcp   memhashd_node2_1
```

After that a cluster is available on one of the three endpoints:
- 127.0.0.1:8001
- 127.0.0.1:8002
- 127.0.0.1:8003

## Usage

### List of keys

The following call returns a list of keys stored on the server.
```sh
% curl -i http://127.0.0.1:8001/v1/keys
```
```http
HTTP/1.1 200 OK
Content-Type: application/json
Date: Wed, 19 Jul 2017 11:00:40 GMT
Content-Length: 14

["1","2","3"]
```

In cluster mode, the list contains only the keys from the target node, it does
not aggregate the keys from the whole cluster.

### Load keys

The following command load a key from the store:
```sh
% curl -i http://127.0.0.1:8001/v1/keys/1
```

```http
HTTP/1.1 200 OK
Content-Type: application/json
Date: Wed, 19 Jul 2017 11:08:29 GMT
Content-Length: 301

{
  "action": "load",
  "meta": {
    "index": 1,
    "expire_time": "10s",
    "accessed_at": "2017-07-19T14:08:29.134627167+03:00",
    "created_at": "2017-07-19T14:08:27.256200005+03:00",
    "updated_at": "2017-07-19T14:08:27.256200121+03:00"
  },
  "data": [
    "a"
  ],
  "node": {
    "id": "5c3cb886-8609-4851-ac50-f8c04d2fee65",
    "addr": "127.0.0.1:2373"
  }
}
```

### Store key

The following command stores a key into a store with a timeout in 10 seconds.
Than means, a key will be automatically purged from the store after the 10
seconds:
```sh
% curl -iX PUT http://127.0.0.1:8001/v1/keys/1 \
    -H 'Content-Type: application/json' \
    -d '{"data": ["a"], "expire_time": "10s"}'
```

```http
HTTP/1.1 200 OK
Content-Type: application/json
Date: Wed, 19 Jul 2017 11:04:24 GMT
Content-Length: 287

{
  "action": "store",
  "meta": {
    "index": 2,
    "expire_time": "10s",
    "accessed_at": "0001-01-01T00:00:00Z",
    "created_at": "2017-07-19T14:00:24.627500435+03:00",
    "updated_at": "2017-07-19T14:04:24.264979799+03:00"
  },
  "data": [
    "a"
  ],
  "node": {
    "id": "5c3cb886-8609-4851-ac50-f8c04d2fee65",
    "addr": "127.0.0.1:2373"
  }
}
```

### Delete key

The following commands removes the key from the store:
```sh
% curl -iX DELETE http://127.0.0.1:8001/v1/keys/1
```
```http
HTTP/1.1 200 OK
Content-Type: application/json
Date: Wed, 19 Jul 2017 11:23:41 GMT
Content-Length: 244

{
  "action": "delete",
  "meta": {
    "index": 0,
    "expire_time": "0s",
    "accessed_at": "0001-01-01T00:00:00Z",
    "created_at": "0001-01-01T00:00:00Z",
    "updated_at": "0001-01-01T00:00:00Z"
  },
  "node": {
    "id": "5c3cb886-8609-4851-ac50-f8c04d2fee65",
    "addr": "127.0.0.1:2373"
  }
}
```

### List index

The following command loads the data at the given position in a list (this
operation is supported only for list types):

```sh
% curl -X PUT http://127.0.0.1:8001/v1/keys/1 \
    -H 'Content-Type: application/json' \
    -d '{"data": ["a", "b", "c"]}'
%
% curl http://127.0.0.1:8001/v1/keys/1/index \
    -H 'Content-Type: application/json' \
    -d '{"index": 2}'
```
```http
HTTP/1.1 200 OK
Content-Type: application/json
Date: Wed, 19 Jul 2017 11:29:41 GMT
Content-Length: 299

{
  "action": "index",
  "meta": {
    "index": 2,
    "expire_time": "0s",
    "accessed_at": "2017-07-19T14:29:41.385449109+03:00",
    "created_at": "2017-07-19T14:26:45.403289817+03:00",
    "updated_at": "2017-07-19T14:28:52.802453851+03:00"
  },
  "data": "c",
  "node": {
    "id": "5c3cb886-8609-4851-ac50-f8c04d2fee65",
    "addr": "127.0.0.1:2373"
  }
}
```

### Dict index

The following command loads the data at the given item in a dict (this
operation is supported only for dictionary types):
```sh
% curl http://127.0.0.1:8001/v1/keys/1/item \
    -H 'Content-Type: application/json' \
    -d '{"item": "user"}'
```
```http
HTTP/1.1 200 OK
Content-Type: application/json
Date: Wed, 19 Jul 2017 11:27:19 GMT
Content-Length: 304

{
  "action": "item",
  "meta": {
    "index": 1,
    "expire_time": "0s",
    "accessed_at": "2017-07-19T14:27:19.386854312+03:00",
    "created_at": "2017-07-19T14:26:45.403289817+03:00",
    "updated_at": "2017-07-19T14:26:45.403289918+03:00"
  },
  "data": "ybubnov",
  "node": {
    "id": "5c3cb886-8609-4851-ac50-f8c04d2fee65",
    "addr": "127.0.0.1:2373"
  }
}
```


## License

The memhashd is distributed under MIT license, therefore you are free to do
with code whatever you want. See the [LICENSE](LICENSE) file for full license
text.
