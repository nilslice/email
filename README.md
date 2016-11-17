### Usage:
`$ go get github.com/nilslice/email`

```go
package main

import (
    "fmt"
    "net/mail"
    "github.com/nilslice/email"
)

func main() {
    to := mail.Address{
        Name: "Foo",
        Address: "foo@example.com",
    }
    from := mail.Address{
        Name: "Bar",
        Address: "bar@example.com",
    }
    msg := email.Message{
        To:      to.Address,
        From:    from.Address,
        Subject: "A simple email",
        Body:    "Plain text email body. HTML not yet supported, but send a PR!",
    }

    err := msg.Send()
    if err != nil {
        fmt.Println(err.Error())
    }
}

```

### Under the hood
`email` looks at a `Message`'s `To` field, splits the string on the @ symbol and
issues an MX lookup to find the mail exchange server(s). Then it iterates over
all the possibilities in combination with commonly used SMTP ports for non-SSL
clients: `25, 2525, & 587`

It stops once it has an active client connected to a mail server and sends the
initial information, the message, and then closes the connection.

Currently, this doesn't support any additional headers or `To` field formatting
(the recipient's email must be the only string `To` takes). Although these would
be fairly strightforward to implement, I don't need them yet.. so feel free to
contribute anything you find useful.

#### Warning
Be cautious of how often you run this locally or in testing, as it's quite
likely your IP will be blocked/blacklisted if it is not already.
