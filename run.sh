# see https://stackoverflow.com/questions/44376846/creating-a-password-in-bash
# create postgres password
(tr -dc 'A-Za-z0-9!?%=' < /dev/urandom | head -c 20) > postgres_pw.txt
# create session encryption key
(tr -dc 'A-Za-z0-9!?%=' < /dev/urandom | head -c 100) > encryption_key.txt
docker compose down
docker compose up -d
