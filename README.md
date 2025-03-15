# Web Drones!

<p align="center"><img src="https://github.com/user-attachments/assets/46ddb4ce-667f-4673-a869-3014de5a15e4" width="250"/></p>

## Welcome!

Web Drones is a game about exploring the world with drones to create more drones. Web drones are controled via a REST api, allowing players to learn how to make HTTP requests and write code while playing the game.

I am currently hosting the game at https://webdrones.net, and it can be ran in containers with the provided `docker-compose.yaml`.

API documentation at https://ryan-michael-19.github.io/web-drones/

The current rate limit is set to 2 requests per second.

All users refresh every Sunday at 8AM!

Contact ryan-michael-tech@gmail.com for support.

## Tutorial

Let's play Web Drones using cURL. The first thing we need to do is create a user and store the cookie we will use to access the game.

```shell
curl -X POST -c cookie-jar  -u ${Username}:${Password} https://webdrones.net/newUser | jq
```

Along with the cookie cURL adds to ``cookie-jar``, You will recieve json with information about your bots and the scrap metal mines in the area. 

```json
{
  "bots": [
    {
      "coordinates": {
        "x": -5,
        "y": -5
      },
      "identifier": "37dcdfe2-03bd-4aae-a6f3-8d0f35062d74",
      "inventory": 0,
      "name": "Gretchen",
      "status": "IDLE"
    },
    .
    .
    .
  ],
  "mines": [
    {
      "x": 5.059399139783757,
      "y": 43.97327960824991
    },
    .
    .
    .
  ]
}

```
We can use Gretchen's identifier to move them to a scrap mine.

```shell
curl -X POST -b cookie-jar \
    --header "Content-Type: application/json" \
    -d '{"x": 5.059399139783757, "y": 43.97327960824991}' \
    https://webdrones.net/bots/37dcdfe2-03bd-4aae-a6f3-8d0f35062d74/move | jq
```

We'll get the following response which shows Gretchen moving towards the scrap mine.

```json
{
  "coordinates": {
    "x": -4.9985212553898535,
    "y": -4.992800864916925
  },
  "identifier": "37dcdfe2-03bd-4aae-a6f3-8d0f35062d74",
  "inventory": 0,
  "name": "Gretchen",
  "status": "MOVING"
}
```

Try moving other bots while waiting for Gretchen to reach the mine!

We can also get updates on where Gretchen is with the following request:

```shell
curl -b cookie-jar https://webdrones.net/bots/37dcdfe2-03bd-4aae-a6f3-8d0f35062d74 | jq
```

```json
{
  "coordinates": {
    "x": 5.059399139783757,
    "y": 43.97327960824991
  },
  "identifier": "37dcdfe2-03bd-4aae-a6f3-8d0f35062d74",
  "inventory": 0,
  "name": "Gretchen",
  "status": "IDLE"
}
```

Once Gretchen has reached a mine we can direct her to extract the scrap metal from it.

```shell
curl -X POST -b cookie-jar https://webdrones.net/bots/37dcdfe2-03bd-4aae-a6f3-8d0f35062d74/extract | jq
```

Notice her inventory goes up by one!

```json
{
  "coordinates": {
    "x": 5.059399139783757,
    "y": 43.97327960824991
  },
  "identifier": "37dcdfe2-03bd-4aae-a6f3-8d0f35062d74",
  "inventory": 1,
  "name": "Gretchen",
  "status": "IDLE"
}
```

Let's look for a new mine.

```shell
curl -b cookie-jar https://webdrones.net/mines | jq
```
```json
[
  {
    "x": 0.15225387779620547,
    "y": -29.790030102078568
  },
  .
  .
  .
]
```

Now we can send Getchen to the mine like before.

```shell
curl -X POST -b cookie-jar \
    --header "Content-Type: application/json" \
    -d '{"x": 0.15225387779620547, "y": -29.790030102078568}' \
    https://webdrones.net/bots/37dcdfe2-03bd-4aae-a6f3-8d0f35062d74/move | jq
```

We can also track her location like before.

```shell
curl -b cookie-jar https://webdrones.net/bots/37dcdfe2-03bd-4aae-a6f3-8d0f35062d74 | jq
```

```json
{
  "coordinates": {
    "x": 0.15225387779620547,
    "y": -29.790030102078568
  },
  "identifier": "37dcdfe2-03bd-4aae-a6f3-8d0f35062d74",
  "inventory": 1,
  "name": "Gretchen",
  "status": "IDLE"
}
```

And extract scrap like before.

```shell
curl -X POST -b cookie-jar https://webdrones.net/bots/37dcdfe2-03bd-4aae-a6f3-8d0f35062d74/extract | jq
```

```json
{
  "coordinates": {
    "x": 0.15225387779620547,
    "y": -29.790030102078568
  },
  "identifier": "37dcdfe2-03bd-4aae-a6f3-8d0f35062d74",
  "inventory": 2,
  "name": "Gretchen",
  "status": "IDLE"
}
```

Repeat this process a third time using the previous commands so Gretchen has an inventory of three.

From there, we can create a new bot!

```shell
curl -X POST -b cookie-jar \
    --header "Content-Type: application/json" \
    -d '{"NewBotName": "Samuel"}' \
    https://webdrones.net/bots/37dcdfe2-03bd-4aae-a6f3-8d0f35062d74/newBot | jq
```

```json
{
  "coordinates": {
    "x": 6.8445106156946025,
    "y": -28.323905093494762
  },
  "identifier": "1b2c4e9d-98b4-480f-b788-49fa041d7f45",
  "inventory": 0,
  "name": "Samuel",
  "status": "IDLE"
}
```

Can you think of ways to automate bots moving and making more bots?