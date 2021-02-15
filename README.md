# todolist

Basic Api to store a list of tasks to do. For now it is only possible to store and retrieve list items. Update and deletion is still to be done.

### Installation

To create and start containers running the api, database and test database:
```sh
$ docker-compose up
```

To create and start containers only of the app and the database
```sh
$ docker-compose up todolist
```

To create and start containers only of the dabatase that is used by tests
```sh
$ docker-compose up testmysql
```

To clean up all the containers, volumes and created docker image
```sh
$ docker rm todolist_api todolist_mysql -f
$ docker image rm todolist_todolist
$ docker volume rm todolist_datavolume todolist_testdatavolume
```

## Usage

## Add item

```sh
$ curl -v PUT http://127.0.0.1:8000/items --data '{"title":"new title","description":"description","dueDate":"2021-03-01T15:00:00Z","comments":[{"text":"here goes some text for the comment"}],"labels":[{"text":"here a label"}]}'
```
request object:
```sh
{
  "title": "new title",
  "description": "description",
  "dueDate": "2021-03-01T15:00:00Z",
  "comments": [
    {"text": "here goes some text for the comment"}
  ],
  "labels": [
    {"text": "here a label"}
  ]
}
```
returns http status 201 Created and an id:
```sh
{"id":1}
```

## Get item from the based on id 
```sh
$ curl -v GET http://127.0.0.1:8000/items/1
```
returns http status 200 OK (400 Not Foun if the id doesn't exist) and an json object:
```sh
{
  "id": 6,
  "title": "new title",
  "description": "description",
  "labels": [
    {"id": 1,"text": "here a label"}
  ],
  "comments": [
    {"id": 1,"text": "here goes some text for the comment"}
  ],
  "status": false,
  "dueDate": "2021-03-01T15:00:00Z"
}
```

## Get all the list items
```sh
$ curl -v GET http://127.0.0.1:8000/items
```

returns http status 200 OK and a json array of objects
```sh
[
  {
    "id": 1,
    "title": "new title",
    "description": "description",
    "labels": [{"id": 1,"text": "here a label"}],
    "comments": [{"id": 1,"text": "here goes some text for the comment"}],
    "status": false,
    "dueDate": "2021-03-01T15:00:00Z"
  }
]
```

