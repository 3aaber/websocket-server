# WebSocket Server With Session Key

Using [gin-gonic](https://github.com/gin-gonic/gin) and [gorilla](https://github.com/gorilla) to create websocket server

## Features

- Keep Track of sessions
- Check Session Ids sent from user to validate
- Have Timeout for established sessions
- Detect Duplicate Sessions

## How To Use

`make`

## TODO

- [x] Handler to check session id

- [ ] Rate Limit

- [x] Session Timeout

- [x] Single Session Validation

- [x] Detect Duplicate Sessions

- [ ] Add the client information to session, client IP, true client IP(x-real-ip)


## References

- [gin-gonic](https://github.com/gin-gonic/gin)
- [gorilla](https://github.com/gorilla)
