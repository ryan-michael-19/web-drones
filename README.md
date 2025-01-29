# Web Drones!

<p align="center"><img src="https://github.com/user-attachments/assets/46ddb4ce-667f-4673-a869-3014de5a15e4" width="250"/></p>

## Welcome!

Web Drones is a game about exploring the world with drones to create more drones. Web drones are controled via a REST api, allowing players to learn how to make HTTP requests and write code while playing the game.

I am currently hosting the game at https://webdrones.net, and it can be ran in containers with the provided `docker-compose.yaml`.

API documentation at https://ryan-michael-19.github.io/web-drones/

After you create a new user, you must run the `/init` endpoint to start the game or you will get server errors accessing other endpoints (fix coming soon!)

The current rate limit is set to 2 requests per second.
