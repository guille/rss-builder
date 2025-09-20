# rss-builder

I use RSS to keep up with sites that interest me. However, not every website provides RSS, or one that is granular enough for my interest. So I built this.

Every night, a scheduled Github Actions workflow runs main.go, which builds a RSS xml file and publishes it to Github Pages. I can then use my RSS reader to subscribe to these endpoints.

## Limitations

- Built in an afternoon by someone new to Go.
- It makes no attempt to get around bot protection measures.
