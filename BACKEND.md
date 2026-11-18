# Running the backend

The backend runs in Docker Compose with PostgreSQL. From the repository root, create the local secrets and environment files before starting it:

```sh
cp .env.example .env

# see https://stackoverflow.com/questions/44376846/creating-a-password-in-bash
(tr -dc 'A-Za-z0-9=' < /dev/urandom | head -c 20) > postgres_pw.txt
(tr -dc 'A-Za-z0-9=' < /dev/urandom | head -c 100) > encryption_key.txt
```

Keep `.env`, `postgres_pw.txt`, and `encryption_key.txt` local; they are ignored by Git. Use strong unique values for real deployments.

## Development

Start the backend, database, and development nginx:

```sh
docker compose --profile dev up -d
```

The backend is available on port `8080`, and the dev nginx serves on port `80`. Initialize the database schema once:

```sh
docker compose --profile init run --rm init_db
```

Stop the development services with:

```sh
docker compose --profile dev down
```

## Production

Set `CERTBOT_EMAIL` in `.env`, then start the production profile:

```sh
docker compose --profile prod up -d
```

The production nginx serves HTTP and HTTPS on ports `80` and `443`. Certbot obtains a Let's Encrypt certificate for `webdrones.net` and checks for renewal periodically. The domain must resolve to this host, and both ports must be reachable from the internet. Initialize the schema with the `init` command above if this is a new database.

Stop production services with:

```sh
docker compose --profile prod down
```

## Run the Go server directly

The server reads database connection settings from environment variables and reads secrets from the files named by `POSTGRES_PASSWORD_FILE` and `SESSION_KEY_FILE`. When running outside Compose, set `DB_HOSTNAME`, `POSTGRES_USER`, `POSTGRES_DB`, `POSTGRES_PASSWORD_FILE`, and `SESSION_KEY_FILE` for your environment. The database must be reachable at the configured hostname and port `5432`.
