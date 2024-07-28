# check for vault secret files before running
APP_CONFIG=/vault/secrets/config
DATABASE_CONFIG=/vault/secrets/database
if [ -f "$APP_CONFIG" ];
then
    source "$APP_CONFIG"
fi

if [ -f "$DATABASE_CONFIG" ];
then
    source "$DATABASE_CONFIG"
fi

# running actual command
exec ./app "$@"
