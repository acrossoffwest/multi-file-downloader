## Multi-files downloader

For build alpine and buster execution files run:

    docker-compose up

Execution file will appear in `./builds` directory.

    ./builds/mfd-alpine [timeout in seconds] "[Url]:[output file path];...;[Url-n]:[output file-n path]"
    ./builds/mfd-buster [timeout in seconds] "[Url]:[output file path];...;[Url-n]:[output file-n path]"