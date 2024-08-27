# Vas3k Atomizer

- [Vas3k Atomizer](#vas3k-atomizer)
  - [Motivation](#motivation)
    - [Features](#features)
  - [Installation](#installation)
    - [Build from source](#build-from-source)
    - [Prebuilt binaries](#prebuilt-binaries)
    - [Docker](#docker)
  - [Configuration](#configuration)
  - [Usage](#usage)
    - [Example Requests](#example-requests)
  - [License](#license)
  - [Contributing](#contributing)
  - [Acknowledgements](#acknowledgements)

## Motivation 

I read a significant portion of my blogs via self-hosted miniflux. The club provides a JSON feed for subscription, but it has some subjective drawbacks:
- article content is given as markdown, which miniflux cannot render
- I have not yet had a case when I wanted to block an author, but bitter experience has taught me that this feature is necessary.
- I am not interested in reading intro posts. Yes, there are feeds for each type of post, but I would like a consolidated feed

### Features

- publication type blocklist
- author blocklist
- content md->html rendering
- atom feed

## Installation

### Build from source

- Go build: `go install -trimpath github.com/ninedraft/vas3katomizer@latest`
- Cloning && building
```sh
git clone github.com/ninedraft/vas3katomizer
cd vas3katomizer
go install ./
```

### Prebuilt binaries

Go to [releases page](github.com/ninedraft/vas3katomizer/releases) and download binary for OS && ARCH combination you need.

### Docker

Use image `ghcr.io/ninedraft/vas3katomizer:latest`. It supports `linux/arm64` and `linux/amd64`.

## Configuration

Set up the required environment variables:

- `VAS3KCLUB_TOKEN`: The token used for authentication with the Vas3k Club API. **(Required)**
- `SERVE_AT`: The address where the server will be hosted. Defaults to `localhost:8390` (`0.0.0.0:8390` for docker image)
- `VAS3KCLUB_ENDPOINT`: The endpoint for fetching the feed. Defaults to `https://vas3k.club/`.
- `LOG_LEVEL`: The log level for the application. The logging level, such as `DEBUG`, `INFO`, `WARN`, `ERROR`. Defaults to `DEBUG`.
- `BLOCKED_TYPES`: list feed item types you don't want to see. Defaults to `intro`. Currently club supports `intro`, `question`, `project`, `post`. Values in the list can be separated by space characters or any of `,;|/`.
- `BLOCKED_AUTHORS`: list of authors you don't want to see. Username is the last part in author page URL (`https://vas3k.club/user/admin/` -> `admin`). Values in the list can be separated by space characters or any of `,;|/`.

## Usage

The server provides several HTTP endpoints to fetch and convert the feeds:

- **`GET /`**: Serves a basic index page.
- **`GET /feed/{format}`**: Fetches the latest feed and converts it into the specified format (`atom`, `json`, or `rss`).
- **`GET /page/{page}/{format}`**: Fetches the feed for a specific page and converts it into the specified format.

### Example Requests

- Fetch the latest feed in Atom format:
  
  ```http
  GET http://localhost:8390/feed/atom
  ```

- Fetch the second page of the feed in RSS format:
  
  ```http
  GET http://localhost:8390/page/2/rss
  ```

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

## Contributing

Contributions are welcome! Please feel free to submit a pull request or open an issue if you find a bug or have a feature request.

## Acknowledgements

- [Gorilla Feeds](https://github.com/gorilla/feeds) for feed generation.
- [Vas3k Club](https://vas3k.club/) for providing the platform that this tool interacts with.
