cd "$(dirname "${BASH_SOURCE[0]}")"
if [ -z "$1" ]
  then
        echo "You must provide a binary name"
        exit 1
fi

go build -o $1
