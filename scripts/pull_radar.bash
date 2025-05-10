#!/usr/bin/env bash
set -euf -o pipefail

# from https://stackoverflow.com/questions/59895/how-do-i-get-the-directory-where-a-bash-script-is-located-from-within-the-script
SCRIPT_DIR=$(cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd)
SECRET_DIR="$(dirname $SCRIPT_DIR)/secret"
mkdir -p $SECRET_DIR
RADAR_DIR="$(dirname $SCRIPT_DIR)/static/radar"
mkdir -p $RADAR_DIR
# Only run if we have an API key
if [ ! -f $SECRET_DIR/github ] ; then 
    printf "$SECRET_DIR/github api key does not exist. Exiting.\n" 1>&2
    exit 1
fi
GITHUB_KEY="$(cat $SECRET_DIR/github)"

OWNER='ryan-michael-19'
REPO='web-drones'
GITHUB_URL=https://api.github.com/repos/$OWNER/$REPO/actions/artifacts

# TODO: If you automate this, run curl with -s
LATEST_RADAR_BUILD_URL=$(curl \
  -H "Accept: application/vnd.github+json" \
  -H "Authorization: Bearer $GITHUB_KEY" \
  -H "X-GitHub-Api-Version: 2022-11-28" \
  $GITHUB_URL | \
    jq '.artifacts | map(select(.name == "webdrones-radar")) | sort_by(.created_at) | reverse | .[0].archive_download_url' -r)


echo $LATEST_RADAR_BUILD_URL

curl -L \
  -H "Accept: application/vnd.github+json" \
  -H "Authorization: Bearer $GITHUB_KEY" \
  -H "X-GitHub-Api-Version: 2022-11-28" \
  "$LATEST_RADAR_BUILD_URL" \
  --output $SCRIPT_DIR/radar.zip

# unzip doesn't take pipes. ugh.
unzip -o $SCRIPT_DIR/radar.zip -d $RADAR_DIR
rm $SCRIPT_DIR/radar.zip