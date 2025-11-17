# rss-builder

[![Last scheduled run](https://github.com/guille/rss-builder/actions/workflows/deploy.yml/badge.svg?event=schedule)](https://github.com/guille/rss-builder/actions/workflows/deploy.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/guille/rss-builder)](https://goreportcard.com/report/github.com/guille/rss-builder)

I use RSS to keep up with sites that interest me. However, not every website provides RSS, or one that is granular enough for my interest. So I built this.

Every night, a scheduled Github Actions workflow runs main.go, which builds a RSS xml file and publishes it to [Github Pages](https://guille.github.io/rss-builder/). I can then use my RSS reader to subscribe to these endpoints.

## Limitations

- Built in an afternoon by someone new to Go.
- It makes no attempt to get around bot protection measures.
